package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/plantyplantman/bcapi/pkg/product"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

func main() {
	service, err := product.NewDefaultService()
	if err != nil {
		log.Fatalln(err)
	}

	ps, err := service.FetchProducts(func(d *gorm.DB) *gorm.DB {
		return d.Where("name LIKE ?", "ARCHLINE%")
	}, product.WithStockInformation())
	if err != nil {
		log.Fatalln(err)
	}

	bc := bigc.MustGetClient()
	for _, p := range ps {
		if p.BCID == "" {
			log.Println("No BCID for ", p.Sku, " ", p.ProdName)
			continue
		}

		id, err := strconv.Atoi(p.BCID)
		if err != nil {
			log.Println(err)
			continue
		}

		bcp, err := bc.GetProductById(id)
		if err != nil {
			log.Println(err)
			continue
		}

		_, err = bc.UpdateProduct(bcp, bigc.WithUpdateProductInventoryLevel(p.StockInformation.Total))
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func run() {
	c := bigc.MustGetClient()

	ps, err := c.GetAllProducts(map[string]string{"keyword": "moogoo", "categories:in": fmt.Sprint(bigc.RETIRED_PRODUCTS)})
	if err != nil {
		log.Fatalln(err)
	}

	for _, p := range ps {
		if strings.HasPrefix(p.Sku, "/") {
			continue
		}

		new_cats := lo.Filter(p.Categories, func(c int, _ int) bool {
			return c != bigc.RETIRED_PRODUCTS
		})

		_, err := c.UpdateProduct(&p, bigc.WithUpdateProductCategories(unique(new_cats)))
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func unique(intSlice []int) []int {
	keys := make(map[int]bool)
	list := []int{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
