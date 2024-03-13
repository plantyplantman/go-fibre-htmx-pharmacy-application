package main

import (
	"log"

	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/samber/lo"
)

func main() {
	c := bigc.MustGetClient()
	ps, err := c.GetAllProducts(map[string]string{"is_visible": "false"})
	if err != nil {
		panic(err)
	}

	ps = lo.Filter(ps, func(item bigc.Product, index int) bool {
		return item.SalePrice != 0.0
	})

	for _, p := range ps {
		_, err := c.UpdateProduct(&p,
			bigc.WithUpdateProductCategoriesWithoutSaleIDs(p.Categories),
			bigc.WithUpdateProductSalePrice(0))
		if err != nil {
			log.Println(err)
		}
	}
}
