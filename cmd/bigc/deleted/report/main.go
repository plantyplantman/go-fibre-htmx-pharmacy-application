package main

import (
	"log"
	"strings"

	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/samber/lo"
)

func main() {
	c := bigc.MustGetClient()

	ps, err := c.GetAllProducts(map[string]string{})
	if err != nil {
		log.Fatalln(err)
	}

	var rv bigc.Products
	rv = lo.Filter(ps, func(p bigc.Product, _ int) bool {
		return strings.HasPrefix(p.Sku, "/")
	})

	err = rv.Export("231101__web__deleted-report.tsv")
	if err != nil {
		log.Fatal(err)
	}

	log.Println(`Exported to C:\Users\admin\Develin Management Dropbox\Zihan\code\api\231101__web__deleted-report.tsv`)
}
