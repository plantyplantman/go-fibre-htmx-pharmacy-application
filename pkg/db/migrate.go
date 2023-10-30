package db

import (
	"github.com/plantyplantman/bcapi/pkg/entities"
	"gorm.io/gorm"
)

var models = []interface{}{
	&entities.Product{},
	&entities.StockInformation{},
	&entities.PromoInformation{},
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(models...)
}
