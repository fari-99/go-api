package auths

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"

	"go-api/helpers"
	"go-api/modules/middleware"
	"go-api/modules/models"
)

type fakeRepo struct{}

func (fakeRepo) AuthenticatePassword(_ *gin.Context, input RequestAuthUser) (*models.Users, bool, error) {
	return &models.Users{
		Base:     models.Base{ID: 7},
		Username: "remember-me-user",
		Email:    input.Email,
		Roles:    "admin",
	}, false, nil
}

func (fakeRepo) GetUserDetails(_ *gin.Context, id uint64) (models.Users, error) {
	return models.Users{
		Base:     models.Base{ID: models.IDType(id)},
		Username: "remember-me-user",
		Email:    "a@b.co",
		Roles:    "admin",
	}, nil
}

// configs.GetRedis caches its client per process, so every test shares one
// miniredis (started in TestMain) and flushes it instead of starting its own.
var sharedRedis *miniredis.Miniredis

func TestMain(m *testing.M) {
	sharedRedis, _ = miniredis.Run()

	_ = os.Setenv("REDIS_SESSION_HOST", sharedRedis.Host())
	_ = os.Setenv("REDIS_SESSION_PORT", sharedRedis.Port())
	_ = os.Setenv("REDIS_SESSION_DB", "0")
	_ = os.Setenv("TOTAL_LOGIN_SESSION", "5")
	_ = os.Setenv("JWT_HMAC_HASH", "HS512")
	_ = os.Setenv("JWT_SECRET_TOKEN", "0123456789abcdef0123456789abcdef")
	_ = os.Setenv("JWT_REFRESH_TOKEN", "fedcba9876543210fedcba9876543210")
	_ = os.Setenv("JWT_ACCESS_TOKEN_EXPIRED", "1")
	_ = os.Setenv("JWT_ACCESS_TOKEN_EXPIRED_TYPE", "days")
	_ = os.Setenv("JWT_REFRESH_TOKEN_EXPIRED", "7")
	_ = os.Setenv("JWT_REFRESH_TOKEN_EXPIRED_TYPE", "days")

	code := m.Run()
	sharedRedis.Close()
	os.Exit(code)
}

func setupRememberMeRouter(t *testing.T) (*gin.Engine, *miniredis.Miniredis) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	sharedRedis.FlushAll()
	mr := sharedRedis

	ctrl := controller{service: NewService(fakeRepo{})}
	router := gin.New()
	router.POST("/users/auth", ctrl.AuthenticateAction)
	router.POST("/users/sessions/refresh", middleware.RefreshAuthMiddleware(middleware.BaseMiddleware{}), ctrl.RefreshSession)
	return router, mr
}

type loginResult struct {
	AccessToken  string
	RefreshToken *http.Cookie
	Body         string
	Uuid         string
}

func login(t *testing.T, router *gin.Engine, rememberMe bool) loginResult {
	t.Helper()

	body := `{"email":"a@b.co","password":"secret","remember_me":false}`
	if rememberMe {
		body = strings.Replace(body, `"remember_me":false`, `"remember_me":true`, 1)
	}

	req := httptest.NewRequest(http.MethodPost, "/users/auth", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login failed, status=%d body=%s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil || resp.Data.AccessToken == "" {
		t.Fatalf("could not read access_token from %s (err=%v)", rec.Body.String(), err)
	}

	result := loginResult{AccessToken: resp.Data.AccessToken, Body: rec.Body.String()}
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == "refresh_token" {
			result.RefreshToken = cookie
		}
	}
	return result
}

func refresh(router *gin.Engine, cookie *http.Cookie, authorization string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/users/sessions/refresh", nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestLogin_WithoutRememberMe_HasNoRefreshToken(t *testing.T) {
	router, mr := setupRememberMeRouter(t)

	result := login(t, router, false)

	if result.RefreshToken != nil {
		t.Fatalf("expected no refresh_token cookie, got %+v", result.RefreshToken)
	}
	if strings.Contains(result.Body, "refresh_token") {
		t.Fatalf("response body must not contain a refresh token: %s", result.Body)
	}

	// redis: access session exists, nothing refresh related exists
	var accessKeys, refreshKeys int
	for _, key := range mr.Keys() {
		switch {
		case strings.HasSuffix(key, ":access_token"):
			accessKeys++
		case strings.HasSuffix(key, ":refresh_token"):
			refreshKeys++
		}
	}
	// uuid:access_token + username:access_token (zset)
	if accessKeys != 2 {
		t.Fatalf("expected 2 access keys, got %d (keys=%v)", accessKeys, mr.Keys())
	}
	if refreshKeys != 0 {
		t.Fatalf("expected no refresh keys in redis, got %d (keys=%v)", refreshKeys, mr.Keys())
	}

	// the session must still be listed as a device
	sessions, err := helpers.GetAllSessions(context.Background(), "remember-me-user", "")
	if err != nil {
		t.Fatalf("GetAllSessions failed: %s", err.Error())
	}
	if len(sessions) != 1 {
		t.Fatalf("expected the no-refresh session to be listed, got %d", len(sessions))
	}
}

func TestRefresh_WithoutRefreshToken_IsUnauthorizedNotPanic(t *testing.T) {
	router, _ := setupRememberMeRouter(t)
	result := login(t, router, false)

	// no cookie, no header
	if rec := refresh(router, nil, ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("no token: expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}

	// access token presented as if it were a refresh token
	if rec := refresh(router, nil, "Bearer "+result.AccessToken); rec.Code != http.StatusUnauthorized {
		t.Fatalf("access token as refresh: expected 401, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRefreshAuth_SessionMissingRefreshKey_ReturnsNotExists(t *testing.T) {
	router, _ := setupRememberMeRouter(t)
	_ = login(t, router, false)

	// service level: uuid whose redis session has no refresh key (family empty too)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodPost, "/", nil)
	ctx.Set("uuid", "uuid-without-refresh-key")

	authData, isExists, err := NewService(fakeRepo{}).RefreshAuth(ctx)
	if err != nil || isExists || authData != nil {
		t.Fatalf("expected (nil, false, nil), got (%v, %v, %v)", authData, isExists, err)
	}

	// family check on a uuid that never refreshed must be a clean miss
	isUsed, err := helpers.CheckFamily(context.Background(), "remember-me-user", "uuid-without-refresh-key")
	if err != nil || isUsed {
		t.Fatalf("expected clean family miss, got used=%v err=%v", isUsed, err)
	}
}

func TestLogin_WithRememberMe_RefreshStillWorks(t *testing.T) {
	router, mr := setupRememberMeRouter(t)

	result := login(t, router, true)
	if result.RefreshToken == nil || result.RefreshToken.Value == "" {
		t.Fatalf("expected refresh_token cookie with remember me")
	}
	if !result.RefreshToken.HttpOnly {
		t.Fatalf("refresh_token cookie must stay HttpOnly")
	}
	if !mr.Exists(mustUuidKey(t, mr, ":refresh_token")) {
		t.Fatalf("expected refresh key in redis, keys=%v", mr.Keys())
	}

	rec := refresh(router, result.RefreshToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("refresh with remember me: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var renewed *http.Cookie
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == "refresh_token" {
			renewed = cookie
		}
	}
	if renewed == nil || renewed.Value == result.RefreshToken.Value {
		t.Fatalf("expected a rotated refresh_token cookie after refresh")
	}
}

// mustUuidKey finds the uuid-scoped redis key (not the username zset) with the given suffix.
func mustUuidKey(t *testing.T, mr *miniredis.Miniredis, suffix string) string {
	t.Helper()
	for _, key := range mr.Keys() {
		if strings.HasSuffix(key, suffix) && !strings.HasPrefix(key, "remember-me-user:") {
			return key
		}
	}
	t.Fatalf("no key with suffix %s in %v", suffix, mr.Keys())
	return ""
}

// expireAccessToken simulates the access token aging out: its redis key is
// gone and its zset score (the expiry) is in the past.
func expireAccessToken(t *testing.T, mr *miniredis.Miniredis) string {
	t.Helper()
	uuidKey := mustUuidKey(t, mr, ":access_token")
	uuid := strings.TrimSuffix(uuidKey, ":access_token")
	mr.Del(uuidKey)
	if _, err := mr.ZAdd("remember-me-user:access_token", 1, uuid); err != nil {
		t.Fatalf("zadd failed: %s", err.Error())
	}
	return uuid
}

func TestRefresh_AfterAccessTokenExpiredInRedis_StillWorks(t *testing.T) {
	router, mr := setupRememberMeRouter(t)
	result := login(t, router, true)
	expireAccessToken(t, mr)

	rec := refresh(router, result.RefreshToken, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected refresh to work without the access key in redis, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPruning_ExpiredAccessKeepsRefreshableSession(t *testing.T) {
	router, mr := setupRememberMeRouter(t)
	result := login(t, router, true)
	uuid := expireAccessToken(t, mr)

	// any listing / login prunes expired tokens
	sessions, err := helpers.GetAllSessions(context.Background(), "remember-me-user", "")
	if err != nil {
		t.Fatalf("GetAllSessions failed: %s", err.Error())
	}
	if len(sessions) != 1 || sessions[0].Uuid != uuid {
		t.Fatalf("expected the refreshable session to survive pruning, got %+v", sessions)
	}
	if !mr.Exists(uuid + ":refresh_token") {
		t.Fatalf("refresh key must survive access token expiry, keys=%v", mr.Keys())
	}

	if rec := refresh(router, result.RefreshToken, ""); rec.Code != http.StatusOK {
		t.Fatalf("refresh after pruning: expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestPruning_ExpiredAccessWithoutRefreshRemovesSession(t *testing.T) {
	router, mr := setupRememberMeRouter(t)
	_ = login(t, router, false)
	expireAccessToken(t, mr)

	sessions, err := helpers.GetAllSessions(context.Background(), "remember-me-user", "")
	if err != nil {
		t.Fatalf("GetAllSessions failed: %s", err.Error())
	}
	if len(sessions) != 0 {
		t.Fatalf("expected the expired no-refresh session to be gone, got %+v", sessions)
	}
	if len(mr.Keys()) != 0 {
		t.Fatalf("expected redis to be clean, keys=%v", mr.Keys())
	}
}
