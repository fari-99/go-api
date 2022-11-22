package middleware

import (
	"net/http"
	"strings"

	"go-api/helpers"
	"go-api/modules/configs"

	"github.com/gin-gonic/gin"

	"github.com/casbin/casbin/v2"
	_ "github.com/go-sql-driver/mysql"
)

func PermissionMiddleware() gin.HandlerFunc {
	return RBACHandler
}

func RBACHandler(ctx *gin.Context) {
	enforcer := configs.GetPermissionInstance()

	uuid, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(uuid.(string))

	//subject := currentUser.GetSubject()
	roles := strings.Split(currentUser.Roles, ",")

	routes := ctx.Request.URL.Path
	method := ctx.Request.Method

	//fmt.Printf("------------ subject := %s\n", subject)
	//fmt.Printf("------------ roles := %v\n", roles)
	//fmt.Printf("------------ routes := %s\n", routes)
	//fmt.Printf("------------ method := %s\n", method)

	// check subject permission
	//if subjectPermission, err := checkPermission(subject, routes, method); err != nil {
	//	return err
	//} else if !subjectPermission {
	//	//fmt.Printf("--- SUBJECT DON'T HAVE PERMISSION ---")
	//	return routing.NewHTTPError(http.StatusUnauthorized, "user don't have permission for this url")
	//}

	// check role permission
	for _, role := range roles {
		permission := Permission{
			Subject: role,
			Object:  routes,
			Action:  method,
		}

		if rolePermission, err := CheckPermission(enforcer, permission); err != nil {
			helpers.NewResponse(ctx, http.StatusUnauthorized, err.Error())
			ctx.Abort()
			return
		} else if rolePermission {
			ctx.Next()
			return
		}
	}

	//fmt.Printf("--- DON'T HAVE PERMISSION FOR ANY ROLES ---")
	helpers.NewResponse(ctx, http.StatusUnauthorized, "user don't have role that have permission for this url")
	ctx.Abort()
	return
}

type Permission struct {
	Subject string
	Object  string
	Action  string
}

func CheckPermission(enforcer *casbin.Enforcer, permission Permission) (bool, error) {
	has, err := enforcer.Enforce(permission.Subject, permission.Object, permission.Action)
	return has, err
}
