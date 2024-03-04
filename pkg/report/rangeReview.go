package report

import (
	"io"
	"log"
	"slices"
	"strconv"
	"strings"

	"github.com/gocarina/gocsv"
	"github.com/plantyplantman/bcapi/api/presenter"
	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/samber/lo"
)

type RangeReviewReport []RangeReviewLine
type RangeReviewLine struct {
	Sku             string  `csv:"sku" json:"sku"`
	Name            string  `csv:"name" json:"name"`
	OnWeb           bool    `csv:"on_web" json:"on_web"`
	Price           float64 `csv:"price" json:"price"`
	SalePrice       float64 `csv:"sale_price" json:"sale_price"`
	SalePercent     float64 `csv:"sale_percent" json:"sale_percent"`
	IsDeletedMinfos bool    `csv:"is_deleted_minfos" json:"is_deleted_minfos"`
	IsDeletedWeb    bool    `csv:"is_deleted" json:"is_deleted"`
	IsVisible       bool    `csv:"is_visible" json:"is_visible"`
	IsRetired       bool    `csv:"is_retired" json:"is_retired"`
	NumTiggles      int     `csv:"num_tiggles" json:"num_tiggles"`
	BCID            string  `csv:"bcid" json:"bcid"`
	Petrie          int     `csv:"petrie" json:"petrie"`
	Bunda           int     `csv:"bunda" json:"bunda"`
	Con             int     `csv:"con" json:"con"`
	Franklin        int     `csv:"franklin" json:"franklin"`
	Web             int     `csv:"web" json:"web"`
	Total           int     `csv:"total" json:"total"`
}

func NewRangeReview(products []*presenter.Product) (RangeReviewReport, error) {

	c := bigc.MustGetClient()
	var report RangeReviewReport = lo.FilterMap(products, func(p *presenter.Product, _ int) (RangeReviewLine, bool) {
		var bp = &bigc.Product{}
		var deleted = false
		if p.BCID != "" {
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

		salePercent := 0.0
		if p.Price > 0 {
			salePercent = (p.Price - bp.SalePrice) / p.Price
		}

		return RangeReviewLine{
			Sku:             p.Sku,
			Name:            p.ProdName,
			Price:           p.Price,
			Petrie:          p.StockInformation.Petrie,
			Bunda:           p.StockInformation.Bunda,
			Con:             p.StockInformation.Con,
			Franklin:        p.StockInformation.Franklin,
			Web:             p.StockInformation.Web,
			Total:           p.StockInformation.Total,
			SalePrice:       bp.SalePrice,
			SalePercent:     salePercent,
			IsDeletedMinfos: strings.HasPrefix(p.ProdName, "\\"),
			IsDeletedWeb:    deleted,
			IsVisible:       bp.IsVisible,
			IsRetired:       slices.Contains(bp.Categories, bigc.RETIRED_PRODUCTS),
			NumTiggles:      numTiggles(bp.Sku),
			BCID:            p.BCID,
			OnWeb:           p.BCID != "",
		}, true
	})

	return report, nil
}

func (report *RangeReviewReport) Export(w io.Writer) error {
	if err := gocsv.Marshal(report, w); err != nil {
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
