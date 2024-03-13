package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/gocarina/gocsv"
	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/plantyplantman/bcapi/pkg/product"
	"github.com/samber/lo"
)

func main() {
	delete()
}

func check() {
	path := `C:\Users\admin\Develin Management Dropbox\Zihan\files\out\240226\neverSold-retired.csv`
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Read()

	lines, err := r.ReadAll()
	if err != nil {
		panic(err)
	}

	skus := lo.Map(lines, func(line []string, _ int) string {
		sid := strings.TrimSpace(line[3])
		sid = strings.TrimPrefix(sid, "//")
		return sid
	})

	s, err := product.NewDefaultService()
	if err != nil {
		panic(err)
	}

	ps, err := s.FetchProducts(product.WithSkus(skus...))
	if err != nil {
		panic(err)
	}

	if len(ps) != len(skus) {
		fmt.Println("missing", len(skus)-len(ps), "products")
		return
	}

	for _, p := range ps {
		if p.StockInformation.Total > 0 {
			fmt.Println(p.Sku, p.StockInformation.Total)
		}
	}

}

func delete() {
	path := `C:\Users\admin\Develin Management Dropbox\Zihan\files\out\240226\neverSold-retired.csv`
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var ps = make([]bigc.Product, 0)
	if err := gocsv.UnmarshalFile(f, &ps); err != nil {
		panic(err)
	}

	ids := lo.Map(ps, func(p bigc.Product, _ int) int {
		return p.ID
	})

	bc := bigc.MustGetClient()
	bc.MaxWorkers = 10

	for _, id := range ids {
		if err := bc.DeleteProduct(id); err != nil {
			log.Println(err)
		}
	}
}

func openFilterExport() {
	path := `C:\Users\admin\Develin Management Dropbox\Zihan\files\out\240219\Never Sold.csv`
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Read()

	lines, err := r.ReadAll()
	if err != nil {
		panic(err)
	}

	skus := lo.FilterMap(lines, func(line []string, _ int) (string, bool) {
		sku := strings.TrimSpace(line[1])
		return sku, strings.HasPrefix(sku, "//")
	})

	bc := bigc.MustGetClient()
	ps := make([]bigc.Product, 0)
	for _, sku := range skus {
		p, err := bc.GetProductFromSku(sku)
		if err != nil {
			log.Println(err)
			continue
		}

		ps = append(ps, p)
	}

	ps = lo.FilterMap(ps, func(p bigc.Product, _ int) (bigc.Product, bool) {
		return p, slices.Contains(p.Categories, 1230) && len(p.Variants) == 1
	})

	f, err = os.Create("neverSold-retired.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err = gocsv.MarshalFile(ps, f); err != nil {
		panic(err)
	}
	fmt.Println("done")
}
