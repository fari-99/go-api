package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"

	"go-api/helpers"
	"go-api/modules/configs"

	"github.com/gin-gonic/gin"

	"github.com/casbin/casbin/v3"
	_ "github.com/go-sql-driver/mysql"
)

// PermissionMiddleware returns the RBAC handler. RBAC_ENABLED=false turns the
// Casbin check off (the auth middleware still applies) so a fresh install can
// be bootstrapped; it is read once at startup, so changing it needs a restart.
func PermissionMiddleware() gin.HandlerFunc {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("RBAC_ENABLED")), "false") {
		log.Println("WARNING: RBAC_ENABLED=false, permission checks are DISABLED; any logged-in user can call every gated route")
		return func(ctx *gin.Context) {}
	}

	return RBACHandler
}

func RBACHandler(ctx *gin.Context) {
	enforcer := configs.GetPermissionInstance()

	// must run after the auth middleware, which sets "uuid"
	uuid, ok := ctx.Get("uuid")
	uuidStr, isString := uuid.(string)
	if !ok || !isString || uuidStr == "" {
		helpers.NewResponse(ctx, http.StatusUnauthorized, "You must login to access")
		ctx.Abort()
		return
	}

	currentUser, err := helpers.GetCurrentUser(ctx, uuidStr)
	if err != nil || currentUser == nil {
		helpers.NewResponse(ctx, http.StatusUnauthorized, "You must login to access")
		ctx.Abort()
		return
	}

	// subject := currentUser.GetSubject()
	// drop blank entries: strings.Split("", ",") yields [""], and Casbin allows an
	// empty subject when no policy rows are loaded, so a role-less user would pass
	var roles []string
	for _, role := range strings.Split(currentUser.Roles, ",") {
		if role = strings.TrimSpace(role); role != "" {
			roles = append(roles, role)
		}
	}

	routes := ctx.Request.URL.Path
	method := ctx.Request.Method

	// fmt.Printf("------------ subject := %s\n", subject)
	// fmt.Printf("------------ roles := %v\n", roles)
	// fmt.Printf("------------ routes := %s\n", routes)
	// fmt.Printf("------------ method := %s\n", method)

	// check subject permission
	// if subjectPermission, err := checkPermission(subject, routes, method); err != nil {
	//	return err
	// } else if !subjectPermission {
	//	//fmt.Printf("--- SUBJECT DON'T HAVE PERMISSION ---")
	//	return routing.NewHTTPError(http.StatusUnauthorized, "user don't have permission for this url")
	// }

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
			// gin continues to the next handler when this one returns
			return
		}
	}

	// fmt.Printf("--- DON'T HAVE PERMISSION FOR ANY ROLES ---")
	helpers.NewResponse(ctx, http.StatusForbidden, "user don't have role that have permission for this url")
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
