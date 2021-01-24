package test_controllers

import (
	"encoding/base64"
	"github.com/kataras/iris/v12"
	"go-api/configs"
	"go-api/helpers"
)

type CryptsController struct{}

type InputEncryptDecrypt struct {
	Passphrase    string `json:"passphrase"` // general encrypt
	Data          string `json:"data"`
	RsaRandomness string `json:"rsa_randomness"` // rsa
}

func (controller *CryptsController) EncryptDecryptAction(ctx iris.Context) {
	var input InputEncryptDecrypt
	_ = ctx.ReadJSON(&input)

	encryptHelper := helpers.NewEncryptionBase()
	encryptHelper.SetPassphrase(input.Passphrase)
	encrypted, err := encryptHelper.Encrypt([]byte(input.Data))
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, err.Error())
		return
	}

	decryptHelper := helpers.NewEncryptionBase()
	decryptHelper.SetPassphrase(input.Passphrase)
	decrypted, err := decryptHelper.Decrypt(encrypted)
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, err.Error())
		return
	}

	result := map[string]interface{}{
		"encrypt": string(encrypted),
		"decrypt": string(decrypted),
	}

	_, _ = configs.NewResponse(ctx, iris.StatusOK, result)
	return
}

func (controller *CryptsController) EncryptDecryptRsaAction(ctx iris.Context) {
	var input InputEncryptDecrypt
	_ = ctx.ReadJSON(&input)

	var useRandomness bool
	if input.RsaRandomness != "" {
		useRandomness = true
	}

	encryptHelper := helpers.NewEncryptionBase()
	encryptHelper.SetUseRandomness(useRandomness, input.RsaRandomness)
	encryptHelper.SetPassphrase(input.Passphrase)
	privateKey, publicKey, err := encryptHelper.GenerateRSAKey()
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, err.Error())
		return
	}

	encrypted, err := encryptHelper.EncryptRSA([]byte(input.Data))
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, err.Error())
		return
	}

	decryptHelper := helpers.NewEncryptionBase()
	decryptHelper.SetUseRandomness(useRandomness, input.RsaRandomness)
	encryptHelper.SetPassphrase(input.Passphrase)
	decryptHelper.SetRsaKey(privateKey, publicKey)

	decrypted, err := decryptHelper.DecryptRSA(string(encrypted))
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, err.Error())
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

	_, _ = configs.NewResponse(ctx, iris.StatusOK, result)
	return
}

type InputSignMessage struct {
	Messages string `json:"messages"`
}

func (controller *CryptsController) SignMessageAction(ctx iris.Context) {
	var input InputSignMessage
	_ = ctx.ReadJSON(&input)

	signHelper := helpers.NewEncryptionBase()
	signature, err := signHelper.SignData(input.Messages)
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, err.Error())
		return
	}

	result := map[string]interface{}{
		"message":   input.Messages,
		"signature": signature,
	}

	_, _ = configs.NewResponse(ctx, iris.StatusOK, result)
	return
}

type InputVerifyMessage struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

func (controller *CryptsController) VerifyMessageAction(ctx iris.Context) {
	var input InputVerifyMessage
	_ = ctx.ReadJSON(&input)

	signature, _ := base64.RawURLEncoding.DecodeString(input.Signature)

	signHelper := helpers.NewEncryptionBase()
	isVerified, err := signHelper.VerifyData(input.Message, string(signature))
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, err.Error())
		return
	}

	if !isVerified {
		_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, "this message is not valid")
		return
	}

	_, _ = configs.NewResponse(ctx, iris.StatusOK, "this message is valid")
	return
}
