package db

import (
	"github.com/plantyplantman/bcapi/pkg/entities"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Seed(db *gorm.DB, data []*entities.Product) error {
	var batchSize = 1000
	if len(data) < 1000 {
		batchSize = len(data)
	}
	return db.Clauses(
		clause.OnConflict{
			Columns: []clause.Column{
				{Name: "sku"},
			},
			UpdateAll: true,
		},
	).CreateInBatches(data, batchSize).Error
}
