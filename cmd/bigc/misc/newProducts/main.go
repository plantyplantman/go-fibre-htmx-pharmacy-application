package main

import (
	"fmt"
	"os"

	"github.com/gocarina/gocsv"
	"github.com/plantyplantman/bcapi/api/presenter"
	"github.com/plantyplantman/bcapi/pkg/product"
	"github.com/plantyplantman/bcapi/pkg/report"
	"github.com/samber/lo"
)

func main() {
	files := []string{
		`C:\\Users\\admin\\Develin Management Dropbox\\Zihan\\files\\out\\231204\\231204__MS-Web__not-on-site.csv`,
		`C:\\Users\\admin\\Develin Management Dropbox\\Zihan\\files\\out\\240109\\240109__MS-Web__not-on-site.csv`,
		`C:\\Users\\admin\\Develin Management Dropbox\\Zihan\\files\\out\\240115\\240115__MS-Web__not-on-site.csv`,
		`C:\\Users\\admin\\Develin Management Dropbox\\Zihan\\files\\out\\240116\\240116__MS-Web__not-on-site.csv`,
		`C:\\Users\\admin\\Develin Management Dropbox\\Zihan\\files\\out\\240130\\240130__MS-Web__not-on-site.csv`,
		`C:\\Users\\admin\\Develin Management Dropbox\\Zihan\\files\\out\\240202\\240202__MS-Web__not-on-site.csv`,
		`C:\\Users\\admin\\Develin Management Dropbox\\Zihan\\files\\out\\240209\\240209__MS-Web__not-on-site.csv`,
		`C:\\Users\\admin\\Develin Management Dropbox\\Zihan\\files\\out\\240210\\240210__MS-Web__not-on-site.csv`,
		`C:\\Users\\admin\\Develin Management Dropbox\\Zihan\\files\\out\\240215\\240215__MS-Web__not-on-site.csv`,
		`C:\\Users\\admin\\Develin Management Dropbox\\Zihan\\files\\out\\240219\\240219__MS-Web__not-on-site.csv`,
		`C:\\Users\\admin\\Develin Management Dropbox\\Zihan\\files\\out\\240223\\240223__MS-Web__not-on-site.csv`,
	}

	nosrs := report.NotOnSiteReport{}
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			panic(err)
		}

		tmp := report.NotOnSiteReport{}
		if err := gocsv.UnmarshalFile(f, &tmp); err != nil {
			panic(err)
		}

		nosrs = append(nosrs, tmp...)
	}

	skus := lo.Map(nosrs, func(r report.NotOnSiteReportLine, _ int) string {
		return r.Sku
	})

	s, err := product.NewDefaultService()
	if err != nil {
		panic(err)
	}

	ps, err := s.FetchProducts(product.WithSkus(skus...), product.WithStockInformation())
	if err != nil {
		panic(err)
	}

	for _, p := range ps {
		if p.OnWeb == 0 {
			fmt.Println(p.Sku, p.ProdName, p.StockInformation.Total)
		}
	}

	retv := lo.FilterMap(ps, func(p *presenter.Product, _ int) (report.NotOnSiteReportLine, bool) {
		return report.NotOnSiteReportLine{
			Sku:              p.Sku,
			ProdName:         p.ProdName,
			Price:            p.Price,
			Soh:              p.StockInformation.Total,
			StockInformation: p.StockInformation,
		}, p.OnWeb == 0
	})

	op := `C:\Users\admin\Develin Management Dropbox\Zihan\files\out\240228\231204-240223__MS-Web__not-on-site.csv`
	f, err := os.Create(op)
	if err != nil {
		panic(err)
	}
	if err := gocsv.MarshalFile(retv, f); err != nil {
		panic(err)
	}
	fmt.Println("Exported to ", op)
}
