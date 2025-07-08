package users

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"go-api/constant"
	"go-api/helpers"
	"go-api/helpers/notifications"
	"go-api/modules/configs/rabbitmq"
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
	currentUser, _ := helpers.GetCurrentUser(ctx, uuidSession.(string))

	userProfile, err := c.service.UserProfile(ctx, string(currentUser.ID))
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

func (c controller) ChangePasswordAction(ctx *gin.Context) {
	var input RequestChangePassword
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	exists, err := c.service.ChangePassword(ctx, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"error":         err.Error(),
			"error_message": "error changing your password",
		})
		return
	} else if !exists {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error_message": "user not found",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "successfully changed your password")
	return
}

func (c controller) ForgotPasswordAction(ctx *gin.Context) {
	var input ForgotPasswordRequest
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	userCode, notFound, err := c.service.ForgotPassword(ctx, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"error":         err.Error(),
			"error_message": "error handling forgot password action",
		})
		return
	} else if notFound {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error_message": "user not found",
		})
		return
	}

	// send emails
	emails := notifications.Email{
		Subject: "Your forgotten password",
		Body:    fmt.Sprintf("your code for reset password := %s and expired at := %s", userCode.Code, userCode.ExpiredAt.Format("2006-01-02 15:04:05")),
		From:    "no-reply@fadhlan.com",
		To:      []string{input.Email},
	}

	err = notifications.SendEmail(emails)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, "failed to send email to you, please try again or contact administrator")
		return
	}

	// send events
	queueData := map[string]interface{}{
		"action":    "forgot-password",
		"input":     input,
		"user_code": userCode,
	}

	queueDataMarshal, _ := json.Marshal(queueData)

	queueSetup := rabbitmq.NewBaseQueue("", constant.QueueUserAction)
	defer queueSetup.Close() // close connection after it's done

	queueSetup.SetupQueue(nil, nil)                  // use default queue config
	queueSetup.AddPublisher(nil, nil)                // use default publisher config
	_ = queueSetup.Publish(string(queueDataMarshal)) // publish queue

	helpers.NewResponse(ctx, http.StatusOK, "reset password token and link successfully send to your email")
	return
}

func (c controller) ForgotUsernameAction(ctx *gin.Context) {
	var input ForgotUsernameRequest
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	exists, err := c.service.ForgotUsername(ctx, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"error":         err.Error(),
			"error_message": "error handling forgot password action",
		})
		return
	} else if !exists {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error_message": "user not found",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "your username successfully send to your email")
	return
}

func (c controller) ResetPasswordAction(ctx *gin.Context) {
	var input ResetPasswordRequest
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	err = c.service.ResetPassword(ctx, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"error":         err.Error(),
			"error_message": "error handling reset password action",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "your password successfully changed")
	return
}
