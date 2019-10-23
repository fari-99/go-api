package utils

import (
	"log"
	"sync"

	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"

	"github.com/kataras/iris"
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
