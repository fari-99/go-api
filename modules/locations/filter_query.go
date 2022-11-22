package locations

import (
	"fmt"

	"gorm.io/gorm"

	"go-api/modules/models"
)

type FilterQueryLocations struct {
	Code string `json:"code"`
	Name string `json:"name"`

	Order   string `json:"order"`    // desc/asc
	OrderBy string `json:"order_by"` // column name
}

func (r repository) FilterLocations(db *gorm.DB, filter FilterQueryLocations, limit, offset int) {
	locationTable := (models.Locations{}).TableName()

	if filter.Name != "" {
		db.Where(fmt.Sprintf("%s.complete_name = ?", locationTable), filter.Name)
	}

	if filter.Code != "" {
		db.Where(fmt.Sprintf("%s.code = ?", locationTable), filter.Code)
	}

	db.Order("name asc")
	if filter.Order != "" && filter.OrderBy != "" {
		db.Order(fmt.Sprintf("%s.%s %s", locationTable, filter.OrderBy, filter.Order))
	}

	if limit != 0 {
		db.Limit(limit)
	}

	if offset != 0 {
		db.Offset(offset)
	}
}

type FilterQueryLocationLevel struct {
	Name string `json:"name"`

	Order   string `json:"order"`    // desc/asc
	OrderBy string `json:"order_by"` // column name
}

func (r repository) FilterLocationLevels(db *gorm.DB, filter FilterQueryLocationLevel, limit, offset int) *gorm.DB {
	locationTable := (models.Locations{}).TableName()

	if filter.Name != "" {
		db.Where(fmt.Sprintf("%s.complete_name = ?", locationTable), filter.Name)
	}

	db.Order("name asc")
	if filter.Order != "" && filter.OrderBy != "" {
		db.Order(fmt.Sprintf("%s.%s %s", locationTable, filter.OrderBy, filter.Order))
	}

	if limit != 0 {
		db.Limit(limit)
	}

	if offset != 0 {
		db.Offset(offset)
	}

	return db
}
