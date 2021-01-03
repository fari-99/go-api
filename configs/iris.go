package configs

import (
	"log"
	"sync"

	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"

	"github.com/kataras/iris/v12"
)

type irisAppUtil struct {
	app *iris.Application
}

var irisAppInstance *irisAppUtil
var onceIrisApp sync.Once

// GetIrisApplication get iris Application instance
func GetIrisApplication() *iris.Application {
	onceIrisApp.Do(func() {
		log.Println("Initialize iris application instance...")

		app := iris.New()

		// recover from any http-relative panics
		app.Use(recover.New())

		// Log everything to terminal
		app.Use(logger.New())

		irisAppInstance = &irisAppUtil{
			app: app,
		}
	})

	return irisAppInstance.app
}
