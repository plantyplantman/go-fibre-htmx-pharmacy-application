package db_test

import (
	"strings"
	"testing"

	"github.com/plantyplantman/bcapi/api/presenter"
	"github.com/plantyplantman/bcapi/pkg/product"
	"github.com/plantyplantman/bcapi/pkg/report"
	"gorm.io/gorm"
)

func TestNewRangeReview(t *testing.T) {
	service, err := product.NewDefaultService()
	if err != nil {
		t.Fatal(err)
	}

	ps, err := fetchProductsByBrand("duro-tuss", service)
	if err != nil {
		t.Fatal(err)
	}

	report, err := report.NewRangeReview(ps)
	if err != nil {
		t.Fatal(err)
	}

	if len(report) == 0 {
		t.Fatal("report is empty")
	}
}

func fetchProductsByBrand(brand string, service product.Service) ([]*presenter.Product, error) {
	products, err := service.FetchProducts(
		func(d *gorm.DB) *gorm.DB {
			return d.Where("name LIKE ?", "%"+strings.ToUpper(strings.TrimSpace(brand))+"%")
		},
		product.WithStockInformation())
	if err != nil {
		return nil, err
	}
	return products, nil
}
