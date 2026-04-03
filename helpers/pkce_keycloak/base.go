package pkce_keycloak

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	gohelper "github.com/fari-99/go-helper"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"

	"go-api/pkg/keycloaks"
	keycloaksModel "go-api/pkg/keycloaks/models"
)

const KeyRedisRedirectAuth = "[redirect-auth][%s][%s][%d]"

type PkceKeycloakHelper struct {
	redisClient         *redis.Client
	keycloakConfigs     KeycloakConfigs
	ctx                 context.Context
	keycloakFrontEndUrl string
}

type KeycloakConfigs struct {
	Realm        string
	ClientID     string
	ClientSecret string
}

type RedisRedirectAuthData struct {
	VerifyKey   string          `json:"verify_key"`
	Input       json.RawMessage `json:"input"`
	RedirectUrl string          `json:"redirect_url"`
}

func NewPkceKeycloakHelper(ctx context.Context, redisClient *redis.Client) (*PkceKeycloakHelper, error) {
	keycloakConfigs := KeycloakConfigs{
		Realm:        os.Getenv("KEYCLOAK_REALM"),
		ClientID:     os.Getenv("KEYCLOAK_PKCE_CLIENT_ID"),
		ClientSecret: os.Getenv("KEYCLOAK_PKCE_CLIENT_SECRET"),
	}

	if keycloakConfigs.Realm == "" {
		return nil, errors.New("keycloak config host or keycloak config realm is empty")
	}

	if redisClient == nil {
		return nil, errors.New("redis client is nil")
	}

	return &PkceKeycloakHelper{
		redisClient:         redisClient,
		keycloakConfigs:     keycloakConfigs,
		ctx:                 ctx,
		keycloakFrontEndUrl: os.Getenv("KEYCLOAK_FRONTEND_URL"),
	}, nil
}

func (p *PkceKeycloakHelper) SetKeycloakConfigs(keycloakConfigs KeycloakConfigs) (*PkceKeycloakHelper, error) {
	if keycloakConfigs.Realm == "" {
		return nil, errors.New("keycloak config host or keycloak config realm is empty")
	}

	p.keycloakConfigs = keycloakConfigs
	return p, nil
}

type GetRedirectUrl struct {
	UserID      uint64
	Action      string
	Input       interface{}
	RedirectUrl string
}

func (p *PkceKeycloakHelper) GenerateRedirectUrl(input GetRedirectUrl) (string, error) {
	redisClient := p.redisClient
	ctx := p.ctx
	keycloakConfigs := p.keycloakConfigs

	authKeycloak := p.keycloakFrontEndUrl + "/realms/" + keycloakConfigs.Realm + "/protocol/openid-connect/auth"
	clientID := keycloakConfigs.ClientID
	responseType := "code"
	scope := "openid"
	verifier := oauth2.GenerateVerifier()
	codeChallenge := oauth2.S256ChallengeFromVerifier(verifier)
	codeChallengeMethod := "S256"
	state := gohelper.GenerateRandString(16, "alphanum")

	redisData := RedisRedirectAuthData{
		VerifyKey:   verifier,
		Input:       setInputAsJson(input.Input),
		RedirectUrl: input.RedirectUrl,
	}

	redisDataMarshal, _ := json.Marshal(redisData)

	redisKey := fmt.Sprintf(KeyRedisRedirectAuth, state, input.Action, input.UserID)
	err := redisClient.Set(ctx, redisKey, string(redisDataMarshal), time.Duration(5)*time.Minute).Err()
	if err != nil {
		return "", err
	}

	urlData, err := url.Parse(authKeycloak)
	if err != nil {
		return "", err
	}

	query := urlData.Query()
	query.Set("client_id", clientID)
	query.Set("redirect_uri", input.RedirectUrl)
	query.Set("response_type", responseType)
	query.Set("scope", scope)
	query.Set("state", state)
	query.Set("code_challenge", codeChallenge)
	query.Set("code_challenge_method", codeChallengeMethod)
	query.Set("prompt", "login")

	urlData.RawQuery = query.Encode()

	return urlData.String(), nil
}

func (p *PkceKeycloakHelper) GetAccessToken(action string, userID uint64, code, state, iss string) (string, error) {
	redirectData, err := p.getRedisData(action, userID, iss, state)
	if err != nil {
		return "", err
	}

	keycloakConfigs := p.keycloakConfigs

	keycloakAuth := keycloaksModel.AuthenticateKeycloak{
		ClientID:     keycloakConfigs.ClientID,
		ClientSecret: keycloakConfigs.ClientSecret,
		GrantType:    "authorization_code",
		Scope:        "openid",
		Code:         code,
		CodeVerifier: redirectData.VerifyKey,
		RedirectUri:  redirectData.RedirectUrl,
	}

	accessToken, err := getAccessTokenByAuthCode(keycloakAuth)
	if err != nil {
		return "", err
	}

	// no validation because get fresh form keycloak
	parsedToken, _, err := new(jwt.Parser).ParseUnverified(accessToken, jwt.MapClaims{})
	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	userIDCheck := fmt.Sprintf("%d", userID)
	kcUserID, err := parsedToken.Claims.GetSubject()
	if err != nil || subtle.ConstantTimeCompare([]byte(kcUserID), []byte(userIDCheck)) == 0 {
		return "", errors.New("identity mismatch")
	}

	return accessToken, err
}

func (p *PkceKeycloakHelper) getRedisData(action string, userID uint64, iss, state string) (data *RedisRedirectAuthData, err error) {
	redisClient := p.redisClient
	ctx := p.ctx

	expectedIssuer := p.keycloakFrontEndUrl + "/realms/" + p.keycloakConfigs.Realm
	if subtle.ConstantTimeCompare([]byte(iss), []byte(expectedIssuer)) == 0 {
		return nil, fmt.Errorf("issuer data is not valid")
	}

	redisKey := fmt.Sprintf(KeyRedisRedirectAuth, state, action, userID)
	redisData, err := redisClient.GetDel(ctx, redisKey).Result()
	if err != nil {
		return nil, err
	}

	var redirectData RedisRedirectAuthData
	err = json.Unmarshal([]byte(redisData), &redirectData)
	if err != nil {
		return nil, err
	}

	return &redirectData, nil
}

func getAccessTokenByAuthCode(keycloakAuth keycloaksModel.AuthenticateKeycloak) (accessToken string, err error) {
	if err = keycloakAuth.ValidateGrantAuthoritizationCode(); err != nil {
		return "", err
	}

	openIDHelper := keycloaks.NewOpenIDHelper()
	respAccessToken, err := openIDHelper.GetAccessToken(keycloakAuth)
	if err != nil {
		return "", err
	}

	return respAccessToken.AccessToken, err
}

func setInputAsJson(input interface{}) json.RawMessage {
	var rawMessage json.RawMessage
	bytes, _ := json.Marshal(input)
	_ = json.Unmarshal(bytes, &rawMessage)
	return rawMessage
}
