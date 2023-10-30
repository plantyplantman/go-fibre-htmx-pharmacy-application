package entities

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/samber/lo"
)

type Product struct {
	gorm.Model
	Sku               string `gorm:"unique"`
	Name              string
	Price             float64
	CostPrice         float64
	OnWeb             int
	IsVariant         bool
	BCID              string
	StockInformations []StockInformation `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	PromoInformation  PromoInformation   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type PromoInformation struct {
	gorm.Model
	PromoPrice float64

	ProductID int64
}

type StockInformation struct {
	gorm.Model
	Location Store
	Soh      int

	ProductID int64
}

type ProductWithCombinedStockInformation struct {
	Product
	CombinedStockInformation
}

type CombinedStockInformation struct {
	Petrie   int
	Bunda    int
	Con      int
	Franklin int
	Web      int
	Total    int
}

func NewProductWithCombinedStockInformation(p Product) ProductWithCombinedStockInformation {
	return ProductWithCombinedStockInformation{
		Product:                  p,
		CombinedStockInformation: newCombinedStockInformation(p.StockInformations),
	}
}

func newCombinedStockInformation(sis []StockInformation) CombinedStockInformation {
	agg := CombinedStockInformation{}
	lo.Reduce(
		sis,
		func(agg *CombinedStockInformation, si StockInformation, _ int) *CombinedStockInformation {
			soh := si.Soh
			switch strings.ToLower(fmt.Sprint(si.Location)) {
			case "petrie":
				agg.Petrie = soh
			case "bunda":
				agg.Bunda = soh
			case "con":
				agg.Con = soh
			case "franklin":
				agg.Franklin = soh
			}
			if soh < 0 {
				soh = 0
			}
			agg.Total = agg.Total + soh

			return agg
		},
		&agg,
	)
	return agg
}
