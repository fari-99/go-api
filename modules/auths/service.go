package auths

import (
	"github.com/gin-gonic/gin"
	"go-api/helpers"
	"go-api/helpers/token_generator"
	"go-api/modules/models"
)

type Service interface {
	AuthenticateUser(ctx *gin.Context, input RequestAuthUser) (*token_generator.SignedToken, bool, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return service{repo: repo}
}

func (s service) AuthenticateUser(ctx *gin.Context, input RequestAuthUser) (*token_generator.SignedToken, bool, error) {
	if err := input.Validate(); err != nil {
		return nil, false, err
	}

	userModel, notFound, err := s.repo.AuthenticatePassword(input)
	if err != nil {
		return nil, false, err
	} else if notFound {
		return nil, notFound, nil
	}

	// generate JWT token
	tokenHelper := token_generator.NewJwt().SetCtx(ctx)
	tokenHelper, err = tokenHelper.SetClaim(*userModel)
	if err != nil {
		return nil, false, err
	}

	token, err := tokenHelper.SignClaims()
	if err != nil {
		return nil, false, err
	}

	err = s.setRedisSession(token, userModel)
	if err != nil {
		return nil, false, err
	}

	return token, false, nil
}

func (s service) setRedisSession(token *token_generator.SignedToken, userModel *models.Users) error {
	dataSession := helpers.SessionData{
		Token: helpers.SessionToken{
			AccessExpiredAt:  token.AccessExpiredAt,
			AccessUuid:       token.AccessUuid,
			RefreshExpiredAt: token.RefreshExpiredAt,
			RefreshUuid:      token.RefreshUuid,
		},

		UserID:        userModel.ID,
		UserDetails:   userModel,
		Authorization: true,
	}

	return helpers.SetRedisSession(dataSession)
}
