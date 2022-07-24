package users

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go-api/helpers"
)

type controller struct {
	service Service
}

func (c controller) CreateAction(ctx *gin.Context) {
	var input RequestCreateUser
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	_, err = c.service.CreateUser(ctx, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusOK, gin.H{
			"error":         err.Error(),
			"error_message": "failed to create user, please try again",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "User successfully created")
	return
}

func (c controller) UserProfileAction(ctx *gin.Context) {
	uuidSession, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUserRefresh(uuidSession.(string))

	userProfile, err := c.service.UserProfile(ctx, currentUser.ID)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusOK, gin.H{
			"error":         err.Error(),
			"error_message": "failed to get user profile, please try again",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, userProfile)
	return
}
