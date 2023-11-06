package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/samber/lo"
)

func main() {
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
