package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/plantyplantman/bcapi/pkg/env"
	"github.com/plantyplantman/bcapi/pkg/parser"
	"github.com/plantyplantman/bcapi/pkg/product"
	"github.com/plantyplantman/bcapi/pkg/report"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	path := `C:\Users\admin\Develin Management Dropbox\Zihan\files\in\231123\231123__petrie__black-friday.xml`
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	p := parser.NewXmlParser(b)
	r := report.Campaigns{}
	err = p.Parse(&r)
	if err != nil {
		log.Fatal(err)
	}

	var skus = make([]string, 0)
	for _, o := range r.Campaign.Offers.Offer {
		for _, p := range o.Products.Product {
			skus = append(skus, p.EAN)
		}
	}

	for _, sku := range skus {
		fmt.Println(sku)
	}

	var connString = env.TEST_NEON
	if connString == "" {
		log.Fatalln("TEST_NEON_CONNECTION_STRING not set")
	}
	DB, err := gorm.Open(postgres.Open(connString), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	repo := product.NewRepository(DB)
	service := product.NewService(repo)

	ps, err := service.FetchProducts(product.WithSkus(skus...))
	if err != nil {
		log.Fatalln(err)
	}

	c := bigc.MustGetClient()

	for _, p := range ps {
		if p.IsVariant {
			ids := strings.Split(p.BCID, "/")
			if len(ids) < 2 {
				log.Println("no variant id")
				continue
			}
			pid, err := strconv.Atoi(ids[0])
			if err != nil {
				log.Println(err)
				continue
			}

			vid, err := strconv.Atoi(ids[1])
			if err != nil {
				log.Println(err)
				continue
			}

			v, err := c.GetVariantById(vid, pid, map[string]string{})
			if err != nil {
				log.Println(err)
			}

			_, err = c.UpdateVariant(v, bigc.WithUpdateVariantSku(p.Sku))
			if err != nil {
				log.Println(err)
				continue
			}
		} else {
			id, err := strconv.Atoi(p.BCID)
			if err != nil {
				log.Println(err)
				continue
			}
			bp, err := c.GetProductById(id)
			if err != nil {
				log.Println(err)
				continue
			}

			_, err = c.UpdateProduct(bp, bigc.WithUpdateProductSku(p.Sku))
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}

}
