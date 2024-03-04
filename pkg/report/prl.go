package report

import (
	"log"
	"strconv"
	"strings"

	"github.com/plantyplantman/bcapi/api/presenter"
	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/plantyplantman/bcapi/pkg/product"
	"github.com/samber/lo"
)

type ProductRetailList struct {
	Report
	Lines []*ProductRetailListLine
}

type ProductRetailListLine struct {
	Mnpn     int    `csv:"MNPN"`
	Sku      string `csv:"Barcode"`
	ProdNo   int    `csv:"Product No."`
	ProdName string `csv:"Product Name"`
	Price    Price  `csv:"Retail, Retail Price"`
	Cost     Price  `csv:"Cost"`
}

func (prl *ProductRetailList) ToTable() presenter.Table {
	headers := []string{
		"MNPN",
		"Barcode",
		"Product No.",
		"Product Name",
		"Price",
	}
	rows := make([]presenter.Row, 0)
	for _, l := range prl.Lines {
		rows = append(rows, l.ToPresenterRow())
	}

	return presenter.Table{
		Headers: headers,
		Rows:    rows,
	}
}

func (prl *ProductRetailListLine) ToPresenterRow(cns ...string) presenter.Row {
	return presenter.Row{
		Cells: []string{
			strconv.Itoa(prl.Mnpn),
			prl.Sku,
			strconv.Itoa(prl.ProdNo),
			prl.ProdName,
			prl.Price.String(),
		},
		ClassName: strings.Join(cns, " "),
	}
}

func (prl *ProductRetailList) ToSkuMap() map[string]*ProductRetailListLine {
	m := make(map[string]*ProductRetailListLine, len(prl.Lines))
	for _, line := range prl.Lines {
		m[strings.TrimSpace(line.Sku)] = line
	}
	return m
}

func (prl *ProductRetailList) NotOnSite(r product.Service) (NotOnSiteReport, error) {
	skuM := prl.ToSkuMap()
	ps, err := r.FetchProducts(product.WithSkus(lo.Keys(skuM)...), product.WithStockInformation())
	if err != nil {
		return nil, err
	}

	var retv NotOnSiteReport = lo.FilterMap(ps, func(p *presenter.Product, _ int) (NotOnSiteReportLine, bool) {
		return NotOnSiteReportLine{
			Sku:              p.Sku,
			ProdName:         p.ProdName,
			Price:            p.Price,
			StockInformation: p.StockInformation,
			Source:           prl.Source,
			Date:             DateTime{prl.Date},
		}, p.StockInformation.Web == 0
	})

	return retv, nil
}

func (prl *ProductRetailList) Edited(c *bigc.Client, r product.Service) (NotOnSiteReport, error) {
	notOnSiteReport := NotOnSiteReport{}
	skuM := prl.ToSkuMap()
	ps, err := r.FetchProducts(product.WithSkus(lo.Keys(skuM)...), product.WithStockInformation())
	if err != nil {
		return nil, err
	}
	pM := lo.Associate(ps, func(p *presenter.Product) (string, *presenter.Product) {
		return p.Sku, p
	})

	for _, l := range prl.Lines {
		p, ok := pM[l.Sku]
		if !ok {
			log.Println(err)
			continue
		}

		if p.OnWeb == 0 {
			notOnSiteReportLine := NotOnSiteReportLine{
				Sku:              l.Sku,
				ProdName:         l.ProdName,
				StockInformation: p.StockInformation,
				Price:            l.Price.float64,
				Source:           prl.Source,
				Date:             DateTime{prl.Date},
			}
			notOnSiteReport = append(notOnSiteReport, notOnSiteReportLine)
			continue
		}

		if p.IsVariant {
			ids := strings.Split(p.BCID, "/")
			if len(ids) < 2 {
				log.Println("Invalid BCID for variant", p.BCID)
				continue
			}
			pId, err := strconv.Atoi(ids[0])
			if err != nil {
				log.Println(err)
				continue
			}
			vId, err := strconv.Atoi(ids[1])
			if err != nil {
				log.Println(err)
				continue
			}
			original, err := c.GetVariantById(vId, pId, map[string]string{})
			if err != nil {
				log.Println(err)
				continue
			}

			if _, err := c.UpdateVariant(original,
				bigc.WithUpdateVariantPrice(l.Price.float64),
				bigc.WithUpdateVariantCostPrice(p.CostPrice),
				bigc.WithUpdateVariantInventoryLevel(p.StockInformation.Total),
				bigc.WithUpdateVariantRetailPrice(0.0),
			); err != nil {
				log.Println(err)
				continue
			}
			continue
		}

		bcid, err := strconv.Atoi(p.BCID)
		if err != nil {
			log.Println(err)
			continue
		}

		original, err := c.GetProductById(bcid)
		if err != nil {
			log.Println(err)
			continue
		}
		if _, err := c.UpdateProduct(original,
			bigc.WithUpdateProductPrice(l.Price.float64),
			bigc.WithUpdateProductCostPrice(p.CostPrice),
			bigc.WithUpdateProductInventoryLevel(p.StockInformation.Total),
		); err != nil {
			log.Println(err)
			continue
		}
	}

	return notOnSiteReport, nil
}

func (prl *ProductRetailList) Deleted(c *bigc.Client, r product.Service) (NotOnSiteReport, error) {
	notOnSiteReport := NotOnSiteReport{}
	skuM := prl.ToSkuMap()
	ps, err := r.FetchProducts(product.WithSkus(lo.Keys(skuM)...), product.WithStockInformation())
	if err != nil {
		return nil, err
	}
	psM := lo.Associate(ps, func(p *presenter.Product) (string, *presenter.Product) {
		return p.Sku, p
	})

	for _, l := range prl.Lines {
		p, ok := psM[l.Sku]
		if !ok {
			log.Printf("product not found in sku map. sku: %s", l.Sku)
			continue
		}

		if p.OnWeb == 0 {
			notOnSiteReport = append(notOnSiteReport, NotOnSiteReportLine{
				Sku:              l.Sku,
				ProdName:         l.ProdName,
				Price:            l.Price.float64,
				StockInformation: p.StockInformation,
				Source:           prl.Source,
				Action:           "",
				Date:             DateTime{prl.Date},
			})
			continue
		}

		if p.IsVariant {
			ids := strings.Split(p.BCID, "/")
			if len(ids) < 2 {
				log.Println("Invalid BCID for variant", p.BCID)
				continue
			}
			pId, err := strconv.Atoi(ids[0])
			if err != nil {
				log.Println(err)
				continue
			}
			vId, err := strconv.Atoi(ids[1])
			if err != nil {
				log.Println(err)
				continue
			}
			original, err := c.GetVariantById(vId, pId, map[string]string{})
			if err != nil {
				log.Println(err)
				continue
			}

			var newSku string
			if !strings.HasPrefix(original.Sku, "/") {
				newSku = "/" + l.Sku
			} else {
				newSku = original.Sku
			}

			if p.StockInformation.Total == 0 {
				if !strings.HasPrefix(newSku, "//") {
					newSku = "/" + newSku
				}
				if _, err := c.UpdateVariant(original,
					bigc.WithUpdateVariantPrice(l.Price.float64),
					bigc.WithUpdateVariantCostPrice(p.CostPrice),
					bigc.WithUpdateVariantInventoryLevel(p.StockInformation.Total),
					bigc.WithUpdateVariantRetailPrice(0.0),
					bigc.WithUpdateVariantSalePrice(0.0),
					bigc.WithUpdateVariantPurchasingDisabled(true),
					bigc.WithUpdateVariantSku(newSku),
				); err != nil {
					log.Println(err)
					continue
				}
			} else {
				if _, err := c.UpdateVariant(original,
					bigc.WithUpdateVariantPrice(l.Price.float64),
					bigc.WithUpdateVariantCostPrice(p.CostPrice),
					bigc.WithUpdateVariantInventoryLevel(p.StockInformation.Total),
					bigc.WithUpdateVariantRetailPrice(0.0),
					bigc.WithUpdateVariantSalePrice(0.0),
					bigc.WithUpdateVariantSku(newSku),
				); err != nil {
					log.Println(err)
					continue
				}
			}
			continue
		}

		bcid, err := strconv.Atoi(p.BCID)
		if err != nil {
			log.Println(err)
			continue
		}
		original, err := c.GetProductById(bcid)
		if err != nil {
			log.Println(err)
			continue
		}

		soh := p.StockInformation.Total

		var newSku string
		if !strings.HasPrefix(original.Sku, "/") {
			newSku = "/" + l.Sku
		} else {
			newSku = original.Sku
		}
		if p.StockInformation.Total == 0 {
			if !strings.HasPrefix(newSku, "//") {
				newSku = "/" + newSku
			}
		}

		if p.StockInformation.Total == 0 {
			if _, err := c.UpdateProduct(
				original,
				bigc.WithUpdateProductSku(newSku),
				bigc.WithUpdateProductInventoryLevel(soh),
				bigc.WithUpdateProductCategoriesWithoutSaleIDs(original.Categories),
				bigc.WithUpdateProductIsVisible(false),
				bigc.WithUpdateProductCategoryIsRetired(true),
			); err != nil {
				log.Println(err)
			}
			continue
		}

		if _, err := c.UpdateProduct(
			original,
			bigc.WithUpdateProductSku(newSku),
			bigc.WithUpdateProductCategoriesWithoutSaleIDs(original.Categories),
			bigc.WithUpdateProductPrice(p.Price),
			bigc.WithUpdateProductCostPrice(p.CostPrice),
		); err != nil {
			log.Println(err)
			continue
		}
	}
	return notOnSiteReport, nil
}

func DoMultistore(prlM map[string]*ProductRetailList, s product.Service, c *bigc.Client) NotOnSiteReport {
	var notOnSiteReport NotOnSiteReport
	for k := range prlM {
		switch k {
		case "new":
			nosr, err := prlM[k].NotOnSite(s)
			if err != nil {
				log.Println(err)
				continue
			}
			notOnSiteReport = append(notOnSiteReport, nosr...)
		case "edited":
			nosr, err := prlM[k].Edited(c, s)
			if err != nil {
				log.Println(err)
				continue
			}
			notOnSiteReport = append(notOnSiteReport, nosr...)
		case "clean":
			nosr, err := prlM[k].Deleted(c, s)
			if err != nil {
				log.Println(err)
				continue
			}
			notOnSiteReport = append(notOnSiteReport, nosr...)
		default:
			continue
		}
	}

	return notOnSiteReport
}
