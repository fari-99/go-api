package helpers

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

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
