package configs

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	sqladapter "github.com/Blank-Xu/sql-adapter"
	"github.com/casbin/casbin/v2"
)

type permissionUtil struct {
	enforcer *casbin.Enforcer
}

var permissionInstance *permissionUtil
var permissionOnce sync.Once

func GetPermissionInstance() *casbin.Enforcer {
	permissionOnce.Do(func() {
		databaseConfig := DatabaseBase()
		db, err := databaseConfig.SetConnection()

		if err != nil {
			panic(err)
		}

		isDebug, _ := strconv.ParseBool(os.Getenv("DATABASE_DEBUG"))
		maxLifetime, _ := strconv.ParseInt(os.Getenv("DATABASE_MAX_CONNECTION_LIFETIME_MYSQL"), 10, 64)
		maxIdleConn, _ := strconv.ParseInt(os.Getenv("DATABASE_MAX_IDLE_CONNECTION_MYSQL"), 10, 64)
		maxOpenConn, _ := strconv.ParseInt(os.Getenv("DATABASE_MAX_OPEN_CONNECTION_MYSQL"), 10, 64)

		db.DB().SetConnMaxLifetime(time.Second * time.Duration(maxLifetime)) // sets the maximum amount of time a connection may be reused.
		db.DB().SetMaxIdleConns(int(maxIdleConn))                            // sets the maximum number of connections in the idle
		db.DB().SetMaxOpenConns(int(maxOpenConn))                            // sets the maximum number of open connections to the database.
		db.SingularTable(true)                                               // Set as singular table
		db.LogMode(isDebug)                                                  // check database log mode

		adapterSQL, err := sqladapter.NewAdapter(db.DB(), "mysql", "api_rule_access")
		if err != nil {
			panic(err)
		}

		enforcer, err := casbin.NewEnforcer("./modules/configs/rbac_model.conf", adapterSQL)
		if err != nil {
			panic(err)
		}

		// Load the policy from DB.
		if err = enforcer.LoadPolicy(); err != nil {
			fmt.Println("LoadPolicy failed, err: ", err)
		}

		enforcer.AddFunction("RouteMatch", RouteMatchFunction)

		permissionInstance = &permissionUtil{
			enforcer: enforcer,
		}
	})

	return permissionInstance.enforcer
}

func routeMatch(key1 string, key2 string) bool {
	key2 = normalizeTemplateUrl(key2)
	key2 = strings.Replace(key2, "/*", "/.*", -1)

	var tokens []string

	re := regexp.MustCompile(`\{([^/]+)\}`)
	key2 = re.ReplaceAllStringFunc(key2, func(s string) string {
		tokens = append(tokens, s[1:len(s)-1])
		return "([^/]+)"
	})

	re = regexp.MustCompile("^" + key2 + "$")
	matches := re.FindStringSubmatch(key1)
	if matches == nil {
		return false
	}
	matches = matches[1:]

	if len(tokens) != len(matches) {
		panic(errors.New("RouteMatch: number of tokens is not equal to number of values"))
	}

	values := map[string]string{}

	for key, token := range tokens {
		if _, ok := values[token]; !ok {
			values[token] = matches[key]
		}
		if values[token] != matches[key] {
			return false
		}
	}
	return true
}

func normalizeTemplateUrl(key2 string) string {
	key2 = strings.TrimRight(key2, "*")
	template, start, end := "", -1, -1
	for i := 0; i < len(key2); i++ {
		if key2[i] == '<' && start < 0 {
			start = i
		} else if key2[i] == '>' && start >= 0 {
			name := key2[start+1 : i]
			for j := start + 1; j < i; j++ {
				if key2[j] == ':' {
					name = key2[start+1 : j]
					break
				}
			}
			template += key2[end+1:start] + "{" + name + "}"
			end = i
			start = -1
		}
	}
	if end < 0 {
		template = key2
	} else if end < len(key2)-1 {
		template += key2[end+1:]
	}

	return template
}

func RouteMatchFunction(args ...interface{}) (interface{}, error) {
	name1 := args[0].(string)
	name2 := args[1].(string)

	return routeMatch(name1, name2), nil
}
