package helpers

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"go-api/modules/configs"
)

// These tests exercise the real Redis session store (same instance used by
// redis_lock_test.go), so they need the REDIS_SESSION_* env vars from
// global.env plus TOTAL_LOGIN_SESSION, which normally only comes from .env
// via godotenv autoload. We set it explicitly so the test doesn't depend on
// the process cwd matching the repo root.
func setTotalLoginSession(t *testing.T, value string) {
	t.Helper()
	old, hadOld := os.LookupEnv("TOTAL_LOGIN_SESSION")
	if err := os.Setenv("TOTAL_LOGIN_SESSION", value); err != nil {
		t.Fatalf("failed to set TOTAL_LOGIN_SESSION: %s", err.Error())
	}
	t.Cleanup(func() {
		if hadOld {
			_ = os.Setenv("TOTAL_LOGIN_SESSION", old)
		} else {
			_ = os.Unsetenv("TOTAL_LOGIN_SESSION")
		}
	})
}

func newTestSessionData(uuidValue string) SessionData {
	now := time.Now()
	return SessionData{
		Token: SessionToken{
			Uuid:             uuidValue,
			AccessExpiredAt:  now.Add(1 * time.Hour),
			RefreshExpiredAt: now.Add(24 * time.Hour),
		},
		UserID:        "1",
		Authorization: true,
		UserAgent:     "go-test",
		IPAddress:     "127.0.0.1",
	}
}

// cleanupSession removes every redis key a session/family test may have
// touched, regardless of whether assertions failed midway, so repeated runs
// stay idempotent.
func cleanupSession(t *testing.T, username string, uuids ...string) {
	t.Helper()
	t.Cleanup(func() {
		ctx := context.Background()
		redisSession := configs.GetRedis(configs.REDIS_SESSION_PREFIX)
		for _, id := range uuids {
			_, _ = RemoveRedisSession(ctx, username, id)
			keyRedis := getKeyRedis(username, id)
			redisSession.Del(ctx, keyRedis.KeyFamily)
		}
	})
}

func TestSessionLifecycle_CheckTokenAndRemove(t *testing.T) {
	setTotalLoginSession(t, "5")

	ctx := context.Background()
	username := fmt.Sprintf("test-session-%s", uuid.NewString())
	tokenUuid := uuid.NewString()
	cleanupSession(t, username, tokenUuid)

	totalLogin, err := SetupLoginSession(ctx, username, newTestSessionData(tokenUuid))
	if err != nil {
		t.Fatalf("SetupLoginSession failed: %s", err.Error())
	}
	if totalLogin != 1 {
		t.Fatalf("expected total login 1, got %d", totalLogin)
	}

	isExistAccess, isExistRefresh, err := CheckToken(ctx, username, tokenUuid)
	if err != nil {
		t.Fatalf("CheckToken failed: %s", err.Error())
	}
	if !isExistAccess || !isExistRefresh {
		t.Fatalf("expected both access and refresh token to exist, got access=%v refresh=%v", isExistAccess, isExistRefresh)
	}

	totalLogin, err = RemoveRedisSession(ctx, username, tokenUuid)
	if err != nil {
		t.Fatalf("RemoveRedisSession failed: %s", err.Error())
	}
	if totalLogin != 0 {
		t.Fatalf("expected total login 0 after removal, got %d", totalLogin)
	}

	isExistAccess, isExistRefresh, err = CheckToken(ctx, username, tokenUuid)
	if err != nil {
		t.Fatalf("CheckToken after removal failed: %s", err.Error())
	}
	if isExistAccess || isExistRefresh {
		t.Fatalf("expected token to be gone after removal, got access=%v refresh=%v", isExistAccess, isExistRefresh)
	}
}

func TestSetupLoginSession_RejectsOverLimit(t *testing.T) {
	setTotalLoginSession(t, "2")

	ctx := context.Background()
	username := fmt.Sprintf("test-limit-%s", uuid.NewString())
	uuidA := uuid.NewString()
	uuidB := uuid.NewString()
	uuidC := uuid.NewString()
	cleanupSession(t, username, uuidA, uuidB, uuidC)

	if _, err := SetupLoginSession(ctx, username, newTestSessionData(uuidA)); err != nil {
		t.Fatalf("expected 1st session to succeed: %s", err.Error())
	}
	if _, err := SetupLoginSession(ctx, username, newTestSessionData(uuidB)); err != nil {
		t.Fatalf("expected 2nd session to succeed: %s", err.Error())
	}

	if _, err := SetupLoginSession(ctx, username, newTestSessionData(uuidC)); err == nil {
		t.Fatalf("expected 3rd session to be rejected once TOTAL_LOGIN_SESSION=2 is reached")
	}

	isExistAccess, _, err := CheckToken(ctx, username, uuidC)
	if err != nil {
		t.Fatalf("CheckToken failed: %s", err.Error())
	}
	if isExistAccess {
		t.Fatalf("rejected session should not have been persisted")
	}
}

// TestRefreshRotation_ReusedOldTokenIsDetectedAsCollision mirrors the flow
// documented in SetFamily/CheckFamily: refreshing rotates old -> new uuid,
// and replaying the old (already-rotated) refresh token must be flagged as
// reuse, which revokes the token family it produced (the still-valid new
// session gets logged out too, forcing re-authentication).
func TestRefreshRotation_ReusedOldTokenIsDetectedAsCollision(t *testing.T) {
	setTotalLoginSession(t, "5")

	ctx := context.Background()
	username := fmt.Sprintf("test-family-%s", uuid.NewString())
	oldUuid := uuid.NewString()
	newUuid := uuid.NewString()
	cleanupSession(t, username, oldUuid, newUuid)

	// 1. initial login
	if _, err := SetupLoginSession(ctx, username, newTestSessionData(oldUuid)); err != nil {
		t.Fatalf("initial SetupLoginSession failed: %s", err.Error())
	}

	// 2. legitimate refresh: old session removed, new one created, family linked
	if _, err := RemoveRedisSession(ctx, username, oldUuid); err != nil {
		t.Fatalf("RemoveRedisSession (rotation) failed: %s", err.Error())
	}
	if _, err := SetupLoginSession(ctx, username, newTestSessionData(newUuid)); err != nil {
		t.Fatalf("SetupLoginSession (rotation) failed: %s", err.Error())
	}
	if err := SetFamily(ctx, username, oldUuid, newUuid, time.Now().Add(24*time.Hour)); err != nil {
		t.Fatalf("SetFamily failed: %s", err.Error())
	}

	// sanity: rotated session is alive before the replay
	isExistAccess, isExistRefresh, err := CheckToken(ctx, username, newUuid)
	if err != nil {
		t.Fatalf("CheckToken(new) failed: %s", err.Error())
	}
	if !isExistAccess || !isExistRefresh {
		t.Fatalf("expected rotated session to exist before replay check")
	}

	// 3. attacker (or a stale client) replays the old, already-rotated refresh token
	isUsed, err := CheckFamily(ctx, username, oldUuid)
	if err != nil {
		t.Fatalf("CheckFamily failed: %s", err.Error())
	}
	if !isUsed {
		t.Fatalf("expected reused old refresh token to be flagged as collision")
	}

	// 4. the whole family (the rotated session it produced) must be revoked
	isExistAccess, isExistRefresh, err = CheckToken(ctx, username, newUuid)
	if err != nil {
		t.Fatalf("CheckToken(new) after collision failed: %s", err.Error())
	}
	if isExistAccess || isExistRefresh {
		t.Fatalf("expected rotated session to be revoked after reuse was detected, got access=%v refresh=%v", isExistAccess, isExistRefresh)
	}

	// 5. the family marker itself is consumed, so re-checking it is a clean miss
	isUsedAgain, err := CheckFamily(ctx, username, oldUuid)
	if err != nil {
		t.Fatalf("CheckFamily (second check) failed: %s", err.Error())
	}
	if isUsedAgain {
		t.Fatalf("expected family marker to be cleared after being consumed once")
	}
}

// TestGetAllSessions_ExcludesExpiredAndFallsBackToRefreshCopy covers the two
// edge cases around listing devices: a session whose access-token copy has
// already aged out (shorter TTL than the refresh token) must still show up
// by falling back to the refresh copy, while a session that is genuinely
// expired (its refresh entry's score is in the past) must not show up at all.
func TestGetAllSessions_ExcludesExpiredAndFallsBackToRefreshCopy(t *testing.T) {
	setTotalLoginSession(t, "5")

	ctx := context.Background()
	username := fmt.Sprintf("test-list-%s", uuid.NewString())
	liveUuid := uuid.NewString()
	accessLapsedUuid := uuid.NewString()
	expiredUuid := uuid.NewString()
	cleanupSession(t, username, liveUuid, accessLapsedUuid, expiredUuid)

	redisSession := configs.GetRedis(configs.REDIS_SESSION_PREFIX)

	beforeCreate := time.Now()
	for _, id := range []string{liveUuid, accessLapsedUuid, expiredUuid} {
		if _, err := SetupLoginSession(ctx, username, newTestSessionData(id)); err != nil {
			t.Fatalf("SetupLoginSession(%s) failed: %s", id, err.Error())
		}
	}

	// simulate the access token key having already aged out while the refresh
	// token (and its zset entry) are still alive.
	accessLapsedKeys := getKeyRedis(username, accessLapsedUuid)
	if err := redisSession.Del(ctx, accessLapsedKeys.KeyAccess).Err(); err != nil {
		t.Fatalf("failed to simulate lapsed access key: %s", err.Error())
	}

	// simulate a genuinely expired session by pushing its zset score into the past.
	expiredKeys := getKeyRedis(username, expiredUuid)
	pastScore := float64(time.Now().Add(-1 * time.Hour).Unix())
	backdated := redis.Z{Score: pastScore, Member: expiredUuid}
	if err := redisSession.ZAdd(ctx, expiredKeys.KeyTotalAccess, backdated).Err(); err != nil {
		t.Fatalf("failed to backdate access zset score: %s", err.Error())
	}
	if err := redisSession.ZAdd(ctx, expiredKeys.KeyTotalRefresh, backdated).Err(); err != nil {
		t.Fatalf("failed to backdate refresh zset score: %s", err.Error())
	}

	sessions, err := GetAllSessions(ctx, username, liveUuid)
	if err != nil {
		t.Fatalf("GetAllSessions failed: %s", err.Error())
	}

	seen := make(map[string]SessionRedisData)
	for _, s := range sessions {
		seen[s.Uuid] = s
	}

	if _, ok := seen[liveUuid]; !ok {
		t.Fatalf("expected live session %s to be listed", liveUuid)
	}
	if !seen[liveUuid].IsCurrent {
		t.Fatalf("expected live session %s to be marked as current", liveUuid)
	}

	// expiry must reflect the zset score (the authoritative expiry the store already
	// prunes against), not some separately-tracked copy that could drift from it.
	wantAccessExpiry := beforeCreate.Add(1 * time.Hour)
	wantRefreshExpiry := beforeCreate.Add(24 * time.Hour)
	if diff := seen[liveUuid].AccessExpiredAt.Sub(wantAccessExpiry); diff < -5*time.Second || diff > 5*time.Second {
		t.Fatalf("expected access_expired_at near %s, got %s", wantAccessExpiry, seen[liveUuid].AccessExpiredAt)
	}
	if diff := seen[liveUuid].RefreshExpiredAt.Sub(wantRefreshExpiry); diff < -5*time.Second || diff > 5*time.Second {
		t.Fatalf("expected refresh_expired_at near %s, got %s", wantRefreshExpiry, seen[liveUuid].RefreshExpiredAt)
	}
	if _, ok := seen[accessLapsedUuid]; !ok {
		t.Fatalf("expected session %s with a lapsed access key to still be listed via refresh fallback", accessLapsedUuid)
	}
	if _, ok := seen[expiredUuid]; ok {
		t.Fatalf("expired session %s must not be listed", expiredUuid)
	}
}

func TestCheckFamily_UnrotatedTokenIsNotFlagged(t *testing.T) {
	ctx := context.Background()
	username := fmt.Sprintf("test-family-fresh-%s", uuid.NewString())
	freshUuid := uuid.NewString()
	cleanupSession(t, username, freshUuid)

	isUsed, err := CheckFamily(ctx, username, freshUuid)
	if err != nil {
		t.Fatalf("CheckFamily failed: %s", err.Error())
	}
	if isUsed {
		t.Fatalf("a refresh token that was never rotated must not be flagged as reused")
	}
}
