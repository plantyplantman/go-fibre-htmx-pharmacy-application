package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/plantyplantman/bcapi/api/presenter"
	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/plantyplantman/bcapi/pkg/env"
	"github.com/plantyplantman/bcapi/pkg/parser"
	"github.com/plantyplantman/bcapi/pkg/product"
	"github.com/plantyplantman/bcapi/pkg/report"
	"github.com/samber/lo"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	b, err := os.ReadFile("C:/Users/admin/source/repos/minfos-test/minfos-test/bin/Debug/231031__petrie__promos.xml")
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
		fmt.Println(o.OfferName)
		if o.OfferName == "Aug Sept 20%" ||
			o.OfferName == "Aug Sept 10%" ||
			o.OfferName == "Catalouge Set Sales" ||
			o.OfferName == "Aug Sept Oct Vitamin Promo" ||
			o.OfferName == "August Sept Oct Perm Promo" ||
			o.OfferName == "40% off skincare" ||
			o.OfferName == "Deleted Promo July 60% off" ||
			o.OfferName == "Deleted Promo July 50% off" ||
			o.OfferName == "Deleted Promo July 40% off" ||
			o.OfferName == "Deleted July Promo 30% off" ||
			o.OfferName == "Deleted Promo July 20% off" {
			for _, p := range o.Products.Product {
				skus = append(skus, p.EAN)
			}
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
		c := bigc.MustGetClient()

		ps, err := service.FetchProducts(product.WithSkus(skus...))
		if err != nil {
			log.Fatalln(err)
		}

		pBCIDs := lo.FilterMap(ps, func(p *presenter.Product, _ int) (string, bool) {
			tokens := strings.Split(p.BCID, "/")
			return tokens[0], len(tokens) == 1
		})

		bcPs, err := c.GetAllProducts(
			map[string]string{
				"id:in": "[" + lo.Reduce(
					pBCIDs, func(agg string, item string, _ int) string {
						return agg + "," + item
					}, "") + "]"})
		if err != nil {
			log.Fatalln(err)
		}

		for _, p := range bcPs {
			_, err = c.UpdateProduct(&p,
				bigc.WithUpdateProductCategoriesWithoutSaleIDs(p.Categories),
				bigc.WithUpdateProductSalePrice(0))
			if err != nil {
				log.Println(err)
				continue
			}
		}

		type bcid struct {
			pId int
			vId int
		}
		vBCIDs := lo.FilterMap(ps, func(p *presenter.Product, _ int) (*bcid, bool) {
			tokens := strings.Split(p.BCID, "/")
			if len(tokens) == 2 {
				pid, err := strconv.Atoi(tokens[0])
				if err != nil {
					return nil, false
				}
				vid, err := strconv.Atoi(tokens[1])
				if err != nil {
					return nil, false
				}
				return &bcid{
					pId: pid,
					vId: vid,
				}, true
			}
			return nil, false
		})

		for _, bcid := range vBCIDs {
			v, err := c.GetVariantById(bcid.vId, bcid.pId, map[string]string{})
			if err != nil {
				log.Fatalln(err)
			}
			_, err = c.UpdateVariant(v, bigc.WithUpdateVariantSalePrice(0))
			if err != nil {
				log.Println(err)
				continue
			}
		}

	}

}
