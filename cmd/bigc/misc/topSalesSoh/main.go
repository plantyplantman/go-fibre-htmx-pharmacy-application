package main

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gocarina/gocsv"
	"github.com/plantyplantman/bcapi/api/presenter"
	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/plantyplantman/bcapi/pkg/parser"
	"github.com/plantyplantman/bcapi/pkg/product"
	"github.com/samber/lo"
)

type BigCommerceTopSalesReport struct {
	Lines []BigCommerceTopSalesReportLine
}
type BigCommerceTopSalesReportLine struct {
	Name         string  `csv:"name"`
	Sku          string  `csv:"product_code_or_sku"`
	Brand        string  `csv:"brand"`
	ProductCode  string  `csv:"product_code"`
	Orders       int     `csv:"orders"`
	Revenue      float64 `csv:"revenue"`
	QuantitySold int     `csv:"qty_sold"`
	Visits       int     `csv:"visits"`
}

type Retv []RetvLine
type RetvLine struct {
	Name         string
	Sku          string
	Brand        string
	Orders       int
	Revenue      float64
	QuantitySold int
	Visits       int

	presenter.StockInformation
}

func main() {
	fp := filepath.Join(`C:\Users\admin\Downloads\`, `Merchandising Products for 2023-10-01 - 2023-12-12.csv`)
	f, err := os.Open(fp)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	pars, err := parser.NewCsvParser(f, parser.WithComma(','))
	if err != nil {
		panic(err)
	}

	report := BigCommerceTopSalesReport{}
	if err := pars.Parse(&report.Lines); err != nil {
		panic(err)
	}

	skuM1 := lo.Associate(report.Lines, func(l BigCommerceTopSalesReportLine) (string, BigCommerceTopSalesReportLine) {
		return strings.TrimSpace(l.Sku), l
	})
	skuM := make(map[string]BigCommerceTopSalesReportLine)
	for k, v := range skuM1 {
		sku, err := bigc.CleanBigCommerceSku(k)
		if err != nil {
			log.Println(err)
			continue
		}
		skuM[sku] = v
	}

	service, err := product.NewDefaultService()
	if err != nil {
		panic(err)
	}

	ps, err := service.FetchProducts(product.WithSkus(lo.Keys(skuM)...), product.WithStockInformation())
	if err != nil {
		panic(err)
	}

	var retv Retv = lo.FilterMap(ps, func(p *presenter.Product, _ int) (RetvLine, bool) {
		bcp, ok := skuM[p.Sku]
		if !ok {
			return RetvLine{}, false
		}
		return RetvLine{
			StockInformation: p.StockInformation,
			Name:             p.ProdName,
			Sku:              p.Sku,
			Brand:            bcp.Brand,
			Orders:           bcp.Orders,
			Revenue:          bcp.Revenue,
			QuantitySold:     bcp.QuantitySold,
			Visits:           bcp.Visits,
		}, true
	})

	sort.Slice(retv, func(i, j int) bool {
		return retv[i].QuantitySold > retv[j].QuantitySold
	})

	f, err = os.Create("topSaleSoh.csv")
	if err != nil {
		panic(err)
	}

	if err = gocsv.MarshalFile(retv, f); err != nil {
		panic(err)
	}

	log.Println("Successfully exported to topSaleSoh.csv")
}
