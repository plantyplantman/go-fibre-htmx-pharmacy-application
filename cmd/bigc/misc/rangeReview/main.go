package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/plantyplantman/bcapi/api/presenter"
	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/plantyplantman/bcapi/pkg/parser"
	"github.com/plantyplantman/bcapi/pkg/product"
	"github.com/plantyplantman/bcapi/pkg/report"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type RangeReviewReport []RangeReviewLine
type RangeReviewLine struct {
	Sku         string  `csv:"sku"`
	Name        string  `csv:"name"`
	OnWeb       bool    `csv:"on_web"`
	SohWeb      int     `csv:"soh_web"`
	Price       float64 `csv:"price"`
	SalePrice   float64 `csv:"sale_price"`
	SalePercent float64 `csv:"sale_percent"`
	presenter.StockInformation

	IsDeletedMinfos bool   `csv:"is_deleted_minfos"`
	IsDeletedWeb    bool   `csv:"is_deleted"`
	IsVisible       bool   `csv:"is_visible"`
	IsRetired       bool   `csv:"is_retired"`
	NumTiggles      int    `csv:"num_tiggles"`
	BCID            string `csv:"bcid"`
}

func main() {
	brands := []string{"clean logic", "nizoral"}
	date := time.Now().Format("060102")
	for _, brand := range brands {
		path := filepath.Join(`C:\Users\admin\Develin Management Dropbox\Zihan\files\out\`, date, date+`__range-review__`+brand+`.csv`)
		if err := run(brand, path); err != nil {
			log.Println(err)
			continue
		}
	}
	// if err := runByXml(`C:\\Users\\admin\\Desktop\\xmasconfec.xml`, `C:\\Users\\admin\\Desktop\\xmasconfec.csv`); err != nil {
	// 	log.Fatal(err)
	// }
}

func run(brand string, outPath string) error {
	ps, err := fetchProductsByBrand(brand)
	if err != nil {
		return err
	}

	report, err := newRangeReview(ps)
	if err != nil {
		return err
	}

	if err := export(outPath, report); err != nil {
		return err
	}

	return nil
}

func runByXml(xmlPath string, outPath string) error {
	file, err := os.Open(xmlPath)
	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	p := parser.NewXmlParser(bytes)

	r := report.Campaigns{}
	if err = p.Parse(&r); err != nil {
		return err
	}

	var skus = make([]string, 0)
	for _, o := range r.Campaign.Offers.Offer {
		for _, p := range o.Products.Product {
			skus = append(skus, p.EAN)
		}
	}

	service, err := product.NewDefaultService()
	if err != nil {
		return err
	}

	ps, err := service.FetchProducts(product.WithSkus(skus...), product.WithStockInformation())
	if err != nil {
		return err
	}

	report, err := newRangeReview(ps)
	if err != nil {
		return err
	}

	return export(outPath, report)
}

func fetchProductsByBrand(brand string) ([]*presenter.Product, error) {
	service, err := product.NewDefaultService()
	if err != nil {
		return nil, err
	}

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

func newRangeReview(products []*presenter.Product) (RangeReviewReport, error) {

	c := bigc.MustGetClient()
	var report RangeReviewReport = lo.FilterMap(products, func(p *presenter.Product, _ int) (RangeReviewLine, bool) {
		var bp = &bigc.Product{}
		var deleted = false
		if p.BCID != "" {
			// Variants
			if strings.Contains(p.BCID, "/") {
				ids := strings.Split(p.BCID, "/")
				spid := ids[0]
				pid, err := strconv.Atoi(spid)
				if err != nil {
					log.Println(err)
					return RangeReviewLine{}, false
				}

				svid := ids[1]
				vid, err := strconv.Atoi(svid)
				if err != nil {
					log.Println(err)
					return RangeReviewLine{}, false
				}

				vbp, err := c.GetVariantById(vid, pid, map[string]string{})
				if err != nil {
					log.Println(err)
					return RangeReviewLine{}, false
				}

				bp, err = c.GetProductById(pid)
				if err != nil {
					log.Println(err)
					return RangeReviewLine{}, false
				}

				return RangeReviewLine{
					Sku:              vbp.Sku,
					Name:             fmt.Sprintf("%s/%s", bp.Name, vbp.OptionValues[0].OptionDisplayName),
					SohWeb:           vbp.InventoryLevel,
					StockInformation: p.StockInformation,
					Price:            p.Price,
					SalePrice:        vbp.SalePrice,
					SalePercent:      p.Price - vbp.SalePrice/p.Price,
					IsDeletedMinfos:  strings.HasPrefix(p.ProdName, "\\"),
					IsDeletedWeb:     deleted,
					IsVisible:        bp.IsVisible,
					IsRetired:        slices.Contains(bp.Categories, bigc.RETIRED_PRODUCTS),
					NumTiggles:       numTiggles(bp.Sku),
					BCID:             p.BCID,
					OnWeb:            p.BCID != "",
				}, true
			}

			id, err := strconv.Atoi(p.BCID)
			if err != nil {
				log.Println(err)
				return RangeReviewLine{}, false
			}

			bp, err = c.GetProductById(id)
			if err != nil {
				log.Println(err)
				return RangeReviewLine{}, false
			}

			deleted = maybeDeleted(bp)
		}

		return RangeReviewLine{
			Sku:              p.Sku,
			Name:             p.ProdName,
			SohWeb:           p.StockInformation.Web,
			StockInformation: p.StockInformation,
			Price:            p.Price,
			SalePrice:        bp.SalePrice,
			SalePercent:      p.Price - bp.SalePrice/p.Price,
			IsDeletedMinfos:  strings.HasPrefix(p.ProdName, "\\"),
			IsDeletedWeb:     deleted,
			IsVisible:        bp.IsVisible,
			IsRetired:        slices.Contains(bp.Categories, bigc.RETIRED_PRODUCTS),
			NumTiggles:       numTiggles(bp.Sku),
			BCID:             p.BCID,
			OnWeb:            p.BCID != "",
		}, true
	})

	return report, nil
}

func export(path string, report RangeReviewReport) error {
	if err := os.MkdirAll(filepath.Dir(path), 0770); err != nil {
		log.Fatalln(err)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	err = gocsv.Marshal(report, f)
	if err != nil {
		return err
	}

	return nil
}

func maybeDeleted(p *bigc.Product) bool {
	if p == nil {
		return false
	}
	return strings.HasPrefix(p.Sku, "//") && !p.IsVisible && slices.Contains(p.Categories, bigc.RETIRED_PRODUCTS)
}

func numTiggles(sku string) int {
	if strings.HasPrefix(sku, "//") {
		return 2
	}
	if strings.HasPrefix(sku, "/") {
		return 1
	}
	return 0
}

func isDeletedMinfos(name string) bool {
	c := name[0]
	return c == '\\' || c == '#' || c == '!'
}
