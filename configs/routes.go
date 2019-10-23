package configs

import (
	"goService/utils"
	"os"

	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"

	_ "goService/routes"
)

type Routes struct {
	DB *gorm.DB
}

/**
 * setup routers
 * @return void
 */
func (r *Routes) Setup(host string, port string) {

	//setup routers
	app := utils.GetIrisApplication()

	// Set logging level
	app.Logger().SetLevel(os.Getenv("LOG_LEVEL"))

	//start server
	app.Run(iris.Addr(host+":"+port), iris.WithoutServerError(iris.ErrServerClosed))
}
