package main

import (
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/plantyplantman/bcapi/api/presenter"
	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/plantyplantman/bcapi/pkg/env"
	"github.com/plantyplantman/bcapi/pkg/parser"
	"github.com/plantyplantman/bcapi/pkg/product"
	"github.com/plantyplantman/bcapi/pkg/report"
	"github.com/samber/lo"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	var connString = env.TEST_NEON
	if connString == "" {
		log.Fatalln("TEST_NEON_CONNECTION_STRING not set")
	}

	c := bigc.MustGetClient()

	pfPath := `C:\Users\admin\Develin Management Dropbox\Zihan\files\in\231101\231101__web__pf.tsv`
	pf := mustParseProductFile(pfPath)

	psM := lo.Associate(pf.Lines, func(l *report.ProductFileLine) (string, *report.ProductFileLine) {
		return l.Id, l
	})

	DB, err := gorm.Open(postgres.Open(connString), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}

	r := product.NewRepository(DB)
	s := product.NewService(r)

	pps, err := s.FetchProducts(
		product.WithBCIDs(lo.Keys(psM)...),
		product.WithStockInformation(),
	)
	if err != nil {
		log.Fatal(err)
	}

	pps = lo.Filter(pps, func(pp *presenter.Product, _ int) bool {
		return strings.HasPrefix(pp.ProdName, `\`)
	})

	for _, pp := range pps {
		l := psM[pp.BCID]
		if pp.IsVariant {
			ids := strings.Split(l.Id, "/")
			pIdStr, vIdStr := ids[0], ids[1]
			pId, err := strconv.Atoi(pIdStr)
			if err != nil {
				log.Println(err)
				continue
			}
			vId, err := strconv.Atoi(vIdStr)
			if err != nil {
				log.Println(err)
				continue
			}
			v, err := c.GetVariantById(vId, pId, map[string]string{})
			if err != nil {
				log.Println(err)
				continue
			}
			sku := strings.TrimSpace(v.Sku)
			if !strings.HasPrefix(sku, "/") {
				sku = "/" + sku
			}
			if pp.StockInfomation.Total == 0 {
				if !strings.HasPrefix(sku, "//") {
					sku = "/" + sku
				}
				_, err := c.UpdateVariant(v,
					bigc.WithUpdateVariantPurchasingDisabled(true),
					bigc.WithUpdateVariantInventoryLevel(0),
					bigc.WithUpdateVariantSalePrice(0),
					bigc.WithUpdateVariantSku(sku),
				)
				if err != nil {
					log.Println(err)
				}
				continue
			} else {
				if strings.HasPrefix(sku, "//") {
					sku = strings.TrimPrefix(sku, "/")
				}
				_, err := c.UpdateVariant(v,
					bigc.WithUpdateVariantInventoryLevel(pp.StockInfomation.Total),
					bigc.WithUpdateVariantPurchasingDisabled(false),
					bigc.WithUpdateVariantSku(sku),
				)
				if err != nil {
					log.Println(err)
				}
				continue
			}
		} else {
			id, err := strconv.Atoi(pp.BCID)
			if err != nil {
				log.Println(err)
				continue
			}

			p, err := c.GetProductById(id)
			if err != nil {
				log.Println(err)
				continue
			}
			sku := strings.TrimSpace(p.Sku)
			if !strings.HasPrefix(sku, "/") {
				sku = "/" + sku
			}

			if pp.StockInfomation.Total == 0 {
				if !strings.HasPrefix(sku, "//") {
					sku = "/" + sku
				}
				_, err := c.UpdateProduct(p,
					bigc.WithUpdateProductIsVisible(false),
					bigc.WithUpdateProductInventoryLevel(0),
					bigc.WithUpdateProductSalePrice(0),
					bigc.WithUpdateProductSku(sku),
					bigc.WithUpdateProductCategoryIsRetired(true),
					bigc.WithUpdateProductCategoriesWithoutSaleIDs(p.Categories),
				)
				if err != nil {
					log.Println(err)
				}
				continue
			} else {
				if strings.HasPrefix(sku, "//") {
					sku = strings.TrimPrefix(sku, "/")
				}
				_, err := c.UpdateProduct(p,
					bigc.WithUpdateProductInventoryLevel(pp.StockInfomation.Total),
					bigc.WithUpdateProductSku(sku),
					bigc.WithUpdateProductIsVisible(true),
					bigc.WithUpdateProductCategoryIsRetired(false),
				)
				if err != nil {
					log.Println(err)
					continue
				}
			}
		}
	}
}

func mustParseProductFile(path string) *report.ProductFile {
	var (
		f   *os.File
		err error
	)
	if f, err = os.Open(path); err != nil {
		log.Fatal(err)
	}

	var p parser.Parser
	if p, err = parser.NewCsvParser(f); err != nil {
		log.Fatal(err)
	}

	storeRegex := `(\d{6})__(.*?)__(.*?).(.*?)`
	re := regexp.MustCompile(storeRegex)

	var date time.Time
	if date, err = time.Parse("060102", re.FindStringSubmatch(path)[1]); err != nil {
		log.Fatal(err)
	}

	var pf = report.ProductFile{
		Report: report.Report{
			Date:   date,
			Source: path,
			Store:  "web",
		},
		Lines: []*report.ProductFileLine{},
	}

	if err = p.Parse(&pf.Lines); err != nil {
		log.Fatal(err)
	}

	return &pf
}
