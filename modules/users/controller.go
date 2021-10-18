package users

import (
	"github.com/gin-gonic/gin"
	"go-api/helpers"
	"net/http"
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
