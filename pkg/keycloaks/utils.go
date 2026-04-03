package keycloaks

import (
	"fmt"
	"os"
)

func RealmExists(realms string) (isExists bool, err error) {
	if realms == "" {
		panic("realms value is empty")
	}

	switch realms {
	case os.Getenv("USER_KEYCLOAK_REALM"):
		return true, nil
	case os.Getenv("ADMIN_KEYCLOAK_REALM"):
		return true, nil
	default:
		return false, fmt.Errorf("realms [%s] not exists, or not yet setup", realms)
	}
}
