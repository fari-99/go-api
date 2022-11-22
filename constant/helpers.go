package constant

import "fmt"

func GetStatus(name string) (int8, error) {
	var status int8
	switch name {
	case StatusActiveName:
		status = StatusActive
	case StatusNonActiveName:
		status = StatusNonActive
	case StatusDeletedName:
		status = StatusDeleted
	default:
		return 0, fmt.Errorf("status name [%s] not found", name)
	}

	return status, nil
}
