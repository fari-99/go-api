package test_controllers

import (
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"go-api/configs"
	"go-api/helpers"
	"net/http"
)

type CryptsController struct{}

type InputEncryptDecrypt struct {
	Passphrase    string `json:"passphrase"` // general encrypt
	Data          string `json:"data"`
	RsaRandomness string `json:"rsa_randomness"` // rsa
}

func (controller *CryptsController) EncryptDecryptAction(ctx *gin.Context) {
	var input InputEncryptDecrypt
	_ = ctx.BindJSON(&input)

	encryptHelper := helpers.NewEncryptionBase()
	encryptHelper.SetPassphrase(input.Passphrase)
	encrypted, err := encryptHelper.Encrypt([]byte(input.Data))
	if err != nil {
		configs.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	decryptHelper := helpers.NewEncryptionBase()
	decryptHelper.SetPassphrase(input.Passphrase)
	decrypted, err := decryptHelper.Decrypt(encrypted)
	if err != nil {
		configs.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	result := map[string]interface{}{
		"encrypt": string(encrypted),
		"decrypt": string(decrypted),
	}

	configs.NewResponse(ctx, http.StatusOK, result)
	return
}

func (controller *CryptsController) EncryptDecryptRsaAction(ctx *gin.Context) {
	var input InputEncryptDecrypt
	_ = ctx.BindJSON(&input)

	var useRandomness bool
	if input.RsaRandomness != "" {
		useRandomness = true
	}

	encryptHelper := helpers.NewEncryptionBase()
	encryptHelper.SetUseRandomness(useRandomness, input.RsaRandomness)
	encryptHelper.SetPassphrase(input.Passphrase)
	privateKey, publicKey, err := encryptHelper.GenerateRSAKey()
	if err != nil {
		configs.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	encrypted, err := encryptHelper.EncryptRSA([]byte(input.Data))
	if err != nil {
		configs.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	decryptHelper := helpers.NewEncryptionBase()
	decryptHelper.SetUseRandomness(useRandomness, input.RsaRandomness)
	encryptHelper.SetPassphrase(input.Passphrase)
	decryptHelper.SetRsaKey(privateKey, publicKey)

	decrypted, err := decryptHelper.DecryptRSA(string(encrypted))
	if err != nil {
		configs.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	result := map[string]interface{}{
		"key": map[string]interface{}{
			"private": privateKey,
			"public":  publicKey,
		},
		"encrypt": string(encrypted),
		"decrypt": string(decrypted),
	}

	configs.NewResponse(ctx, http.StatusOK, result)
	return
}

type InputSignMessage struct {
	Messages string `json:"messages"`
}

func (controller *CryptsController) SignMessageAction(ctx *gin.Context) {
	var input InputSignMessage
	_ = ctx.BindJSON(&input)

	signHelper := helpers.NewEncryptionBase()
	signature, err := signHelper.SignData(input.Messages)
	if err != nil {
		configs.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	result := map[string]interface{}{
		"message":   input.Messages,
		"signature": signature,
	}

	configs.NewResponse(ctx, http.StatusOK, result)
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

	signHelper := helpers.NewEncryptionBase()
	isVerified, err := signHelper.VerifyData(input.Message, string(signature))
	if err != nil {
		configs.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if !isVerified {
		configs.NewResponse(ctx, http.StatusBadRequest, "this message is not valid")
		return
	}

	configs.NewResponse(ctx, http.StatusOK, "this message is valid")
	return
}
