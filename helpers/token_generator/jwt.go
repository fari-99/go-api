package token_generator

import (
	_ "crypto"
	"go-api/models"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type BaseJwt struct {
	SecretToken   interface{}       `json:"secret_token"`
	SigningMethod jwt.SigningMethod `json:"signing_method"`
	MapClaims     JwtMapClaims      `json:"map_claims"`
}

type JwtMapClaims struct {
	TokenData TokenData `json:"token_data"`
	jwt.StandardClaims
}

func NewJwt() *BaseJwt {
	base := BaseJwt{
		SecretToken:   []byte(os.Getenv("JWT_SECRET_TOKEN")),
		SigningMethod: jwt.GetSigningMethod(os.Getenv("JWT_HMAC_HASH")),
		MapClaims:     JwtMapClaims{},
	}

	return &base
}

func (base *BaseJwt) SetSecretKey(secretKey string) *BaseJwt {
	base.SecretToken = []byte(secretKey)
	return base
}

func (base *BaseJwt) SetClaim(customer models.Customers) *BaseJwt {
	timeDate := time.Now()
	expiredDate := timeDate.AddDate(0, 0, 1).Unix()

	claim := JwtMapClaims{
		TokenData: TokenData{
			Origin: os.Getenv("APP_NAME"),
			UserDetails: &UserDetails{
				ID:       customer.ID,
				Email:    customer.Email,
				Username: customer.Username,
			},
		},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiredDate,
		},
	}

	base.MapClaims = claim
	return base
}

func (base *BaseJwt) SetClaimApp(appData AppData) *BaseJwt {
	timeDate := time.Now()
	expiredDate := timeDate.AddDate(1, 2, 3).Unix() // expired after 1 year, 2 month and 3 days

	claim := JwtMapClaims{
		TokenData: TokenData{
			Authorized: true,
			AppData:    &appData,
		},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiredDate,
		},
	}

	base.MapClaims = claim
	return base
}

func (base *BaseJwt) SignClaim() (signedToken string, err error) {
	token := jwt.NewWithClaims(base.SigningMethod, base.MapClaims)
	signedToken, err = token.SignedString(base.SecretToken)
	return
}

func (base *BaseJwt) ParseToken(jwtToken string) (*JwtMapClaims, error) {
	token, err := jwt.ParseWithClaims(jwtToken, &JwtMapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return base.SecretToken, nil
	})

	if err != nil {
		return &JwtMapClaims{}, err
	}

	if claims, ok := token.Claims.(*JwtMapClaims); ok && token.Valid {
		return claims, nil
	}

	return &JwtMapClaims{}, err
}
