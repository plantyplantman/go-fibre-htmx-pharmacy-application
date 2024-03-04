package main

import (
	"log"
	"slices"

	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/samber/lo"
)

func main() {
	c := bigc.MustGetClient()
	ps, err := c.GetAllProducts(map[string]string{})
	if err != nil {
		log.Fatalln(err)
	}

	ps = lo.Filter(ps, func(item bigc.Product, index int) bool {
		return slices.Contains(item.Categories, bigc.BLACKFRIDAY)
	})

	for _, p := range ps {
		_, err = c.UpdateProduct(&p,
			bigc.WithUpdateProductCategoriesWithoutSaleIDs(p.Categories),
			bigc.WithUpdateProductSalePrice(0))
		if err != nil {
			log.Println(err)
			continue
		}
	}
}
