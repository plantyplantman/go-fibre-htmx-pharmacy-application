package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/plantyplantman/bcapi/api/presenter"
	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/plantyplantman/bcapi/pkg/parser"
	"github.com/plantyplantman/bcapi/pkg/product"
	"github.com/plantyplantman/bcapi/pkg/report"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

func main() {
	bc := bigc.MustGetClient()
	service, err := product.NewDefaultService()
	if err != nil {
		log.Fatalln(err)
	}

	ai_c, err := bigc.GetOpenAiClient()
	if err != nil {
		log.Fatalln(err)
	}

	f := fiber.New(
		fiber.Config{
			Immutable: true,
		},
	)
	f.Use(cors.New())
	f.Use(recover.New())

	f.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	f.Get("/images", func(c *fiber.Ctx) error {
		skusStr := c.Query("skus")
		log.Println(skusStr)
		skus := strings.Split(skusStr, ",")
		ps, err := service.FetchProducts(product.WithSkus(skus...))
		if err != nil {
			log.Fatalln(err)
		}

		var retv = make([][]string, 0)
		for _, p := range ps {
			ids := strings.Split(p.BCID, "/")
			if p.IsVariant && len(ids) == 2 {
				vid, err := strconv.Atoi(ids[1])
				if err != nil {
					log.Println(err)
					continue
				}
				pid, err := strconv.Atoi(ids[0])
				if err != nil {
					log.Println(err)
					continue
				}
				vp, err := bc.GetVariantById(vid, pid, map[string]string{})
				if err != nil {
					log.Println(err)
					continue
				}
				retv = append(retv, []string{vp.Sku, vp.ImageURL})
			} else if !p.IsVariant && len(ids) == 1 && ids[0] != "" {
				pid, err := strconv.Atoi(ids[0])
				if err != nil {
					log.Println(err)
					continue
				}
				pp, err := bc.GetProductById(pid)
				if err != nil {
					log.Println(err)
					continue
				}
				if len(pp.Variants) == 0 {
					continue
				}
				retv = append(retv, []string{pp.Sku, pp.Variants[0].ImageURL})
			} else {
				log.Printf("skipping %s with sku %s due to invalid BCID: %s\n", p.ProdName, p.Sku, p.BCID)
			}
		}

		c.Status(200)
		return c.JSON(retv)
	})

	f.Post("/multistore", func(c *fiber.Ctx) error {
		dtstr := c.Query("date")
		dt, err := time.Parse("060102", strings.TrimSpace(dtstr))
		if err != nil {
			log.Println(err)
			return err
		}
		fh, err := c.FormFile("file")
		if err != nil {
			log.Println(err)
			return err
		}

		ff, err := fh.Open()
		if err != nil {
			log.Println(err)
			return err
		}

		log.Printf("File name: %s\n", fh.Filename)
		log.Printf("File size: %d\n", fh.Size)

		r := csv.NewReader(ff)
		r.Comma = '\t'
		r.LazyQuotes = true

		prls, err := parser.ParseMultistoreInput(r, dt, fh.Filename)
		if err != nil {
			log.Println(err)
			return err
		}

		nosr := report.DoMultistore(prls, service, bc)

		if len(nosr) == 0 {
			c.Status(204)
			return c.SendString("All products on site")
		}

		return c.JSON(nosr)
	})

	f.Post("/multistore/action", func(c *fiber.Ctx) error {
		fh, err := c.FormFile("file")
		if err != nil {
			log.Println(err)
			return err
		}

		ff, err := fh.Open()
		if err != nil {
			log.Println(err)
			return err
		}

		log.Printf("File name: %s\n", fh.Filename)
		log.Printf("File size: %d\n", fh.Size)

		p, err := parser.NewCsvParser(ff)
		if err != nil {
			return err
		}

		nosr := report.NotOnSiteReport{}
		if err := p.Parse(&nosr); err != nil {
			return err
		}

		retv, err := nosr.Action(service, bc, &ai_c)
		if err != nil {
			return err
		}

		return c.JSON(retv)
	})

	f.Post("/promos/new", handlePromosNew(service, bc))
	f.Post("/promos/deleted", handlePromosDeleted(service, bc))

	f.Get("/rangeReview", handleGetRangeReview(service))
	f.Post("/rangeReview", handlePostRangeReview(service))

	f.Post("/product/update", handlePostProductUpdate(bc, service))

	f.Get("/product/zr", handleGetZReport(bc, service))
	f.Post("/product/top", handlePostTopSales(service))
	f.Post("/product/update/skus", handlePostUpdateSkus(service, bc))
	f.Get("/product/skus", func(c *fiber.Ctx) error {
		skus := c.Query("skus")
		if skus == "" {
			return fiber.NewError(400, "skus query param is required")
		}
		ps, err := service.FetchProducts(product.WithSkus(strings.Split(skus, ",")...), product.WithStockInformation())
		if err != nil {
			return err
		}

		return c.JSON(ps)
	})

	log.Fatal(f.Listen(":8080"))
}

func handlePostUpdateSkus(service product.Service, bc *bigc.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// skus := c.Query("skus")
		// if skus == "" {
		// 	return fiber.NewError(400, "skus query param is required")
		// }

		var skus = make([]string, 0)
		if err := json.Unmarshal(c.Body(), &skus); err != nil {
			return err
		}

		ps, err := service.FetchProducts(product.WithSkus(skus...), product.WithStockInformation())
		if err != nil {
			return fiber.NewError(400, "Failed to fetch products from service: "+err.Error())
		}

		psM := lo.FilterMap(ps, func(p *presenter.Product, _ int) (*presenter.Product, bool) {
			return p, !p.IsVariant && p.BCID != ""
		})

		vsM := lo.FilterMap(ps, func(p *presenter.Product, _ int) (*presenter.Product, bool) {
			return p, p.IsVariant && p.BCID != ""
		})

		var retv = Retv{}
		for _, p := range psM {
			ids := strings.Split(p.BCID, "/")
			if len(ids) != 1 {
				continue
			}
			id, err := strconv.Atoi(ids[0])
			if err != nil {
				continue
			}
			pp, err := bc.GetProductById(id)
			if err != nil {
				continue
			}
			updateFuncs := []bigc.ProductUpdateOptFn{}
			if pp.Price != p.Price {
				updateFuncs = append(updateFuncs, bigc.WithUpdateProductPrice(p.Price))
			}
			if pp.InventoryLevel != p.StockInformation.Total {
				updateFuncs = append(updateFuncs, bigc.WithUpdateProductInventoryLevel(p.StockInformation.Total))
			}
			sku := pp.Sku
			if strings.HasPrefix(p.ProdName, "\\") {
				if !strings.HasPrefix(sku, "/") {
					sku = "/" + sku
				}
				if p.StockInformation.Total == 0 && !strings.HasPrefix(pp.Sku, "//") {
					sku = "/" + sku
					updateFuncs = append(updateFuncs, bigc.WithUpdateProductCategoryIsRetired(true), bigc.WithUpdateProductIsVisible(true))
				}

				if pp.Sku != p.Sku {
					updateFuncs = append(updateFuncs, bigc.WithUpdateProductSku(sku))
				}

			}

			if len(updateFuncs) == 0 {
				continue
			}

			newp, err := bc.UpdateProduct(pp, updateFuncs...)
			if err != nil {
				retv.Errors = append(retv.Errors, err.Error())
				continue
			}
			retv.New = append(retv.New, newp)
			retv.Prev = append(retv.Prev, pp)
		}

		for _, p := range vsM {
			ids := strings.Split(p.BCID, "/")
			if len(ids) != 2 {
				continue
			}
			pid, err := strconv.Atoi(ids[0])
			if err != nil {
				continue
			}
			vid, err := strconv.Atoi(ids[1])
			if err != nil {
				continue
			}
			vp, err := bc.GetVariantById(vid, pid, map[string]string{})
			if err != nil {
				continue
			}

			updateFuncs := []bigc.UpdateVariantOpt{}
			if vp.Price != p.Price {
				updateFuncs = append(updateFuncs, bigc.WithUpdateVariantPrice(p.Price))
			}
			if vp.InventoryLevel != p.StockInformation.Total {
				updateFuncs = append(updateFuncs, bigc.WithUpdateVariantInventoryLevel(p.StockInformation.Total))
			}
			sku := vp.Sku
			if strings.HasPrefix(p.ProdName, "\\") {
				if !strings.HasPrefix(sku, "/") {
					sku = "/" + sku
				}
				if p.StockInformation.Total == 0 && !strings.HasPrefix(vp.Sku, "//") {
					sku = "/" + sku
					updateFuncs = append(updateFuncs, bigc.WithUpdateVariantPurchasingDisabled(true))
				}
			}

			if vp.Sku != p.Sku {
				updateFuncs = append(updateFuncs, bigc.WithUpdateVariantSku(sku))
			}

			if len(updateFuncs) == 0 {
				continue
			}
			newv, err := bc.UpdateVariant(vp, updateFuncs...)
			if err != nil {
				log.Println(err)
				retv.Errors = append(retv.Errors, err.Error())
				continue
			}

			retv.New = append(retv.New, newv)
			retv.Prev = append(retv.Prev, vp)

		}

		return c.JSON(retv)
	}
}

type Retv struct {
	Ok     bool
	Errors []string
	New    []any
	Prev   []any
}

func handlePromosNew(service product.Service, bc *bigc.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var retv = Retv{}
		fh, err := c.FormFile("file")
		if err != nil {
			retv.Errors = append(retv.Errors, err.Error())
			retv.Ok = false
			return c.JSON(retv)
		}

		ff, err := fh.Open()
		if err != nil {
			retv.Errors = append(retv.Errors, err.Error())
			retv.Ok = false
			return c.JSON(retv)
		}

		b, err := io.ReadAll(ff)
		if err != nil {
			retv.Errors = append(retv.Errors, err.Error())
			retv.Ok = false
			return c.JSON(retv)
		}

		p := parser.NewXmlParser(b)
		r := report.Campaigns{}
		err = p.Parse(&r)
		if err != nil {
			retv.Errors = append(retv.Errors, err.Error())
			retv.Ok = false
			return c.JSON(retv)
		}

		var skus = make([]string, 0)
		for _, o := range r.Campaign.Offers.Offer {
			for _, p := range o.Products.Product {
				skus = append(skus, p.EAN)
			}
		}

		ps, err := service.FetchProducts(product.WithSkus(skus...))
		if err != nil {
			retv.Errors = append(retv.Errors, err.Error())
			retv.Ok = false
			return c.JSON(retv)
		}

		psSkuM := lo.Associate(ps, func(p *presenter.Product) (string, *presenter.Product) {
			return p.Sku, p
		})

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
							retv.Errors = append(retv.Errors, "expected at 2 ids, got: "+strconv.Itoa(len(ids))+"\tSku:"+"sku")
							continue
						}
						vID, err := strconv.Atoi(ids[1])
						if err != nil {
							retv.Errors = append(retv.Errors, err.Error())
							continue
						}
						pID, err := strconv.Atoi(ids[0])
						if err != nil {
							retv.Errors = append(retv.Errors, err.Error())
							continue
						}

						v, err := bc.GetVariantById(vID, pID, map[string]string{})
						if err != nil {
							retv.Errors = append(retv.Errors, err.Error())
							continue
						}
						if discAmount == 0.0 && p.OfferPrice != 0.0 {
							_, err = bc.UpdateVariant(v, bigc.WithUpdateVariantSalePrice(p.OfferPrice), bigc.WithUpdateVariantSku(sku))
							if err != nil {
								retv.Errors = append(retv.Errors, err.Error())
								continue
							}
						} else {
							newv, err := bc.UpdateVariant(v, bigc.WithUpdateVariantSalePrice(v.Price-(v.Price*discAmount)), bigc.WithUpdateVariantSku(sku))
							if err != nil {
								retv.Errors = append(retv.Errors, err.Error())
								continue
							}

							retv.New = append(retv.New, newv)
							retv.Prev = append(retv.Prev, v)
						}
					} else {
						id, err := strconv.Atoi(ep.BCID)
						if err != nil {
							retv.Errors = append(retv.Errors, err.Error())
							continue
						}
						bcp, err := bc.GetProductById(id)
						if err != nil {
							retv.Errors = append(retv.Errors, err.Error())
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
							} else if discpercent > 0.07 {
								cats = append(cats, bigc.SALE_10, bigc.PRODUCTSALE_10)
							}
						}

						if discAmount == 0.0 && p.OfferPrice != 0.0 {
							newp, err := bc.UpdateProduct(bcp,
								bigc.WithUpdateProductSalePrice(p.OfferPrice),
								bigc.WithUpdateProductCategories(cats),
								bigc.WithUpdateProductSku(sku),
							)
							if err != nil {
								retv.Errors = append(retv.Errors, err.Error())
								continue
							}

							retv.New = append(retv.New, newp)
						} else {
							newp, err := bc.UpdateProduct(bcp,
								bigc.WithUpdateProductSalePrice(bcp.Price-(bcp.Price*discAmount)),
								bigc.WithUpdateProductCategories(cats),
								bigc.WithUpdateProductSku(sku),
							)
							if err != nil {
								retv.Errors = append(retv.Errors, err.Error())
								continue
							}

							retv.New = append(retv.New, newp)
						}

						retv.Prev = append(retv.Prev, bcp)
					}
				}
			}
		}
		retv.Ok = true
		return c.JSON(retv)
	}
}

func handlePromosDeleted(service product.Service, bc *bigc.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var retv = Retv{}
		fh, err := c.FormFile("file")
		if err != nil {
			retv.Errors = append(retv.Errors, err.Error())
			retv.Ok = false
			return c.JSON(retv)
		}

		ff, err := fh.Open()
		if err != nil {
			retv.Errors = append(retv.Errors, err.Error())
			retv.Ok = false
			return c.JSON(retv)
		}

		b, err := io.ReadAll(ff)
		if err != nil {
			retv.Errors = append(retv.Errors, err.Error())
			retv.Ok = false
			return c.JSON(retv)
		}

		p := parser.NewXmlParser(b)
		r := report.Campaigns{}
		err = p.Parse(&r)
		if err != nil {
			retv.Errors = append(retv.Errors, err.Error())
			retv.Ok = false
			return c.JSON(retv)
		}

		var skus = make([]string, 0)
		for _, o := range r.Campaign.Offers.Offer {
			for _, p := range o.Products.Product {
				skus = append(skus, p.EAN)
			}
		}

		ps, err := service.FetchProducts(product.WithSkus(skus...))
		if err != nil {
			retv.Errors = append(retv.Errors, err.Error())
			retv.Ok = false
			return c.JSON(retv)
		}

		pBCIDs := lo.FilterMap(ps, func(p *presenter.Product, _ int) (string, bool) {
			tokens := strings.Split(p.BCID, "/")
			return tokens[0], len(tokens) == 1
		})

		pBCIDs2 := splitSlice(pBCIDs, 100)

		var bcPs []bigc.Product
		for _, ids := range pBCIDs2 {
			tmp, err := bc.GetAllProducts(
				map[string]string{
					"id:in": "[" + lo.Reduce(
						ids, func(agg string, item string, _ int) string {
							return agg + "," + item
						}, "") + "]"})
			if err != nil {
				retv.Errors = append(retv.Errors, err.Error())
				continue
			}
			bcPs = append(bcPs, tmp...)
		}

		for _, p := range bcPs {
			retv.Prev = append(retv.Prev, p)
			newp, err := bc.UpdateProduct(&p,
				bigc.WithUpdateProductCategoriesWithoutSaleIDs(p.Categories),
				bigc.WithUpdateProductSalePrice(0))

			retv.New = append(retv.New, newp)
			if err != nil {
				retv.Errors = append(retv.Errors, err.Error())
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
			v, err := bc.GetVariantById(bcid.vId, bcid.pId, map[string]string{})
			retv.Prev = append(retv.Prev, v)
			if err != nil {
				retv.Errors = append(retv.Errors, err.Error())
			}
			v2, err := bc.UpdateVariant(v, bigc.WithUpdateVariantSalePrice(0))
			if err != nil {
				retv.Errors = append(retv.Errors, err.Error())
			}

			retv.New = append(retv.New, v2)
		}
		retv.Ok = true
		return c.JSON(retv)
	}
}

func splitSlice[T any](slice []T, n int) [][]T {
	var chunks [][]T
	for n < len(slice) {
		slice, chunks = slice[n:], append(chunks, slice[0:n:n])
	}
	chunks = append(chunks, slice)
	return chunks
}

func handleGetRangeReview(service product.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		brandQ := c.Query("brand")
		buffer := make([]byte, len(brandQ))
		copy(buffer, brandQ)
		brand := string(buffer)
		log.Println(brand)
		ps, err := fetchProductsByBrand(brand, service)
		if err != nil {
			log.Println(err)
			return err
		}

		report, err := report.NewRangeReview(ps)
		if err != nil {
			log.Println(err)
			return err
		}

		if len(report) == 0 {
			return fiber.NewError(204, "No products found")
		}

		json, err := json.Marshal(report)
		if err != nil {
			log.Println(err)
			return fiber.NewError(500, "Internal server error, please try again later. "+err.Error())
		}

		c.Set("Content-Type", "application/json")
		return c.Status(200).Send(json)
	}
}

func handlePostRangeReview(service product.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fh, err := c.FormFile("file")
		if err != nil {
			log.Println(err)
			return err
		}

		ff, err := fh.Open()
		if err != nil {
			log.Println(err)
			return err
		}

		b, err := io.ReadAll(ff)
		if err != nil {
			log.Println(err)
			return err
		}

		p := parser.NewXmlParser(b)
		r := report.Campaigns{}
		err = p.Parse(&r)
		if err != nil {
			log.Println(err)
			return err
		}

		var skus = make([]string, 0)
		for _, o := range r.Campaign.Offers.Offer {
			for _, p := range o.Products.Product {
				skus = append(skus, p.EAN)
			}
		}

		ps, err := service.FetchProducts(product.WithSkus(skus...), product.WithStockInformation())
		if err != nil {
			log.Println(err)
			return err
		}

		report, err := report.NewRangeReview(ps)
		if err != nil {
			log.Println(err)
			return err
		}

		if len(report) == 0 {
			return fiber.NewError(204, "No products found")
		}

		json, err := json.Marshal(report)
		if err != nil {
			log.Println(err)
			return fiber.NewError(500, "Internal server error, please try again later. "+err.Error())
		}

		c.Set("Content-Type", "application/json")
		return c.Status(200).Send(json)
	}
}

func fetchProductsByBrand(brand string, service product.Service) ([]*presenter.Product, error) {
	products, err := service.FetchProducts(
		func(d *gorm.DB) *gorm.DB {
			return d.Where("name LIKE ?", "%"+strings.ToUpper(strings.TrimSpace(brand))+"%")
		},
		product.WithStockInformation())
	if err != nil {
		return nil, err
	}
	return products, nil
}

type Blacklist struct {
	Skus   []string
	Brands []string
}

func (bl *Blacklist) ApplyForProducts(ps []bigc.Product) []bigc.Product {
	return lo.FilterMap(ps, func(item bigc.Product, index int) (bigc.Product, bool) {
		if strings.Contains(item.Sku, "-") {
			return item, false
		}

		sku, err := bigc.CleanBigCommerceSku(item.Sku)
		if err != nil {
			return item, false
		}

		if slices.Contains(bl.Skus, sku) {
			return item, false
		}

		for _, b := range bl.Brands {
			x, y := strings.ToUpper(strings.TrimSpace(item.Name)), strings.ToUpper(strings.TrimSpace(b))
			if strings.Contains(x, y) {
				return item, false
			}
		}
		return item, true
	})
}

func (bl *Blacklist) ApplyForVariants(vs []bigc.Variant) []bigc.Variant {
	return lo.FilterMap(vs, func(item bigc.Variant, index int) (bigc.Variant, bool) {
		sku, err := bigc.CleanBigCommerceSku(item.Sku)
		if err != nil {
			return item, false
		}

		if slices.Contains(bl.Skus, sku) {
			return item, false
		}

		return item, true
	})
}

func handlePostProductUpdate(bc *bigc.Client, service product.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var blackList = Blacklist{}
		if err := json.Unmarshal(c.Body(), &blackList); err != nil {
			return err
		}

		data, err := bc.GetAllProducts(map[string]string{"is_visible": "true"})
		if err != nil {
			return err
		}

		ps := filterForActiveProducts(data, &blackList)
		pids := getBCIDsForProducts(ps)

		ppsMap, err := fetchProductsFromServiceByBCIDs(service, pids)
		if err != nil {
			return err
		}

		retv := updateProducts(ps, ppsMap, bc)

		vs := filterForActiveVariants(data, &blackList)
		vids := getBCIDsForVariants(vs)
		vpsMap, err := fetchVariantsFromServiceByBCIDs(service, vids)
		if err != nil {
			return err
		}
		retv2 := updateVariants(vs, vpsMap, bc)
		retv.Errors = append(retv.Errors, retv2.Errors...)
		retv.New = append(retv.New, retv2.New...)
		retv.Prev = append(retv.Prev, retv2.Prev...)
		return c.JSON(retv)
	}
}

func updateVariants(vs []bigc.Variant, vpsMap map[string]*presenter.Product, c *bigc.Client) Retv {
	var retv = Retv{}
	for _, v := range vs {
		if strings.HasPrefix(v.Sku, "//") || v.PurchasingDisabled {
			sku := v.Sku
			if !strings.HasPrefix(sku, "//") {
				if !strings.HasPrefix(sku, "/") {
					sku = "/" + sku
				}
				sku = "/" + sku
			}
			newv, err := c.UpdateVariant(&v,
				bigc.WithUpdateVariantInventoryLevel(0),
				bigc.WithUpdateVariantSalePrice(0),
				bigc.WithUpdateVariantSku(sku),
			)
			if err != nil {
				retv.Errors = append(retv.Errors, err.Error())
				continue
			}

			retv.Prev = append(retv.Prev, v)
			retv.New = append(retv.New, newv)
			continue
		}
		if vp, ok := vpsMap[strconv.Itoa(v.ProductID)+"/"+strconv.Itoa(v.ID)]; ok {
			updateFns := []bigc.UpdateVariantOpt{}

			// update price
			if vp.Price != v.Price {
				updateFns = append(updateFns, bigc.WithUpdateVariantPrice(vp.Price))
			}

			// update stock
			if vp.StockInformation.Total != v.InventoryLevel {
				updateFns = append(updateFns, bigc.WithUpdateVariantInventoryLevel(vp.StockInformation.Total))
			}

			// update sku
			if strings.HasPrefix(vp.ProdName, "\\") {
				sku := v.Sku
				if !strings.HasPrefix(sku, "/") {
					sku = "/" + sku
				}

				if vp.StockInformation.Total == 0 {
					if !strings.HasPrefix(sku, "//") {
						sku = "/" + sku
					}
					updateFns = append(updateFns,
						bigc.WithUpdateVariantPurchasingDisabled(true),
						bigc.WithUpdateVariantInventoryLevel(0))
				}

				if sku != v.Sku {
					updateFns = append(updateFns, bigc.WithUpdateVariantSku(sku))
				}
			}

			if len(updateFns) == 0 {
				continue
			}
			newv, err := c.UpdateVariant(&v, updateFns...)
			if err != nil {
				log.Println(err)
				retv.Errors = append(retv.Errors, err.Error())
				continue
			}
			retv.Prev = append(retv.Prev, v)
			retv.New = append(retv.New, newv)
			continue
		}
	}

	return retv
}

func updateProducts(ps []bigc.Product, ppsMap map[string]*presenter.Product, c *bigc.Client) Retv {
	var retv = Retv{}
	for _, p := range ps {
		if strings.HasPrefix(p.Sku, "//") || !p.IsVisible || slices.Contains(p.Categories, 1230) {
			sku := p.Sku
			if !strings.HasPrefix(sku, "//") {
				if !strings.HasPrefix(sku, "/") {
					sku = "/" + sku
				}
				sku = "/" + sku
			}
			newp, err := c.UpdateProduct(&p,
				bigc.WithUpdateProductInventoryLevel(0),
				bigc.WithUpdateProductCategories(bigc.RemoveSaleCategories(p.Categories)),
				bigc.WithUpdateProductIsVisible(false),
				bigc.WithUpdateProductCategoryIsRetired(true),
				bigc.WithUpdateProductSku(sku))
			if err != nil {
				retv.Errors = append(retv.Errors, err.Error())
			}
			retv.New = append(retv.New, newp)
			retv.Prev = append(retv.Prev, p)
			continue
		}
		if pp, ok := ppsMap[strconv.Itoa(p.ID)]; ok {
			updateFns := []bigc.ProductUpdateOptFn{}

			if pp.Price != p.Price {
				updateFns = append(updateFns, bigc.WithUpdateProductPrice(pp.Price))
			}

			if pp.StockInformation.Total != p.InventoryLevel {
				updateFns = append(updateFns, bigc.WithUpdateProductInventoryLevel(pp.StockInformation.Total))
			}

			if strings.HasPrefix(pp.ProdName, "\\") {
				sku := p.Sku
				if !strings.HasPrefix(sku, "/") {
					sku = "/" + sku
				}

				if pp.StockInformation.Total == 0 {
					sku = "/" + sku
					if !strings.HasPrefix(sku, "//") {
						sku = "/" + sku
					}
					updateFns = append(updateFns,
						bigc.WithUpdateProductIsVisible(false),
						bigc.WithUpdateProductCategoryIsRetired(true),
						bigc.WithUpdateProductInventoryLevel(0))
				}

				if sku != p.Sku {
					updateFns = append(updateFns, bigc.WithUpdateProductSku(sku))
				}
			}

			if len(updateFns) == 0 {
				continue
			}
			newp, err := c.UpdateProduct(&p, updateFns...)
			if err != nil {
				retv.Errors = append(retv.Errors, err.Error())
				continue
			}

			retv.New = append(retv.New, newp)
			retv.Prev = append(retv.Prev, p)
		}
	}
	return retv
}

func fetchVariantsFromServiceByBCIDs(service product.Service, vids []string) (map[string]*presenter.Product, error) {
	pps, err := service.FetchProducts(product.WithBCIDs(vids...), product.WithStockInformation())
	if err != nil {
		return nil, err
	}

	return lo.Associate(pps, func(p *presenter.Product) (string, *presenter.Product) {
		return p.BCID, p
	}), nil
}

func fetchProductsFromServiceByBCIDs(service product.Service, pids []string) (map[string]*presenter.Product, error) {
	pps, err := service.FetchProducts(product.WithBCIDs(pids...), product.WithStockInformation())
	if err != nil {
		return nil, err
	}

	return lo.Associate(pps, func(p *presenter.Product) (string, *presenter.Product) {
		return p.BCID, p
	}), nil
}

func getBCIDsForProducts(ps []bigc.Product) []string {
	ids := lo.Map(ps, func(p bigc.Product, _ int) string {
		return strconv.Itoa(p.ID)
	})

	return lo.Uniq(lo.Filter(ids, func(item string, index int) bool {
		return item != ""
	}))
}

func getBCIDsForVariants(vs []bigc.Variant) []string {
	ids := lo.Map(vs, func(v bigc.Variant, _ int) string {
		return strconv.Itoa(v.ProductID) + "/" + strconv.Itoa(v.ID)
	})

	return lo.Uniq(lo.Filter(ids, func(item string, index int) bool {
		return item != ""
	}))
}

func filterForActiveProducts(ps []bigc.Product, blackList *Blacklist) []bigc.Product {
	return lo.Filter(blackList.ApplyForProducts(ps), func(item bigc.Product, index int) bool {
		return item.Sku != "" && item.IsVisible && !strings.HasPrefix(item.Sku, "//") && !slices.Contains(item.Categories, 1230)
	})
}

func filterForActiveVariants(ps []bigc.Product, blackList *Blacklist) []bigc.Variant {
	var retv []bigc.Variant

	for _, p := range ps {
		for _, v := range p.Variants {
			if !v.PurchasingDisabled && !strings.HasPrefix(v.Sku, "//") {
				retv = append(retv, v)
			}
		}
	}

	return blackList.ApplyForVariants(retv)
}

func handleGetZReport(bc *bigc.Client, service product.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ps, err := bc.GetAllProducts(map[string]string{"include": "variants,images"})
		if err != nil {
			return err
		}
		pIDMap := lo.Associate(ps, func(p bigc.Product) (string, bigc.Product) {
			return strconv.Itoa(p.ID), p
		})

		pps := lo.Filter(ps, func(p bigc.Product, _ int) bool {
			return p.Sku != "" && len(p.Images) > 0 && p.IsVisible && p.InventoryLevel == 0
		})

		var zr Zr
		for _, p := range pps {
			pp := presenter.Product{BCID: strconv.Itoa(p.ID)}
			err := service.FetchProduct(&pp)
			if err != nil {
				log.Println(err)
				continue
			}
			zr.Ls = append(zr.Ls, Zrl{
				Id:         strconv.Itoa(p.ID),
				Sku:        p.Sku,
				Name:       p.Name,
				SohWeb:     p.InventoryLevel,
				SohStores:  pp.StockInformation.Total,
				Price:      p.Price,
				PromoPrice: p.SalePrice,
				IsVariant:  false,
				IsVisible:  p.IsVisible,
				IsDeleted:  strings.HasPrefix(pp.ProdName, "/"),
			})
		}

		vps := lo.Filter(ps, func(p bigc.Product, _ int) bool {
			return p.Sku == "" && len(p.Variants) > 0 && p.IsVisible
		})

		vs := lo.FlatMap(vps, func(p bigc.Product, _ int) []bigc.Variant {
			return p.Variants
		})
		vs = lo.Filter(vs, func(v bigc.Variant, _ int) bool {
			return v.ImageURL != "" && v.InventoryLevel == 0 && !v.PurchasingDisabled
		})

		for _, v := range vs {
			pp := presenter.Product{BCID: fmt.Sprintf("%d/%d", v.ProductID, v.ID)}
			err := service.FetchProduct(&pp)
			if err != nil {
				log.Println(err)
				continue
			}
			var name string
			if len(v.OptionValues) > 0 {
				name = pIDMap[strconv.Itoa(v.ProductID)].Name + v.OptionValues[0].OptionDisplayName
			} else {
				name = pIDMap[strconv.Itoa(v.ProductID)].Name
			}
			zr.Ls = append(zr.Ls, Zrl{
				Id:         fmt.Sprintf("%d/%d", v.ProductID, v.ID),
				Sku:        v.Sku,
				Name:       name,
				SohWeb:     v.InventoryLevel,
				SohStores:  pp.StockInformation.Total,
				Price:      v.Price,
				PromoPrice: v.SalePrice,
				IsVariant:  true,
				IsVisible:  !v.PurchasingDisabled,
				IsDeleted:  strings.HasPrefix(pp.ProdName, "/"),
			})
		}
		return c.JSON(zr)
	}
}

type Zr struct {
	Ls []Zrl
}

type Zrl struct {
	Id         string
	Sku        string
	Name       string
	SohWeb     int
	SohStores  int
	Price      float64
	PromoPrice float64
	IsVariant  bool
	IsVisible  bool
	IsDeleted  bool
}

type BigCommerceTopSalesReport struct {
	Lines []BigCommerceTopSalesReportLine
}
type BigCommerceTopSalesReportLine struct {
	Name         string  `csv:"name"`
	Sku          string  `csv:"product_code_or_sku"`
	Brand        string  `csv:"brand"`
	ProductCode  string  `csv:"product_code"`
	Orders       int     `csv:"orders"`
	Revenue      float64 `csv:"revenue"`
	QuantitySold int     `csv:"qty_sold"`
	Visits       int     `csv:"visits"`
}

type TopSalesReport []TopSalesReportLine
type TopSalesReportLine struct {
	Name         string
	Sku          string
	Brand        string
	Orders       int
	Revenue      float64
	QuantitySold int
	Visits       int

	presenter.StockInformation
}

func handlePostTopSales(service product.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fh, err := c.FormFile("file")
		if err != nil {
			log.Println(err)
			return err
		}

		ff, err := fh.Open()
		if err != nil {
			log.Println(err)
			return err
		}

		pars, err := parser.NewCsvParser(ff, parser.WithComma(','))
		if err != nil {
			panic(err)
		}

		report := BigCommerceTopSalesReport{}
		if err := pars.Parse(&report.Lines); err != nil {
			panic(err)
		}

		skuM1 := lo.Associate(report.Lines, func(l BigCommerceTopSalesReportLine) (string, BigCommerceTopSalesReportLine) {
			return strings.TrimSpace(l.Sku), l
		})
		skuM := make(map[string]BigCommerceTopSalesReportLine)
		for k, v := range skuM1 {
			sku, err := bigc.CleanBigCommerceSku(k)
			if err != nil {
				log.Println(err)
				continue
			}
			skuM[sku] = v
		}

		ps, err := service.FetchProducts(product.WithSkus(lo.Keys(skuM)...), product.WithStockInformation())
		if err != nil {
			return err
		}

		var retv TopSalesReport = lo.FilterMap(ps, func(p *presenter.Product, _ int) (TopSalesReportLine, bool) {
			bcp, ok := skuM[p.Sku]
			if !ok {
				return TopSalesReportLine{}, false
			}
			return TopSalesReportLine{
				StockInformation: p.StockInformation,
				Name:             p.ProdName,
				Sku:              p.Sku,
				Brand:            bcp.Brand,
				Orders:           bcp.Orders,
				Revenue:          bcp.Revenue,
				QuantitySold:     bcp.QuantitySold,
				Visits:           bcp.Visits,
			}, true
		})

		sort.Slice(retv, func(i, j int) bool {
			return retv[i].QuantitySold > retv[j].QuantitySold
		})

		return c.JSON(retv)
	}
}
