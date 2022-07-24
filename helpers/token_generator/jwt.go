package token_generator

import (
	_ "crypto"
	"os"
	"strconv"
	"time"

	"go-api/modules/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/golang-jwt/jwt"
)

type BaseJwt struct {
	SecretToken   interface{}       `json:"secret_token"`
	RefreshToken  interface{}       `json:"refresh_token"`
	SigningMethod jwt.SigningMethod `json:"signing_method"`
	MapClaims     *JwtMapClaims     `json:"map_claims"`

	ctx *gin.Context
}

type JwtMapClaims struct {
	Uuid        string       `json:"uuid"`
	TokenData   TokenData    `json:"token_data"`
	UserDetails *UserDetails `json:"user_details"`
	HasuraClaim HasuraClaim  `json:"hasura_claim"`
	jwt.StandardClaims
}

type SignedToken struct {
	Uuid             string `json:"uuid"`
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	AccessExpiredAt  int64  `json:"access_expired_at"`
	RefreshExpiredAt int64  `json:"refresh_expired_at"`
}

func NewJwt() *BaseJwt {
	base := BaseJwt{
		SecretToken:   []byte(os.Getenv("JWT_SECRET_TOKEN")),
		RefreshToken:  []byte(os.Getenv("JWT_REFRESH_TOKEN")),
		SigningMethod: jwt.GetSigningMethod(os.Getenv("JWT_HMAC_HASH")),
	}

	return &base
}

func (base *BaseJwt) SetCtx(ctx *gin.Context) *BaseJwt {
	base.ctx = ctx
	return base
}

func (base *BaseJwt) SetSecretKey(secretKey string) *BaseJwt {
	base.SecretToken = []byte(secretKey)
	return base
}

func (base *BaseJwt) SetClaim(customer models.Users) (*BaseJwt, error) {
	timeDate := time.Now()

	userDetails, err := EncryptUserDetails(customer)
	if err != nil {
		return nil, err
	}

	claim := JwtMapClaims{
		TokenData: TokenData{
			Origin:      os.Getenv("GO_API_NAME"),
			UserDetails: userDetails,
			AppData:     base.getAppData(),
		},
		HasuraClaim: HasuraClaim{
			AllowedRoles: []string{"customer", "merchant"},
			DefaultRole:  "customer",
		},
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  timeDate.Unix(),
			ExpiresAt: timeDate.AddDate(0, 0, 1).Unix(), // default 1 day
			Issuer:    os.Getenv("APP_NAME"),
		},
	}

	base.MapClaims = &claim
	return base, nil
}

func (base *BaseJwt) getAppData() *AppData {
	var appData AppData
	if base.ctx != nil {
		ctx := base.ctx
		appData.UserAgent = ctx.Request.UserAgent()
		appData.IPList = append(appData.IPList, ctx.Request.RemoteAddr)
	}

	appData.AppName = os.Getenv("GO_API_NAME")
	return &appData
}

func (base *BaseJwt) SetClaimApp(appData AppData) *BaseJwt {
	claim := JwtMapClaims{
		TokenData: TokenData{
			Authorized: true,
			AppData:    &appData,
		},
		StandardClaims: jwt.StandardClaims{},
	}

	base.MapClaims = &claim
	return base
}

func (base *BaseJwt) SignClaims() (signedToken *SignedToken, err error) {
	accessUuid := uuid.New().String()
	accessToken, accessExpired, err := base.signClaims("access_token", accessUuid)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshExpired, err := base.signClaims("refresh_token", accessUuid)
	if err != nil {
		return nil, err
	}

	token := &SignedToken{
		Uuid:             accessUuid,
		AccessToken:      accessToken,
		AccessExpiredAt:  accessExpired,
		RefreshToken:     refreshToken,
		RefreshExpiredAt: refreshExpired,
	}

	return token, nil
}

/**
All JWT uuid, must be signed with ACCESS UUID from JWT UUID
so that we can made only one refresh token for many device,
but still have one access token
*/
func (base *BaseJwt) signClaims(typeClaims string, accessUuid string) (signedToken string, expiredAt int64, err error) {
	expiredDate := base.GetExpiredDate(typeClaims)

	mapClaims := base.MapClaims
	mapClaims.Uuid = accessUuid
	mapClaims.StandardClaims.ExpiresAt = expiredDate

	token := jwt.NewWithClaims(base.SigningMethod, mapClaims)
	signedToken, err = token.SignedString(base.SecretToken)

	return signedToken, expiredDate, err
}

func (base *BaseJwt) GetExpiredDate(typeClaims string) int64 {
	timeDate := time.Now()

	var tokenExpired int64
	var tokenExpiredType string
	switch typeClaims {
	case "access_token":
		tokenExpired, _ = strconv.ParseInt(os.Getenv("JWT_ACCESS_TOKEN_EXPIRED"), 10, 64)
		tokenExpiredType = os.Getenv("JWT_ACCESS_TOKEN_EXPIRED_TYPE")
	case "refresh_token":
		tokenExpired, _ = strconv.ParseInt(os.Getenv("JWT_REFRESH_TOKEN_EXPIRED"), 10, 64)
		tokenExpiredType = os.Getenv("JWT_REFRESH_TOKEN_EXPIRED_TYPE")
	}

	var expiredTime int64
	switch tokenExpiredType {
	case "days":
		expiredTime = timeDate.AddDate(0, 0, int(tokenExpired)).Unix()
	case "months":
		expiredTime = timeDate.AddDate(0, int(tokenExpired), 0).Unix()
	case "years":
		expiredTime = timeDate.AddDate(int(tokenExpired), 0, 0).Unix()
	default:
		panic("token expired date type is not supported, please pick (days, months, years)")
	}

	return expiredTime
}

func (base *BaseJwt) ParseToken(jwtToken string) (*JwtMapClaims, error) {
	token, err := jwt.ParseWithClaims(jwtToken, &JwtMapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return base.SecretToken, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JwtMapClaims); ok && token.Valid {
		if claims.TokenData.UserDetails != "" {
			userDetails, err := DecryptUserDetails(claims.TokenData.UserDetails)
			if err != nil {
				return nil, err
			}

			claims.UserDetails = &userDetails
		}

		return claims, nil
	}

	return nil, err
}
