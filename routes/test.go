package routes

import (
	"go-api/middleware"
	"go-api/test_controllers"
	"log"
)

func (routes *Routes) setupTestRoute() {
	app := routes.ginApp

	// Redis Test Endpoint collection
	testRedis := app.Group("/test-redis")
	{
		log.Println("Setup Test Redis router")

		testRedisController := &test_controllers.TestRedisController{
			DB:    routes.DB,
			Redis: routes.Redis,
		}

		testRedis.POST("/", testRedisController.TestRedisAction)
	}

	// Redis Cache Test Endpoint collection
	redisCache := app.Group("/test-redis-cache")
	{
		log.Println("Setup Test Redis Cache router")

		testRedisController := &test_controllers.RedisCacheController{
			DB:         routes.DB,
			RedisCache: routes.RedisCache,
		}

		redisCache.POST("/", testRedisController.TestRedisCacheAction)
	}

	// Redis Cache Test Endpoint collection
	testRabbit := app.Group("/test-rabbit")
	{
		log.Println("Setup Test RabbitMq Connection router")

		testRabbitMqQueueController := &test_controllers.RabbitMqController{
			QueueSetup: routes.Queue,
		}

		testRabbit.POST("/queue", testRabbitMqQueueController.TestPublishQueueAction)
		testRabbit.POST("/batch-queue", testRabbitMqQueueController.TestBatchPublishQueueAction)
	}

	testEmail := app.Group("/test-email")
	{
		log.Printf("Setup Test EmailDialler Connection router")

		testEmailController := &test_controllers.EmailsController{
			DB:          routes.DB,
			EmailDialer: routes.EmailDialler,
		}

		testEmail.POST("/send-email", testEmailController.SendEmailAction)
	}

	testStateMachine := app.Group("/test-state-machine")
	{
		stateMachineController := &test_controllers.FiniteStateController{
			DB: routes.DB,
		}

		testStateMachine.POST("/get-state", stateMachineController.GetAvailableTransitionsAction)
		testStateMachine.POST("/change-state", stateMachineController.ChangeStateAction)
	}

	testFtp := app.Group("/test-ftp")
	{
		log.Println("Setup Test FTP router")

		ftpController := &test_controllers.FtpController{
			DB: routes.DB,
		}

		testFtp.POST("/send-files", ftpController.SendFtpAction)
		//testFtp.POST("/send-files-location")
		//testFtp.POST("/send-files-open-files")
	}

	testCrypt := app.Group("/test-crypt")
	{
		log.Println("Setup Test Encryption router")
		cryptController := &test_controllers.CryptsController{}

		testCrypt.POST("/encrypt-data", cryptController.EncryptDecryptAction)
		testCrypt.POST("/encrypt-rsa", cryptController.EncryptDecryptRsaAction)
		testCrypt.POST("/sign-message", cryptController.SignMessageAction)
		testCrypt.POST("/verify-message", cryptController.VerifyMessageAction)

		// TODO: Encrypt Files
	}

	twoFactorAuth := app.Group("/test-two-auth")
	{
		log.Println("Setup Test 2FA router")
		twoFactorAuthController := &test_controllers.TwoFactorAuthController{
			DB: routes.DB,
		}

		twoFactorAuth.POST("/new", twoFactorAuthController.NewAuth)
		twoFactorAuth.POST("/validate", twoFactorAuthController.Validate)

		otpMiddleware := middleware.OTPMiddleware(middleware.BaseMiddleware{})
		authMiddleware := middleware.AuthMiddleware(middleware.BaseMiddleware{})
		twoFactorAuth.Use(authMiddleware, otpMiddleware).GET("/test", twoFactorAuthController.TestMiddleware)
	}
}
