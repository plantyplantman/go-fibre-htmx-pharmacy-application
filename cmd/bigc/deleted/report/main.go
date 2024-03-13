package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/plantyplantman/bcapi/api/presenter"
	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/plantyplantman/bcapi/pkg/env"
	"github.com/plantyplantman/bcapi/pkg/product"
	"github.com/samber/lo"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type RETV struct {
	presenter.Product
	PercentOff float64
	SalePrice  float64
}

func main() {
	all()
}

func all() {
	var (
		date    = time.Now().Format("060102")
		outPath = filepath.Join(`C:\Users\admin\Develin Management Dropbox\Zihan\files\out`, date, date+`__deleted-report.csv`)
	)

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

	pps, err := service.FetchProducts(product.WithStockInformation(), func(d *gorm.DB) *gorm.DB {
		return d.Where("name LIKE ?", `\\%`)
	})
	if err != nil {
		log.Fatalln(err)
	}
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

func webOnly() {
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

	retv := lo.Associate(pps, func(p *presenter.Product) (string, *RETV) {
		return p.BCID, &RETV{
			*p,
			0,
			0,
		}
	})

	pps = lo.Filter(pps, func(p *presenter.Product, _ int) bool {
		return p.StockInformation.Total > 0
	})

	pids := lo.FilterMap(pps, func(p *presenter.Product, _ int) (int, bool) {
		if len(strings.Split(p.BCID, "/")) > 1 {
			return 0, false
		}

		if id, err := strconv.Atoi(p.BCID); err == nil {
			return id, true
		}

		return 0, false
	})

	bc := bigc.MustGetClient()
	for _, pid := range pids {
		p, err := bc.GetProductById(pid)
		if err != nil {
			log.Println(err)
			continue
		}
		retv[strconv.Itoa(p.ID)].SalePrice = p.SalePrice
		retv[strconv.Itoa(p.ID)].PercentOff = (p.Price - p.SalePrice) / p.Price * 100
	}

	vids := lo.FilterMap(pps, func(p *presenter.Product, _ int) ([]int, bool) {
		if len(strings.Split(p.BCID, "/")) > 1 {
			pid, err := strconv.Atoi(strings.Split(p.BCID, "/")[0])
			if err != nil {
				return nil, false
			}

			vid, err := strconv.Atoi(strings.Split(p.BCID, "/")[1])
			if err != nil {
				return nil, false
			}

			return []int{pid, vid}, true
		}

		return nil, false
	})

	for _, vid := range vids {
		vp, err := bc.GetVariantById(vid[1], vid[0], map[string]string{})
		if err != nil {
			log.Println(err)
			continue
		}

		retv[strconv.Itoa(vp.ProductID)+"/"+strconv.Itoa(vp.ID)].SalePrice = vp.SalePrice
		retv[strconv.Itoa(vp.ProductID)+"/"+strconv.Itoa(vp.ID)].PercentOff = (vp.Price - vp.SalePrice) / vp.Price * 100
	}

	f, err := os.Create(outPath)
	if err != nil {
		log.Fatalln(err)
	}

	err = gocsv.Marshal(lo.Values(retv), f)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("exported to", outPath)
}
