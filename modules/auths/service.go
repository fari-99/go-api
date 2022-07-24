package auths

import (
	"github.com/gin-gonic/gin"

	"go-api/helpers"
	"go-api/helpers/token_generator"
	"go-api/modules/models"
)

type Service interface {
	AuthenticateUser(ctx *gin.Context, input RequestAuthUser) (totalLogin int64, token *token_generator.SignedToken, isExists bool, err error)
	RefreshAuth(ctx *gin.Context) (token *token_generator.SignedToken, isExists bool, err error)
	SignOutUser(ctx *gin.Context) (totalLogin int64, isExists bool, err error)
	DeleteAllSession(ctx *gin.Context) (isExists bool, err error)
	AllSessions(ctx *gin.Context) (allDevices []models.Users, err error)
	DeleteSession(ctx *gin.Context, uuid string) (totalLogin int64, isExists bool, err error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return service{repo: repo}
}

func (s service) RefreshAuth(ctx *gin.Context) (signedToken *token_generator.SignedToken, isExists bool, err error) {
	// check if refresh token still exists using exists (exists uuid:refresh_token)
	// - if refresh token not exists, send new auth (isExists: false)
	// - if refresh token exists, update token expired_at
	// delete current access & refresh token using del (del uuid:access_token) (del uuid:refresh_token)
	// create new access & refresh token

	uuidSession, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUserRefresh(uuidSession.(string))

	_, isExistRefresh, err := helpers.CheckToken(currentUser.Username, uuidSession.(string))
	if err != nil || !isExistRefresh {
		return nil, isExistRefresh, err
	}

	_, err = helpers.RemoveRedisSession(currentUser.Username, uuidSession.(string))
	if err != nil {
		return nil, false, err
	}

	token, err := s.generateToken(ctx, *currentUser)
	if err != nil {
		return nil, false, err
	}

	_, err = s.setRedisSession(token, currentUser)
	if err != nil {
		return nil, false, err
	}

	return token, true, nil
}

func (s service) DeleteSession(ctx *gin.Context, uuid string) (totalLogin int64, isExists bool, err error) {
	uuidSession, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(uuidSession.(string))

	totalLogin, err = helpers.RemoveRedisSession(currentUser.Username, uuid)
	if err != nil {
		return 0, false, err
	}

	return totalLogin, true, nil
}

func (s service) AllSessions(ctx *gin.Context) (allDevices []models.Users, err error) {
	uuid, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(uuid.(string))

	return helpers.GetAllSessions(currentUser.Username)
}

func (s service) SignOutUser(ctx *gin.Context) (int64, bool, error) {
	uuid, isExist := ctx.Get("uuid")
	if !isExist {
		return 0, false, nil
	}

	currentUser, err := helpers.GetCurrentUser(uuid.(string))
	if err != nil {
		return 0, true, err
	}

	totalLogin, err := helpers.RemoveRedisSession(currentUser.Username, uuid.(string))
	if err != nil {
		return 0, true, err
	}

	return totalLogin, false, nil
}

func (s service) DeleteAllSession(ctx *gin.Context) (bool, error) {
	uuid, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(uuid.(string))

	err := helpers.DeleteAllSession(currentUser.Username, uuid.(string))
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s service) AuthenticateUser(ctx *gin.Context, input RequestAuthUser) (int64, *token_generator.SignedToken, bool, error) {
	if err := input.Validate(); err != nil {
		return 0, nil, false, err
	}

	userModel, notFound, err := s.repo.AuthenticatePassword(input)
	if err != nil {
		return 0, nil, false, err
	} else if notFound {
		return 0, nil, notFound, nil
	}

	token, err := s.generateToken(ctx, *userModel)
	if err != nil {
		return 0, nil, false, err
	}

	totalLogin, err := s.setRedisSession(token, userModel)
	if err != nil {
		return 0, nil, false, err
	}

	return totalLogin, token, false, nil
}

func (s service) generateToken(ctx *gin.Context, userModel models.Users) (signedToken *token_generator.SignedToken, err error) {
	// generate JWT token
	tokenHelper := token_generator.NewJwt().SetCtx(ctx)
	tokenHelper, err = tokenHelper.SetClaim(userModel)
	if err != nil {
		return nil, err
	}

	token, err := tokenHelper.SignClaims()
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s service) setRedisSession(token *token_generator.SignedToken, userModel *models.Users) (totalLogin int64, err error) {
	dataSession := helpers.SessionData{
		Token: helpers.SessionToken{
			AccessExpiredAt:  token.AccessExpiredAt,
			Uuid:             token.Uuid,
			RefreshExpiredAt: token.RefreshExpiredAt,
		},

		UserID:        userModel.ID,
		UserDetails:   userModel,
		Authorization: true,
	}

	return helpers.SetupLoginSession(userModel.Username, dataSession)
}
