package csrf

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/csrf"

	"go-api/helpers"
	"go-api/modules/middleware"
)

func NewCsrfRoutes(app *gin.Engine) {
	log.Println("Setup Test CSRF router")

	testCsrf := app.Group("/test-csrf")
	{
		csrfMiddleware := middleware.CsrfMiddleware()
		testCsrf.Use(csrfMiddleware)
		testCsrf.POST("/post", func(ctx *gin.Context) {
			helpers.NewResponse(ctx, http.StatusOK, gin.H{
				"message": "-CSRF VALID-",
			})
			return
		})

		testCsrf.GET("/get", func(ctx *gin.Context) {
			ctx.Writer.Header().Set("X-CSRF-Token", csrf.Token(ctx.Request))

			helpers.NewResponse(ctx, http.StatusOK, gin.H{
				"message": "Check headers to get CSRF",
			})
			return
		})
	}
}
