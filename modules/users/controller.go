package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/fari-99/go-helper/rabbitmq"
	"github.com/gin-gonic/gin"

	"go-api/constant"
	"go-api/helpers"
	"go-api/helpers/notifications"
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
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
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

	userProfile, err := c.service.UserProfile(ctx, currentUser.ID.Uint64())
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

func (c controller) GetListAction(ctx *gin.Context) {
	pageQuery := ctx.DefaultQuery("page", "1")
	page, _ := strconv.ParseInt(pageQuery, 10, 64)

	limitQuery := ctx.DefaultQuery("limit", "10")
	limit, _ := strconv.ParseInt(limitQuery, 10, 64)

	filter := RequestListUsers{
		Page:    int(page),
		Limit:   int(limit),
		OrderBy: ctx.DefaultQuery("order_by", ""),
		Search:  ctx.DefaultQuery("search", ""),
	}

	items, paginatorData, err := c.service.GetList(ctx, filter)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	result := map[string]interface{}{
		"paginator": paginatorData,
		"items":     items,
	}

	helpers.NewResponse(ctx, http.StatusOK, result)
	return
}

func (c controller) GetDetailAction(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, "invalid user id")
		return
	}

	detail, notFound, err := c.service.UserDetails(ctx, id)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	} else if notFound {
		helpers.NewResponse(ctx, http.StatusNotFound, "user not found")
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, detail)
	return
}

func (c controller) GetUserRolesAction(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, "invalid user id")
		return
	}

	roleIDs, err := c.service.GetUserRoles(ctx, id)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, gin.H{"role_ids": roleIDs})
	return
}

func (c controller) UpdateUserRolesAction(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, "invalid user id")
		return
	}

	var input RequestUserRoles
	if err = ctx.BindJSON(&input); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err = c.service.UpdateUserRoles(ctx, id, input.RoleIDs); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to update user roles, please try again",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "User roles successfully updated")
	return
}

func (c controller) UpdateAction(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, "invalid user id")
		return
	}

	var input RequestUpdateUser
	if err = ctx.BindJSON(&input); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	result, err := c.service.UpdateUserAction(ctx, id, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to update user, please try again",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, result)
	return
}

func (c controller) DeleteAction(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, "invalid user id")
		return
	}

	err = c.service.DeleteUser(ctx, id)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to delete user, please try again",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "User successfully deleted")
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
		var validationErrs validation.Errors
		switch {
		case errors.Is(err, ErrInvalidCurrentPassword):
			helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
				"error":         err.Error(),
				"error_message": "invalid current password",
			})
		case errors.Is(err, ErrWeakNewPassword):
			helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
				"error":         err.Error(),
				"error_message": "new password is not strong enough, please choose a stronger one",
			})
		case errors.As(err, &validationErrs):
			helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
				"error":         err.Error(),
				"error_message": err.Error(),
			})
		default:
			helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
				"error":         err.Error(),
				"error_message": "error changing your password",
			})
		}
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
