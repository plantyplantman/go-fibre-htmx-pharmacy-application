package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/plantyplantman/bcapi/pkg/env"
	"github.com/plantyplantman/bcapi/pkg/product"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	var (
		date    = time.Now().Format("060102")
		outPath = filepath.Join(`C:\Users\admin\Develin Management Dropbox\Zihan\files\out`, date, date+`__web__deleted-report.csv`)
	)

	// c := bigc.MustGetClient()

	// ps, err := c.GetAllProducts(map[string]string{})
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// var rv bigc.Products
	// rv = lo.Filter(ps, func(p bigc.Product, _ int) bool {
	// 	return strings.HasPrefix(p.Sku, "/")
	// })

	var connString = env.TEST_NEON
	if connString == "" {
		log.Fatalln("TEST_NEON_CONNECTION_STRING not set")
	}
	DB, err := gorm.Open(postgres.Open(connString), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	repo := product.NewRepository(DB)
	service := product.NewService(repo)

	// bcids := lo.Map(rv, func(p bigc.Product, _ int) string {
	// 	return strconv.Itoa(p.ID)
	// })

	pps, err := service.FetchProducts(product.WithStockInformation(), func(d *gorm.DB) *gorm.DB {
		return d.Where("name LIKE ?", `\\%`)
	})
	if err != nil {
		log.Fatalln(err)
	}

	// pps = lo.Filter(pps, func(p *presenter.Product, _ int) bool {
	// 	return p.StockInfomation.Total == 0
	// })

	f, err := os.Create(outPath)
	if err != nil {
		log.Fatalln(err)
	}

	err = gocsv.Marshal(pps, f)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("exported to", outPath)
}
