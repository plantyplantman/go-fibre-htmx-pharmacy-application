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

	path := `C:\Users\admin\Desktop\promos\deletedmarch030324.xml`
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
			fmt.Println(p.EAN)
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

	ps, err := service.FetchProducts(product.WithSkus(skus...))
	if err != nil {
		log.Fatalln(err)
	}

	psSkuM := lo.Associate(ps, func(p *presenter.Product) (string, *presenter.Product) {
		return p.Sku, p
	})
	c := bigc.MustGetClient()

	for _, o := range r.Campaign.Offers.Offer {
		discAmount := 0.0
		switch strings.TrimSpace(o.PercentOffDisc) {
		case "0.10":
			discAmount = 0.1
		case "0.20":
			discAmount = 0.2
		case "0.30":
			discAmount = 0.3
		case "0.40":
			discAmount = 0.4
		case "0.50":
			discAmount = 0.5
		case "0.60":
			discAmount = 0.6
		}

		for _, p := range o.Products.Product {
			if ep, ok := psSkuM[p.EAN]; ok {
				sku := strings.TrimSpace(ep.Sku)
				if strings.HasPrefix(ep.ProdName, "\\") || strings.HasPrefix(ep.ProdName, "#") || strings.HasPrefix(ep.ProdName, "#") {
					if !strings.HasPrefix(sku, "//") {
						if !strings.HasPrefix(sku, "/") {
							sku = "/" + sku
						}
						sku = "/" + sku
					}
				}

				if ep.IsVariant {
					ids := strings.Split(ep.BCID, "/")
					if len(ids) != 2 {
						log.Printf("\nexpected at 2 ids, got: %d\tLine: %d", len(ids), 0)
						continue
					}
					vID, err := strconv.Atoi(ids[1])
					if err != nil {
						log.Println(err)
						continue
					}
					pID, err := strconv.Atoi(ids[0])
					if err != nil {
						log.Println(err)
						continue
					}

					v, err := c.GetVariantById(vID, pID, map[string]string{})
					if err != nil {
						log.Println(err)
						continue
					}
					if discAmount == 0.0 && p.OfferPrice != 0.0 {
						_, err = c.UpdateVariant(v, bigc.WithUpdateVariantSalePrice(p.OfferPrice), bigc.WithUpdateVariantSku(sku))
						if err != nil {
							log.Println(err)
							continue
						}
					} else {
						_, err = c.UpdateVariant(v, bigc.WithUpdateVariantSalePrice(v.Price-(v.Price*discAmount)), bigc.WithUpdateVariantSku(sku))
						if err != nil {
							log.Println(err)
							continue
						}
					}
				} else {
					id, err := strconv.Atoi(ep.BCID)
					if err != nil {
						log.Println(err)
						continue
					}
					bcp, err := c.GetProductById(id)
					if err != nil {
						log.Println(err)
						continue
					}

					cats := bigc.RemoveSaleCategories(bcp.Categories)
					cats = append(cats, bigc.PROMOTIONS, bigc.PRODUCTSALE, bigc.CLEARANCE)

					if o.OfferName == "20% off November December" {
						cats = append(cats, bigc.SALE_20, bigc.PRODUCTSALE_20)
					} else {
						cats = append(cats, bigc.PROMO_SET_SALES)
						if discAmount > 0.57 {
							cats = append(cats, bigc.SALE_50, bigc.PRODUCTSALE_60)
						} else if discAmount > 0.47 {
							cats = append(cats, bigc.SALE_50, bigc.PRODUCTSALE_50)
						} else if discAmount > 0.37 {
							cats = append(cats, bigc.SALE_40, bigc.PRODUCTSALE_40)
						} else if discAmount > 0.27 {
							cats = append(cats, bigc.SALE_30, bigc.PRODUCTSALE_30)
						} else if discAmount > 0.17 {
							cats = append(cats, bigc.SALE_20, bigc.PRODUCTSALE_20)
						} else if discAmount > 0.13 {
							cats = append(cats, bigc.SALE_15, bigc.PRODUCTSALE_15)
						} else if discAmount > 0.07 {
							cats = append(cats, bigc.SALE_10, bigc.PRODUCTSALE_10)
						}
					}

					if discAmount == 0.0 && p.OfferPrice != 0.0 {
						discpercent := (bcp.Price - p.OfferPrice) / bcp.Price
						if discpercent > 0.57 {
							cats = append(cats, bigc.SALE_50, bigc.PRODUCTSALE_60)
						} else if discpercent > 0.47 {
							cats = append(cats, bigc.SALE_50, bigc.PRODUCTSALE_50)
						} else if discpercent > 0.37 {
							cats = append(cats, bigc.SALE_40, bigc.PRODUCTSALE_40)
						} else if discpercent > 0.27 {
							cats = append(cats, bigc.SALE_30, bigc.PRODUCTSALE_30)
						} else if discpercent > 0.17 {
							cats = append(cats, bigc.SALE_20, bigc.PRODUCTSALE_20)
						} else if discpercent > 0.13 {
							cats = append(cats, bigc.SALE_15, bigc.PRODUCTSALE_15)
						} else if discpercent > 0.07 {
							cats = append(cats, bigc.SALE_10, bigc.PRODUCTSALE_10)
						}
					}

					if discAmount == 0.0 && p.OfferPrice != 0.0 {
						_, err = c.UpdateProduct(bcp,
							bigc.WithUpdateProductSalePrice(p.OfferPrice),
							bigc.WithUpdateProductCategories(cats),
							bigc.WithUpdateProductSku(sku),
						)
						if err != nil {
							log.Println(err)
							continue
						}
					} else {
						_, err = c.UpdateProduct(bcp,
							bigc.WithUpdateProductSalePrice(bcp.Price-(bcp.Price*discAmount)),
							bigc.WithUpdateProductCategories(cats),
							bigc.WithUpdateProductSku(sku),
						)
						if err != nil {
							log.Println(err)
							continue
						}
					}
				}
			}
		}
	}
}
