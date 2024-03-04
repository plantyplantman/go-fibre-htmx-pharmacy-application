package main

import (
	"log"
	"strconv"

	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/samber/lo"
)

func main() {
	c := bigc.MustGetClient()
	ps, err := c.GetAllProducts(map[string]string{"is_visible": "true", "include": "variants"})
	if err != nil {
		log.Fatalln(err)
	}

	ps = lo.Filter(ps, func(item bigc.Product, index int) bool {
		return len(item.Variants) > 1
	})

	for _, p := range ps {
		parentId := strconv.Itoa(p.ID)
		_, err := c.UpdateProduct(&p, bigc.WithUpdateProductSku(""), bigc.WithUpdateProductBinPickingNumber(parentId), bigc.WithUpdateProductGtin(""))
		if err != nil {
			log.Println(err)
			continue
		}

		for _, v := range p.Variants {
			_, err := c.UpdateVariant(&v, bigc.WithUpdateVariantBinPickingNumber(parentId), bigc.WithUpdateVariantGtin(v.Sku), bigc.WithUpdateVariantUpc(v.Sku))
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}
