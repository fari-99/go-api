package configs

import (
	"log"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
)

type ginAppUtil struct {
	app *gin.Engine
}

var ginAppInstance *ginAppUtil
var onceGinApp sync.Once

// GetGinApplication get gin Application instance
func GetGinApplication() *gin.Engine {
	onceGinApp.Do(func() {
		log.Println("Initialize gin application instance...")

		app := gin.New()

		if os.Getenv("APP_MODE") == gin.DebugMode {
			gin.SetMode(gin.DebugMode)
		} else {
			gin.SetMode(gin.ReleaseMode)
		}

		// set recovery panic
		app.Use(gin.Recovery())

		// Log everything to terminal
		app.Use(gin.Logger())

		ginAppInstance = &ginAppUtil{
			app: app,
		}
	})

	return ginAppInstance.app
}
