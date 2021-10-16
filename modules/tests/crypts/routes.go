package crypts

import (
	"github.com/gin-gonic/gin"
	"log"
)

func NewRoute(app *gin.Engine) {
	log.Println("Setup Test Encryption router")

	testCrypt := app.Group("/test-crypt")
	{
		cryptController := &CryptsController{}

		testCrypt.POST("/encrypt-data", cryptController.EncryptDecryptAction)
		testCrypt.POST("/encrypt-rsa", cryptController.EncryptDecryptRsaAction)
		testCrypt.POST("/sign-message", cryptController.SignMessageAction)
		testCrypt.POST("/verify-message", cryptController.VerifyMessageAction)

		// TODO: Encrypt Files
	}
}
