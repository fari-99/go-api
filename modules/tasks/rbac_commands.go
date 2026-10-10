package tasks

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	gohelper "github.com/fari-99/go-helper"
	"github.com/spf13/cast"
	"github.com/urfave/cli/v3"
	"gorm.io/gorm"

	"go-api/constant/constant_models"
	"go-api/modules/configs"
	"go-api/modules/models"
)

// adminPolicyRoute / adminPolicyMethod are regex patterns understood by the
// RouteMatch function and the regexMatch matcher; "/.+" matches every path.
const (
	adminPolicyRoute  = "/.+"
	adminPolicyMethod = ".*"
)

func (base *BaseCommand) getRBACCommands() []*cli.Command {
	return []*cli.Command{
		{
			Name:        "rbac-seed",
			Usage:       "rbac-seed --email admin@example.com --password '...'",
			Description: "Idempotently create the admin role, its allow-all policy and a first admin user",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "role", Value: "admin", Usage: "role name"},
				&cli.IntFlag{Name: "role-type", Value: 1, Usage: "role type (1 = Customer, 2 = Seller); part of the Casbin subject"},
				&cli.StringFlag{Name: "email", Usage: "email of the admin user (required)"},
				&cli.StringFlag{Name: "username", Usage: "username of the admin user (default: part of the email before @)"},
				&cli.StringFlag{Name: "password", Usage: "password, only used when the user does not exist yet (default: env RBAC_SEED_PASSWORD)"},
			},
			Action: func(ctx context.Context, command *cli.Command) error {
				return rbacSeed(command)
			},
		},
	}
}

func rbacSeed(command *cli.Command) error {
	roleName := command.String("role")
	roleType := command.Int("role-type")
	email := command.String("email")

	userTypes := constant_models.GetUserTypes()
	userTypeName, ok := userTypes[int(roleType)]
	if !ok {
		return fmt.Errorf("unknown role-type %d", roleType)
	}
	if roleName == "" || email == "" {
		return errors.New("--role and --email are required")
	}

	username := command.String("username")
	if username == "" {
		for i, c := range email {
			if c == '@' {
				username = email[:i]
				break
			}
		}
	}

	db := configs.DatabaseBase(configs.MySQLType).GetMysqlConnection(true)

	err := db.Transaction(func(tx *gorm.DB) error {
		// role
		var role models.Roles
		err := tx.Where("role_name = ? AND role_type = ?", roleName, roleType).First(&role).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			role = models.Roles{RoleName: roleName, RoleType: int8(roleType)}
			if err = tx.Create(&role).Error; err != nil {
				return err
			}
			log.Printf("created role %s (id %d)", roleName, role.ID)
		} else if err != nil {
			return err
		} else {
			log.Printf("role %s already exists (id %d)", roleName, role.ID)
		}

		// user
		var user models.Users
		err = tx.Where("email = ?", email).First(&user).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			password := command.String("password")
			if password == "" {
				password = os.Getenv("RBAC_SEED_PASSWORD")
			}
			if password == "" {
				return errors.New("user does not exist: provide --password or RBAC_SEED_PASSWORD")
			}

			hashed, err := gohelper.GeneratePassword(gohelper.Passwords{
				Email: email, Username: username, Password: password,
			}, cast.ToInt8(os.Getenv("PASSWORD_COST")))
			if err != nil {
				return err
			}

			user = models.Users{Username: username, Email: email, Password: *hashed, Status: 99}
			if err = tx.Create(&user).Error; err != nil {
				return err
			}
			log.Printf("created user %s (id %d)", email, user.ID)
		} else if err != nil {
			return err
		} else {
			log.Printf("user %s already exists (id %d), password untouched", email, user.ID)
		}

		// user <-> role
		var count int64
		if err = tx.Model(&models.UserRoles{}).Where("user_id = ? AND role_id = ?", user.ID, role.ID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err = tx.Create(&models.UserRoles{UserID: user.ID, RoleID: role.ID}).Error; err != nil {
				return err
			}
			log.Printf("assigned role %s to %s", roleName, email)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// policy, subject format must match populateUserRoles: "{RoleName}-{UserType}"
	subject := fmt.Sprintf("%s-%s", roleName, userTypeName)
	added, err := configs.GetPermissionInstance().AddPolicy(subject, adminPolicyRoute, adminPolicyMethod)
	if err != nil {
		return err
	}
	if added {
		log.Printf("added policy: %s %s %s", subject, adminPolicyRoute, adminPolicyMethod)
	} else {
		log.Printf("policy already exists: %s %s %s", subject, adminPolicyRoute, adminPolicyMethod)
	}

	log.Printf("rbac seed done; log in as %s, then set RBAC_ENABLED=true (or remove it) and restart", email)
	return nil
}
