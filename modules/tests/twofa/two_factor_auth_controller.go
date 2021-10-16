package twofa

import (
	"bytes"
	"fmt"
	"github.com/dgryski/dgoogauth"
	"github.com/gin-gonic/gin"
	"go-api/helpers"
	"go-api/modules/configs"
	"io"
	"net/http"
	"os"
	"rsc.io/qr"
)

type TwoFactorAuthController struct {
	*configs.DI
}

const secret = "2MXGP5X3FVUEK6W4UB2PPODSP2GKYWUT"

func (controller *TwoFactorAuthController) NewAuth(ctx *gin.Context) {
	account := "test@email.com"
	issuer := os.Getenv("APP_NAME")

	authLink := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s", issuer, account, secret, issuer)
	code, err := qr.Encode(authLink, qr.H)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	img := code.PNG()
	buf := bytes.NewReader(img)

	responseWriter := ctx.Writer
	responseWriter.Header().Set("Content-Type", "image/png")
	responseWriter.WriteHeader(http.StatusOK)
	_, _ = io.Copy(responseWriter, buf)
	return
}

func (controller *TwoFactorAuthController) Validate(ctx *gin.Context) {
	otpConfig := &dgoogauth.OTPConfig{
		Secret:      secret,
		WindowSize:  3,
		HotpCounter: 0,
	}
	otpValue := ctx.DefaultQuery("otp_value", "")

	isAuth, err := otpConfig.Authenticate(otpValue)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if !isAuth {
		helpers.NewResponse(ctx, http.StatusUnauthorized, fmt.Sprintf("failed to authenticate, try again"))
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, fmt.Sprintf("success to authenticate"))
	return
}

func (controller *TwoFactorAuthController) TestMiddleware(ctx *gin.Context) {
	helpers.NewResponse(ctx, http.StatusOK, fmt.Sprintf("success to authenticate"))
	return
}
