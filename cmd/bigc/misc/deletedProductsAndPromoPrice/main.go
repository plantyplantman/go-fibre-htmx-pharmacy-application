package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gocarina/gocsv"
	"github.com/plantyplantman/bcapi/api/presenter"
	"github.com/plantyplantman/bcapi/pkg/parser"
	"github.com/plantyplantman/bcapi/pkg/product"
	"github.com/plantyplantman/bcapi/pkg/report"
	"github.com/samber/lo"
)

func main() {

	promoFilePath := `C:\Users\admin\Develin Management Dropbox\Zihan\files\in\231211\231211__petrie__complete-promos.xml`
	b, err := os.ReadFile(promoFilePath)
	if err != nil {
		log.Fatalln(err)
	}

	pars := parser.NewXmlParser(b)
	var promos = report.Campaigns{}
	if err = pars.Parse(&promos); err != nil {
		log.Fatalln(err)
	}
	fmt.Println(promos.Campaign.Offers.Offer[0].Products.Product[0].EAN)

	var deletedSkus = make(map[string]string)
	for _, o := range promos.Campaign.Offers.Offer {
		for _, p := range o.Products.Product {
			if strings.HasPrefix(p.ProductName, "\\") {
				if o.DollarOffDisc != "0.00" {
					deletedSkus[p.EAN] = "$" + o.DollarOffDisc
				} else if o.PercentOffDisc != "0.00" {
					deletedSkus[p.EAN] = "%" + o.PercentOffDisc
				} else if p.OfferPrice != 0.0 {
					deletedSkus[p.EAN] = strconv.FormatFloat(p.OfferPrice, 'f', 2, 64)
				}
			}
		}
	}

	type ReportLine struct {
		Sku        string
		Name       string
		Price      string
		PromoPrice string
		OnWeb      bool
		presenter.StockInformation
	}
	type Report []ReportLine

	service, err := product.NewDefaultService()
	if err != nil {
		log.Fatalln(err)
	}

	ps, err := service.FetchProducts(product.WithSkus(lo.Keys(deletedSkus)...), product.WithStockInformation())
	if err != nil {
		log.Fatalln(err)
	}

	petStsPath := `C:\Users\admin\Develin Management Dropbox\Zihan\files\in\231211\231211__petrie__sts.TXT`
	f, err := os.Open(petStsPath)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	pars, err = parser.NewCsvParser(f)
	if err != nil {
		log.Fatalln(err)
	}
	sts := report.ProductStockList{}
	err = pars.Parse(&sts.Lines)
	if err != nil {
		log.Fatalln(err)
	}
	skus := lo.FilterMap(sts.Lines, func(l *report.ProductStockListLine, _ int) (string, bool) {
		return l.Sku, !strings.HasPrefix(l.Sku, "\\")
	})

	for _, s := range skus {
		if _, ok := deletedSkus[s]; !ok {
			deletedSkus[s] = "!"
		}
	}

	var report Report = lo.Map(ps, func(p *presenter.Product, _ int) ReportLine {
		disc, ok := deletedSkus[p.Sku]
		if !ok {
			log.Fatalln("sku not found in deletedSkus", p.Sku)
		}
		var promoPrice float64
		switch disc[0] {
		case '$':
			dollarOff, err := strconv.ParseFloat(disc[1:], 64)
			if err != nil {
				log.Fatalln(err)
			}
			promoPrice = p.Price - dollarOff
		case '%':
			percentOff, err := strconv.ParseFloat(disc[1:], 64)
			if err != nil {
				log.Fatalln(err)
			}
			promoPrice = p.Price - (p.Price * percentOff)
		case '!':
			promoPrice = 0.0
		default:
			promoPrice, err = strconv.ParseFloat(disc, 64)
			if err != nil {
				log.Fatalln(err)
			}
		}

		onWeb := false
		if p.OnWeb == 1 {
			onWeb = true
		}

		return ReportLine{
			Sku:              p.Sku,
			Name:             p.ProdName,
			Price:            strconv.FormatFloat(p.Price, 'f', 2, 64),
			PromoPrice:       strconv.FormatFloat(promoPrice, 'f', 2, 64),
			OnWeb:            onWeb,
			StockInformation: p.StockInformation,
		}
	})

	outFilePath := `deletedProductsAndPromos.csv`
	file, err := os.Create(outFilePath)
	if err != nil {
		log.Fatalln(err)
	}
	err = gocsv.MarshalFile(&report, file)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Exported to ", outFilePath)
}

// func delimit(d string, s ...any) string {
// 	return strings.Join(lo.Map(s, func(a any, _ int) string {
// 		return fmt.Sprintf("%v", a)
// 	}), d)
// }
