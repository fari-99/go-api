package middleware

import (
	"go-api/helpers"
	"net/http"

	"github.com/gin-gonic/gin"
)

// VersionMiddleware inits auth middleware config and returns new handler
func VersionMiddleware(versions map[string]bool) gin.HandlerFunc {
	return versionServe(versions)
}

// versionServe checks headers "X-API-Version"
// If the data is valid, continues to next handler
func versionServe(versions map[string]bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		versionHeader := ctx.GetHeader("X-API-Version")
		if versionHeader == "" {
			helpers.NewResponse(ctx, http.StatusNotFound, gin.H{
				"message": "version header is empty",
			})
			ctx.Abort()
			return
		}

		for version, isDeprecated := range versions {
			if version != versionHeader {
				helpers.NewResponse(ctx, http.StatusNotFound, gin.H{
					"message": "your request url are not found for this version or you input the wrong version for this url",
				})
				ctx.Abort()
				return
			}

			if isDeprecated {
				helpers.NewResponse(ctx, http.StatusNotFound, gin.H{
					"message": "your request url are already deprecated, please contact administrator",
				})
				ctx.Abort()
				return
			}
		}

		ctx.Next()
	}
}
