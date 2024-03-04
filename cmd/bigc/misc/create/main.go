package main

import (
	"log"
	"os"
	"strings"

	"github.com/plantyplantman/bcapi/api/presenter"
	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/plantyplantman/bcapi/pkg/parser"
	"github.com/plantyplantman/bcapi/pkg/product"
	"github.com/plantyplantman/bcapi/pkg/report"
	"github.com/samber/lo"
)

func main() {
	service, err := product.NewDefaultService()
	if err != nil {
		log.Fatalln(err)
	}

	c := bigc.MustGetClient()
	ai_c, err := bigc.GetOpenAiClient()
	if err != nil {
		log.Fatal(err)
	}

	path := `C:\Users\admin\Desktop\231120__petrie__archline-sts.TXT`
	sts, err := parseSts(path)
	if err != nil {
		log.Fatalln(err)
	}

	skus := lo.Map(sts.Lines, func(l *report.ProductStockListLine, _ int) string {
		return strings.TrimSpace(l.Sku)
	})

	for _, sku := range skus {
		create(sku, &ai_c, c, service)
	}

}

func parseSts(path string) (*report.ProductStockList, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	p, err := parser.NewCsvParser(f)
	if err != nil {
		return nil, err
	}

	psl := report.ProductStockList{
		Lines: []*report.ProductStockListLine{},
	}
	if err := p.Parse(&psl.Lines); err != nil {
		return nil, err
	}

	return &psl, nil
}

func create(sku string, ai_c *bigc.AI_Client, c *bigc.Client, s product.Service) error {
	p := presenter.Product{Sku: sku}
	if err := s.FetchProduct(&p); err != nil {
		log.Println(err)
	}
	req := bigc.NewProduct{
		Name:           p.ProdName,
		Sku:            p.Sku,
		Weight:         0,
		Width:          0,
		Depth:          0,
		Height:         0,
		Price:          p.Price,
		CostPrice:      p.CostPrice,
		InventoryLevel: p.StockInformation.Total,
	}

	_, err := req.Fill(ai_c)
	if err != nil {
		return err
	}
	_, err = c.CreateProduct(req)
	if err != nil {
		return err
	}

	return nil
}
