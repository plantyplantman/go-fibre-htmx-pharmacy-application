package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/gocarina/gocsv"
	"github.com/plantyplantman/bcapi/api/presenter"
	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/plantyplantman/bcapi/pkg/report"
	"github.com/samber/lo"
)

type Retv struct {
	*presenter.Product
	Categories []string
}

func main() {
	pMap, err := getProductMap(`C:\Users\admin\Develin Management Dropbox\Zihan\files\out\240304\240304__ms-web__pf.tsv`)
	if err != nil {
		log.Fatalln(err)
	}

	ps := lo.Values(pMap)
	slices.SortFunc(ps, func(a, b *presenter.Product) int {
		marginA := (a.Price - a.CostPrice) / a.CostPrice
		marginB := (b.Price - b.CostPrice) / b.CostPrice

		if marginA > marginB {
			return -1
		} else if marginA == marginB {
			return 0
		}
		return 1
	})

	bc := bigc.MustGetClient()
	var retv = make([]*Retv, 0)
	for i := 0; i < 100; i++ {
		p := ps[i]

		var pid int
		if p.IsVariant {
			ids := strings.Split(p.BCID, "/")
			pid, err = strconv.Atoi(ids[0])
			if err != nil {
				log.Println(err)
				continue
			}
		} else {
			pid, err = strconv.Atoi(p.BCID)
			if err != nil {
				log.Println(err)
				continue
			}
		}

		bp, err := bc.GetProductById(pid)
		if err != nil {
			log.Println(err)
			continue
		}

		cats := make([]string, 0)
		for _, catId := range bp.Categories {
			cat, err := bc.GetCategoryFromID(catId)
			if err != nil {
				log.Println(err)
				continue
			}

			cats = append(cats, cat.Name)
		}

		retv = append(retv, &Retv{
			Product:    p,
			Categories: cats},
		)
	}

	f, err := os.Create(`C:\Users\admin\Develin Management Dropbox\Zihan\files\out\240306\240306__ms-web__pf__top-100.csv`)
	if err != nil {
		log.Fatalln(err)
	}

	defer f.Close()

	if err := gocsv.MarshalFile(&retv, f); err != nil {
		log.Fatalln(err)
	}

	log.Println("done")

}

func getCostMap(path string) (map[string]float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var prlwgp = make([]*report.ProductRetailListLine, 0)

	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.LazyQuotes = true
		return r
	})

	if err := gocsv.UnmarshalFile(f, &prlwgp); err != nil {
		return nil, err
	}

	return lo.Associate(prlwgp, func(l *report.ProductRetailListLine) (string, float64) {
		return strings.TrimSpace(l.Sku), l.Cost.Float64()
	}), nil
}

func getProductMap(path string) (map[string]*presenter.Product, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var ps = make([]*presenter.Product, 0)
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = '\t'
		r.LazyQuotes = true
		return r
	})

	if err := gocsv.UnmarshalFile(f, &ps); err != nil {
		return nil, err
	}

	ps = lo.Filter(ps, func(p *presenter.Product, _ int) bool {
		return p.OnWeb == 1 && p.Price > 0 && p.CostPrice > 0 && !strings.HasPrefix(p.ProdName, `\`) && p.StockInformation.Total > 0
	})

	pMap := lo.Associate(ps, func(p *presenter.Product) (string, *presenter.Product) {
		return strings.TrimSpace(p.Sku), p
	})

	return pMap, nil
}
