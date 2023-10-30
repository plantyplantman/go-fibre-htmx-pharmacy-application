package report

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/plantyplantman/bcapi/api/presenter"
	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/plantyplantman/bcapi/pkg/product"
)

type NotOnSiteReport []NotOnSiteReportLine

type NotOnSiteReportLine struct {
	Sku      string   `csv:"Sku"`
	ProdName string   `csv:"ProdName"`
	Price    float64  `csv:"Price"`
	Source   string   `csv:"Source"`
	Action   string   `csv:"Action"`
	Soh      int      `csv:"SOH"`
	Weight   float64  `csv:"Weight"`
	Width    float64  `csv:"Width"`
	Height   float64  `csv:"Height"`
	Depth    float64  `csv:"Depth"`
	Date     DateTime `csv:"Date"`
}

func (r NotOnSiteReport) Action(s product.Service, c *bigc.BigCommerceClient, ai_c *bigc.AI_Client) {
	for _, l := range r {
		if strings.ToLower(strings.TrimSpace(l.Action)) == "add" {
			p := presenter.Product{Sku: l.Sku}
			if err := s.FetchProduct(&p); err != nil {
				log.Println(err)
			}
			var (
				soh       int
				costPrice float64
			)
			soh = p.StockInfomation.Total
			costPrice = p.CostPrice
			if l.Soh != 0 {
				soh = l.Soh
			}
			req := bigc.NewProduct{
				Name:           l.ProdName,
				Sku:            l.Sku,
				Weight:         l.Weight,
				Width:          l.Width,
				Depth:          l.Depth,
				Height:         l.Height,
				Price:          l.Price,
				CostPrice:      costPrice,
				InventoryLevel: soh,
			}

			_, err := req.Fill(ai_c)
			if err != nil {
				log.Println(err)
				continue
			}
			_, err = c.CreateProduct(req)
			if err != nil {
				log.Println(err)
				continue
			}
			fmt.Println("created", l.Sku)
		}
	}
}

type DateTime struct {
	time.Time
}

func (dt *DateTime) UnmarshalCSV(csv string) (err error) {
	return nil
}

func (dt DateTime) MarshalCSV() (string, error) {
	return dt.Format("2006-01-02"), nil
}
