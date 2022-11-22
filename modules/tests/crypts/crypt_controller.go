package crypts

import (
	"encoding/base64"
	"net/http"

	"github.com/fari-99/go-helper/crypts"
	"github.com/gin-gonic/gin"

	"go-api/helpers"
	"go-api/modules/configs"
)

type CryptsController struct {
	*configs.DI
}

type InputEncryptDecrypt struct {
	Passphrase    string `json:"passphrase"` // general encrypt
	Data          string `json:"data"`
	RsaRandomness string `json:"rsa_randomness"` // rsa
}

func (controller *CryptsController) EncryptDecryptAction(ctx *gin.Context) {
	var input InputEncryptDecrypt
	_ = ctx.BindJSON(&input)

	encryptHelper := crypts.NewEncryptionBase()
	encryptHelper.SetPassphrase(input.Passphrase)
	encrypted, err := encryptHelper.Encrypt([]byte(input.Data))
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	decryptHelper := crypts.NewEncryptionBase()
	decryptHelper.SetPassphrase(input.Passphrase)
	decrypted, err := decryptHelper.Decrypt(encrypted)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	result := map[string]interface{}{
		"encrypt": string(encrypted),
		"decrypt": string(decrypted),
	}

	helpers.NewResponse(ctx, http.StatusOK, result)
	return
}

func (controller *CryptsController) EncryptDecryptRsaAction(ctx *gin.Context) {
	var input InputEncryptDecrypt
	_ = ctx.BindJSON(&input)

	var useRandomness bool
	if input.RsaRandomness != "" {
		useRandomness = true
	}

	encryptHelper := crypts.NewEncryptionBase()
	encryptHelper.SetUseRandomness(useRandomness, input.RsaRandomness)
	encryptHelper.SetPassphrase(input.Passphrase)
	rsaKeys, err := encryptHelper.GenerateRSAKey(2048)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	encrypted, err := encryptHelper.EncryptRSA([]byte(input.Data))
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	decryptHelper := crypts.NewEncryptionBase()
	decryptHelper.SetUseRandomness(useRandomness, input.RsaRandomness)
	encryptHelper.SetPassphrase(input.Passphrase)
	decryptHelper.SetRsaPublicKey(rsaKeys.PublicKey)
	decryptHelper.SetRsaPrivateKey(rsaKeys.PrivateKey)

	decrypted, err := decryptHelper.DecryptRSA(encrypted)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	result := map[string]interface{}{
		"key":     rsaKeys,
		"encrypt": string(encrypted),
		"decrypt": string(decrypted),
	}

	helpers.NewResponse(ctx, http.StatusOK, result)
	return
}

type InputSignMessage struct {
	Messages string `json:"messages"`
}

func (controller *CryptsController) SignMessageAction(ctx *gin.Context) {
	var input InputSignMessage
	_ = ctx.BindJSON(&input)

	signHelper := crypts.NewEncryptionBase()
	signature, err := signHelper.SignData(input.Messages)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	result := map[string]interface{}{
		"message":   input.Messages,
		"signature": signature,
	}

	helpers.NewResponse(ctx, http.StatusOK, result)
	return
}

type InputVerifyMessage struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

func (controller *CryptsController) VerifyMessageAction(ctx *gin.Context) {
	var input InputVerifyMessage
	_ = ctx.BindJSON(&input)

	signature, _ := base64.RawURLEncoding.DecodeString(input.Signature)

	signHelper := crypts.NewEncryptionBase()
	isVerified, err := signHelper.VerifyData(input.Message, signature)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if !isVerified {
		helpers.NewResponse(ctx, http.StatusBadRequest, "this message is not valid")
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "this message is valid")
	return
}
