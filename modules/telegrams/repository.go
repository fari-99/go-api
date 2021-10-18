package telegrams

import "go-api/modules/configs"

type Repository interface {
}

type repository struct {
	*configs.DI
}

func NewRepository(di *configs.DI) Repository {
	return repository{DI: di}
}
