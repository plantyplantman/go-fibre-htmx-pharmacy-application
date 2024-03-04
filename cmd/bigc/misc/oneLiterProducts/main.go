package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gocarina/gocsv"
	"github.com/plantyplantman/bcapi/api/presenter"
	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/plantyplantman/bcapi/pkg/product"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type Retv struct {
	Sku    string  `csv:"sku"`
	Name   string  `csv:"name"`
	Weight float64 `csv:"weight"`
	Width  float64 `csv:"width"`
	Depth  float64 `csv:"depth"`
	Height float64 `csv:"height"`
}

func main() {
	service, err := product.NewDefaultService()
	if err != nil {
		panic(err)
	}

	ps, err := service.FetchProducts(func(d *gorm.DB) *gorm.DB {
		return d.Where("name LIKE ? AND on_web = 1", "%"+"1%L%")
	})
	if err != nil {
		panic(err)
	}

	pids := lo.FilterMap(ps, func(p *presenter.Product, _ int) (int, bool) {
		if strings.Contains(p.BCID, "/") {
			return 0, false
		}

		var retv int
		if retv, err = strconv.Atoi(p.BCID); err != nil {
			return 0, false
		}

		return retv, true
	})

	vids := lo.FilterMap(ps, func(p *presenter.Product, _ int) (lo.Tuple2[int, int], bool) {
		if !strings.Contains(p.BCID, "/") {
			return lo.Tuple2[int, int]{}, false
		}

		ids := strings.Split(p.BCID, "/")
		svid := ids[1]
		vid, err := strconv.Atoi(svid)
		if err != nil {
			return lo.Tuple2[int, int]{}, false
		}

		spid := ids[0]
		pid, err := strconv.Atoi(spid)
		if err != nil {
			return lo.Tuple2[int, int]{}, false
		}

		return lo.Tuple2[int, int]{A: pid, B: vid}, true
	})

	bc := bigc.MustGetClient()
	retv := make([]Retv, 0)
	for _, id := range pids {
		p, err := bc.GetProductById(id)
		if err != nil {
			panic(err)
		}

		retv = append(retv, Retv{
			Sku:    p.Sku,
			Name:   p.Name,
			Weight: p.Weight,
			Width:  p.Width,
			Depth:  p.Depth,
			Height: p.Height,
		})
	}

	for _, id := range vids {
		p, err := bc.GetProductById(id.A)
		if err != nil {
			panic(err)
		}

		v, err := bc.GetVariantById(id.B, id.A, map[string]string{})
		if err != nil {
			panic(err)
		}

		retv = append(retv, Retv{
			Sku:    v.Sku,
			Name:   p.Name + "/" + v.OptionValues[0].OptionDisplayName,
			Weight: v.Weight,
			Width:  v.Width,
			Depth:  v.Depth,
			Height: v.Height,
		})
	}

	f, err := os.Create("1L-Products2.csv")
	if err != nil {
		panic(err)
	}
	if err := gocsv.MarshalFile(&retv, f); err != nil {
		panic(err)
	}
	fmt.Println("done")
}
