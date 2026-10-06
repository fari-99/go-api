package roles

import (
	"fmt"

	paginator "github.com/dmitryburov/gorm-paginator"
	"github.com/gin-gonic/gin"

	"go-api/constant/constant_models"
	"go-api/modules/configs"
	"go-api/modules/models"
)

type Service interface {
	GetDetail(ctx *gin.Context, id string) (*models.Roles, bool, error)
	GetList(ctx *gin.Context, filter RequestListFilter) ([]models.Roles, *paginator.Pagination, error)
	Create(ctx *gin.Context, input RequestRole) (*models.Roles, error)
	Update(ctx *gin.Context, id string, input RequestRole) (*models.Roles, error)
	CountPermissions(ctx *gin.Context, id string) (int, error)
	Delete(ctx *gin.Context, id string) error
}

// permissionSubject builds the Casbin subject string for a role, matching
// how the permissions module composes it ("{RoleName}-{UserTypeName}").
func permissionSubject(model models.Roles) string {
	userTypes := constant_models.GetUserTypes()
	return fmt.Sprintf("%s-%s", model.RoleName, userTypes[int(model.RoleType)])
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return service{repo: repo}
}

func (s service) GetDetail(ctx *gin.Context, id string) (*models.Roles, bool, error) {
	return s.repo.GetDetail(ctx, id)
}

func (s service) GetList(ctx *gin.Context, filter RequestListFilter) ([]models.Roles, *paginator.Pagination, error) {
	return s.repo.GetList(ctx, filter)
}

func (s service) Create(ctx *gin.Context, input RequestRole) (*models.Roles, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	model := models.Roles{
		RoleType: input.RoleType,
		RoleName: input.RoleName,
	}

	return s.repo.Create(ctx, model)
}

func (s service) Update(ctx *gin.Context, id string, input RequestRole) (*models.Roles, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	model, notFound, err := s.repo.GetDetail(ctx, id)
	if err != nil {
		return nil, err
	} else if notFound {
		return nil, fmt.Errorf("role %s does not exist", id)
	}

	oldSubject := permissionSubject(*model)

	model.RoleType = input.RoleType
	model.RoleName = input.RoleName

	newSubject := permissionSubject(*model)

	// role_name/role_type changed the Casbin subject string ("{RoleName}-{UserType}"),
	// so any existing permissions must be migrated to the new subject or they'd be
	// silently orphaned (the renamed role would appear to have lost every permission).
	if oldSubject != newSubject {
		if err = migratePermissionSubject(oldSubject, newSubject); err != nil {
			return nil, err
		}
	}

	return s.repo.Update(ctx, *model)
}

func (s service) CountPermissions(ctx *gin.Context, id string) (int, error) {
	model, notFound, err := s.repo.GetDetail(ctx, id)
	if err != nil {
		return 0, err
	} else if notFound {
		return 0, fmt.Errorf("role id [%s] does not exist", id)
	}

	subject := permissionSubject(*model)
	enforcer := configs.GetPermissionInstance()

	// Only "p" policies are counted: this app's Casbin model has no
	// [role_definition] ("g") section, so grouping policies aren't functional.
	policies, err := enforcer.GetFilteredPolicy(0, subject)
	if err != nil {
		return 0, err
	}

	return len(policies), nil
}

func (s service) Delete(ctx *gin.Context, id string) error {
	model, notFound, err := s.repo.GetDetail(ctx, id)
	if err != nil {
		return err
	} else if notFound {
		return fmt.Errorf("role id [%s] does not exist", id)
	}

	subject := permissionSubject(*model)
	enforcer := configs.GetPermissionInstance()

	if _, err = enforcer.RemoveFilteredPolicy(0, subject); err != nil {
		return err
	}

	return s.repo.Delete(ctx, id)
}

// toPolicyArgs converts a policy row ([]string) into the ...interface{} shape
// the Casbin management API's Add/Remove policy methods expect.
func toPolicyArgs(policy []string) []interface{} {
	args := make([]interface{}, len(policy))
	for i, v := range policy {
		args[i] = v
	}
	return args
}

// migratePermissionSubject rewrites every "p" policy whose subject (field index 0)
// is oldSubject so its subject becomes newSubject, preserving the permissions a
// role had before it was renamed/retyped. Grouping ("g") policies are not handled:
// this app's Casbin model has no [role_definition] section, so they aren't functional.
func migratePermissionSubject(oldSubject, newSubject string) error {
	enforcer := configs.GetPermissionInstance()

	policies, err := enforcer.GetFilteredPolicy(0, oldSubject)
	if err != nil {
		return err
	}
	for _, policy := range policies {
		newPolicy := append([]string{}, policy...)
		newPolicy[0] = newSubject

		if _, err = enforcer.RemovePolicy(toPolicyArgs(policy)...); err != nil {
			return err
		}
		if _, err = enforcer.AddPolicy(toPolicyArgs(newPolicy)...); err != nil {
			return err
		}
	}

	return nil
}
