package main

import (
	"log"
	"os"
	"strings"

	"github.com/gocarina/gocsv"
	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/plantyplantman/bcapi/pkg/parser"
	"github.com/plantyplantman/bcapi/pkg/report"
	"github.com/samber/lo"
)

type Retv struct {
	Sku       string
	Name      string
	Price     float64
	SalePrice float64
}

func main() {
	path := `C:\Users\admin\Develin Management Dropbox\Zihan\files\in\240306\240306__petrie__promos.xml`
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	p := parser.NewXmlParser(b)
	r := report.Campaigns{}
	err = p.Parse(&r)
	if err != nil {
		log.Fatal(err)
	}

	var promoMap = map[string]report.Product{}
	for _, o := range r.Campaign.Offers.Offer {
		for _, p := range o.Products.Product {
			promoMap[strings.TrimSpace(p.EAN)] = p
		}
	}

	bc := bigc.MustGetClient()
	ps, err := bc.GetAllProducts(map[string]string{})
	if err != nil {
		log.Fatal(err)
	}

	ps = lo.Filter(ps, func(p bigc.Product, _ int) bool {
		_, ok := promoMap[strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(p.Sku), "/"), "/")]
		return ok && strings.TrimSpace(p.Sku) != ""
	})

	retv := make([]Retv, 0)
	for _, p := range ps {
		retv = append(retv, Retv{
			Sku:       p.Sku,
			Name:      p.Name,
			Price:     p.Price,
			SalePrice: p.SalePrice,
		})
	}

	f, err := os.Create(`C:\Users\admin\Develin Management Dropbox\Zihan\files\out\240313\240313__promos.csv`)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := gocsv.MarshalFile(&retv, f); err != nil {
		log.Fatal(err)
	}

	log.Println("done")
}
