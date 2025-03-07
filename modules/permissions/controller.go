package permissions

import (
	"fmt"
	"net/http"
	"strings"

	"go-api/constant/constant_models"
	"go-api/helpers"
	"go-api/modules/configs"
	"go-api/modules/middleware"

	"github.com/gin-gonic/gin"
)

type controller struct {
	service Service
}

func (r controller) CheckAction(ctx *gin.Context) {
	var input CheckPermissions
	if err := ctx.BindJSON(&input); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
			"error":         err.Error(),
			"error_message": "invalid input check permission",
		})
		return
	}

	uuid, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(uuid.(string))
	enforcer := configs.GetPermissionInstance()
	roles := strings.Split(currentUser.Roles, ",")

	for _, role := range roles {
		permission := middleware.Permission{
			Subject: role,
			Object:  input.Path,
			Action:  input.Method,
		}

		if rolePermission, err := middleware.CheckPermission(enforcer, permission); err != nil {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "error check permission",
				"error":         err.Error(),
			})
			return
		} else if rolePermission {
			helpers.NewResponse(ctx, http.StatusBadRequest, "you have permission to access this route")
			return
		}
	}

	helpers.NewResponse(ctx, http.StatusBadRequest, "you don't have any permission to access this route")
	return
}

func (r controller) GetAction(ctx *gin.Context) {
	enforcer := configs.GetPermissionInstance()
	allRoutes := configs.GetGinApplication().Routes()

	// get all roles
	userRoles, err := r.service.GetRoles(ctx)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
			"error_message": "failed to get user roles",
			"error":         err.Error(),
		})
		return
	}

	// get all user type
	userTypes := constant_models.GetUserTypes()

	type Subject struct {
		Name    string `json:"name"`
		Enabled bool   `json:"enabled"`
	}

	type Data struct {
		Routes  string    `json:"routes"`
		Method  string    `json:"method"`
		Subject []Subject `json:"subject"`
	}

	var allData []Data
	for _, route := range allRoutes {
		routePath := route.Path
		method := route.Method

		var subjects []Subject
		for _, userRole := range userRoles {
			roleName := fmt.Sprintf("%s-%s", userRole.RoleName, userTypes[int(userRole.RoleType)])
			has, _ := enforcer.Enforce(roleName, routePath, method) // check if enforce working or not
			subject := Subject{
				Name:    roleName,
				Enabled: has,
			}

			subjects = append(subjects, subject)
		}

		data := Data{
			Routes:  routePath,
			Method:  method,
			Subject: subjects,
		}

		allData = append(allData, data)
	}

	helpers.NewResponse(ctx, http.StatusOK, allData)
	return
}

func (r controller) EditAction(ctx *gin.Context) {
	var input EditPermissions
	if err := ctx.BindJSON(&input); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
			"error":         err.Error(),
			"error_message": "invalid input edit permission",
		})
		return
	}

	oldPolicy := []string{input.OldPolicy.Subject, input.NewPolicy.Route, input.NewPolicy.Method}
	newPolicy := []string{input.NewPolicy.Subject, input.NewPolicy.Route, input.NewPolicy.Method}

	enforcer := configs.GetPermissionInstance()
	if input.NewPolicy.PType == "p" {
		hasPolicy, err := enforcer.HasPolicy(input.OldPolicy.Subject, input.NewPolicy.Route, input.NewPolicy.Method)
		if err != nil {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "error check permission",
				"error":         err.Error(),
			})
			return
		} else if !hasPolicy {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "error check permission",
				"error":         "policy not found",
			})
			return
		}

		success, err := enforcer.UpdatePolicy(oldPolicy, newPolicy)
		if err != nil {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "error when trying to update permission",
				"error":         err.Error(),
			})
			return
		} else if !success {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "failed to update policy rules, please try again",
			})
			return
		}
	} else if input.NewPolicy.PType == "g" {
		hasPolicy, err := enforcer.HasGroupingPolicy(input.OldPolicy.Subject, input.NewPolicy.Route, input.NewPolicy.Method)
		if err != nil {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "error check permission",
				"error":         err.Error(),
			})
			return
		} else if !hasPolicy {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "error check permission",
				"error":         "group policy not found",
			})
			return
		}

		success, err := enforcer.UpdateGroupingPolicy(oldPolicy, newPolicy)
		if err != nil {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "error when trying to update group permission",
				"error":         err.Error(),
			})
			return
		} else if !success {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "failed to update group policy rules, please try again",
			})
			return
		}
	} else {
		helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
			"error_message": "policy type is invalid, must be p or g",
		})
		return
	}

	err := enforcer.SavePolicy()
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, "failed saving updated policy, please try again")
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "success update policy rule")
	return
}

func (r controller) DeleteAction(ctx *gin.Context) {
	var input DeletePermissions
	if err := ctx.BindJSON(&input); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
			"error":         err.Error(),
			"error_message": "invalid input edit permission",
		})
		return
	}

	enforcer := configs.GetPermissionInstance()
	if input.PType == "p" {
		hasPolicy, err := enforcer.HasPolicy(input.Subject, input.Route, input.Method)
		if err != nil {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "error check permission",
				"error":         err.Error(),
			})
			return
		} else if !hasPolicy {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "error check permission",
				"error":         "policy not found",
			})
			return
		}

		success, err := enforcer.RemovePolicy(input.Subject, input.Route, input.Method)
		if err != nil {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "error when trying to remove permission",
				"error":         err.Error(),
			})
			return
		} else if !success {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "failed to remove policy rules, please try again",
			})
			return
		}
	} else if input.PType == "g" {
		hasPolicy, err := enforcer.HasGroupingPolicy(input.Subject, input.Route, input.Method)
		if err != nil {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "error check permission",
				"error":         err.Error(),
			})
			return
		} else if !hasPolicy {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "error check permission",
				"error":         "group policy not found",
			})
			return
		}

		success, err := enforcer.RemoveGroupingPolicy(input.Subject, input.Route, input.Method)
		if err != nil {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "error when trying to delete group permission",
				"error":         err.Error(),
			})
			return
		} else if !success {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "failed to delete group policy rules, please try again",
			})
			return
		}
	} else {
		helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
			"error_message": "policy type is invalid, must be p or g",
		})
		return
	}

	err := enforcer.SavePolicy()
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, "failed saving deleted policy, please try again")
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "success delete policy permissions")
	return
}

func (r controller) CreateAction(ctx *gin.Context) {
	var input CreatePermissions
	if err := ctx.BindJSON(&input); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
			"error":         err.Error(),
			"error_message": "invalid input create permission",
		})
		return
	}

	enforcer := configs.GetPermissionInstance()
	if input.PType == "p" {
		hasPolicy, err := enforcer.HasPolicy(input.Subject, input.Route, input.Method)
		if err != nil {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "error check permission",
				"error":         err.Error(),
			})
			return
		} else if hasPolicy {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "error check permission",
				"error":         "policy already created",
			})
			return
		}

		success, err := enforcer.AddPolicy(input.Subject, input.Route, input.Method)
		if err != nil {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "error when trying to add permission",
				"error":         err.Error(),
			})
			return
		} else if !success {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "failed to add policy rules, please try again",
			})
			return
		}
	} else if input.PType == "g" {
		hasPolicy, err := enforcer.HasGroupingPolicy(input.Subject, input.Route, input.Method)
		if err != nil {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "error check permission",
				"error":         err.Error(),
			})
			return
		} else if hasPolicy {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "error check permission",
				"error":         "group policy already created",
			})
			return
		}

		success, err := enforcer.AddGroupingPolicy(input.Subject, input.Route, input.Method)
		if err != nil {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "error when trying to add group permission",
				"error":         err.Error(),
			})
			return
		} else if !success {
			helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
				"error_message": "failed to add group policy rules, please try again",
			})
			return
		}
	} else {
		helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
			"error_message": "policy type is invalid, must be p or g",
		})
		return
	}

	err := enforcer.SavePolicy()
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, "failed saving created policy, please try again")
		return
	}

	helpers.NewResponse(ctx, http.StatusCreated, "success create policy permissions")
	return
}
