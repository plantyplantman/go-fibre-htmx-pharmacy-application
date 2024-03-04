package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/plantyplantman/bcapi/api/presenter"
	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/plantyplantman/bcapi/pkg/parser"
	"github.com/plantyplantman/bcapi/pkg/product"
	"github.com/plantyplantman/bcapi/pkg/report"
	"github.com/samber/lo"
)

func main() {
	var ZihanFilesPath = `C:\Users\admin\Develin Management Dropbox\Zihan\files\`
	var inPath = filepath.Join(ZihanFilesPath, `in\`, `231211\`, `231211__petrie__prlwgp.TXT`)
	f, err := os.Open(inPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	pars, err := parser.NewCsvParser(f)
	if err != nil {
		panic(err)
	}

	var prlwgp = report.ProductRetailList{}
	if err = pars.Parse(&prlwgp.Lines); err != nil {
		panic(err)
	}

	skus := lo.FilterMap(prlwgp.Lines, func(l *report.ProductRetailListLine, _ int) (string, bool) {
		return l.Sku, l.ProdName[0] != '#' && l.ProdName[0] != '!' && l.ProdName[0] != '~'
	})

	service, err := product.NewDefaultService()
	if err != nil {
		panic(err)
	}
	batchSize := 50000
	var ps []*presenter.Product
	for i := 0; i < len(skus); i += batchSize {
		if i+batchSize > len(skus) {
			batchSize = len(skus) - i
		}
		var batch = skus[i : i+batchSize]
		tmp, err := service.FetchProducts(product.WithSkus(batch...), product.WithStockInformation())
		if err != nil {
			panic(err)
		}
		ps = append(ps, tmp...)
	}

	ps = lo.Filter(ps, func(p *presenter.Product, _ int) bool {
		return p.BCID != ""
	})

	bc := bigc.MustGetClient()

	for _, p := range ps {
		ids := strings.Split(p.BCID, "/")
		if p.IsVariant && len(ids) == 2 {
			vid, err := strconv.Atoi(ids[1])
			if err != nil {
				log.Println(err)
				continue
			}
			pid, err := strconv.Atoi(ids[0])
			if err != nil {
				log.Println(err)
				continue
			}

			vp, err := bc.GetVariantById(vid, pid, map[string]string{})
			if err != nil {
				log.Println(err)
				continue
			}

			if _, err = bc.UpdateVariant(vp, bigc.WithUpdateVariantCostPrice(p.CostPrice)); err != nil {
				log.Println(err)
				continue
			}
		} else if !p.IsVariant && len(ids) == 1 && ids[0] != "" {
			pid, err := strconv.Atoi(ids[0])
			if err != nil {
				log.Println(err)
				continue
			}

			pp, err := bc.GetProductById(pid)
			if err != nil {
				log.Println(err)
				continue
			}

			if _, err = bc.UpdateProduct(pp, bigc.WithUpdateProductCostPrice(p.CostPrice)); err != nil {
				log.Println(err)
				continue
			}
		} else {
			log.Println("skipping", p.Sku)
			continue
		}
	}
}
