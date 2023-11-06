package routes

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/plantyplantman/bcapi/api/presenter"
	"github.com/plantyplantman/bcapi/pkg/parser"
	"github.com/plantyplantman/bcapi/pkg/product"
	"github.com/plantyplantman/bcapi/pkg/report"
	"github.com/samber/lo"
	"golang.org/x/exp/rand"
)

func AppRouter(f fiber.Router, service product.Service) {
	f.Get("/", func(ctx *fiber.Ctx) error {
		// Render index within layouts/main
		return ctx.Render("partials/hero", fiber.Map{}, "layout")
	})

	f.Get("/products", func(ctx *fiber.Ctx) error {
		var (
			ps  presenter.Products
			err error
		)
		if ps, err = service.FetchProducts(
			product.WithOrderBy(ctx),
			product.WithPaginate(ctx),
			product.WithQuery(ctx),
			product.WithStockInformation(),
			product.WithDeleted(ctx),
			product.WithStockedOnWeb(ctx),
		); err != nil {
			return err
		}
		page := ctx.QueryInt("page", 1)
		log.Println(ps.ToTable(page+1, ctx.QueryInt("limit", 50)))
		return ctx.Render("products",
			ps.ToTable(page+1, ctx.QueryInt("limit", 50)),
			"layout")
	})

	f.Get("/multistore", func(c *fiber.Ctx) error {
		return c.Render("multistore", fiber.Map{}, "layout")
	})

	f.Post("/multistore", func(c *fiber.Ctx) error {
		fh, err := c.FormFile("file")
		if err != nil {
			log.Println(err)
			return err
		}
		path := "./tmp/" + fmt.Sprint(rand.Uint64()) + fh.Filename
		if err = c.SaveFile(fh, path); err != nil {
			log.Println(err)
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			log.Println(err)
			return err
		}
		defer f.Close()

		p, err := parser.NewParser(f, parser.IsMultistore(true))
		if err != nil {
			log.Println(err)
			return err
		}

		var rM = map[string]*report.ProductRetailList{}
		if err := p.Parse(&rM); err != nil {
			log.Println(err)
			return err
		}

		skus := []string{}
		for k := range rM {
			for _, l := range rM[k].Lines {
				skus = append(skus, l.Sku)
			}
		}
		ps, err := service.FetchProducts(product.WithSkus(skus...))
		if err != nil {
			log.Println(err)
			return err
		}

		// products that are not on the website

		psSkuM := map[string]*presenter.Product{}
		for _, p := range lo.Filter(ps, func(p *presenter.Product, _ int) bool {
			return p.OnWeb == 1
		}) {
			psSkuM[p.Sku] = p
		}
		newps := []*presenter.Product{}
		changedps := []*presenter.Product{}
		nochangeps := []*presenter.Product{}
		for k := range rM {
			for _, l := range rM[k].Lines {
				if _, ok := psSkuM[l.Sku]; !ok {
					newps = append(newps, &presenter.Product{
						Sku:       l.Sku,
						ProdName:  l.ProdName,
						Price:     l.Price.Float64(),
						CostPrice: l.Cost.Float64(),
						OnWeb:     0,
						IsVariant: false,
						BCID:      "",
						StockInfomation: presenter.StockInformation{
							Petrie:   0,
							Bunda:    0,
							Con:      0,
							Franklin: 0,
							Web:      0,
							Total:    0,
						},
					})
				} else if psSkuM[l.Sku].Price != l.Price.Float64() {
					changedps = append(changedps, psSkuM[l.Sku])
				} else {
					nochangeps = append(nochangeps, psSkuM[l.Sku])
				}
			}
		}

		newData := []presenter.Row{}
		for _, p := range newps {
			newData = append(newData, p.ToPresenterRow())
		}

		changedData := []presenter.Row{}
		for _, p := range changedps {
			changedData = append(changedData, p.ToPresenterRow())
		}

		nochangeData := []presenter.Row{}
		for _, p := range nochangeps {
			nochangeData = append(nochangeData, p.ToPresenterRow())
		}

		data := fiber.Map{
			"New": presenter.Table{
				Headers: []string{
					"Sku",
					"Name",
					"Price",
					"Cost Price",
					"Petrie",
					"Bunda",
					"Con",
					"Franklin",
					"Web",
					"Total",
					"BCID",
				},
				Rows: newData,
			},
			"Changed": presenter.Table{

				Headers: []string{
					"Sku",
					"Name",
					"Price",
					"Cost Price",
					"Petrie",
					"Bunda",
					"Con",
					"Franklin",
					"Web",
					"Total",
					"BCID",
				},
				Rows: changedData,
			},
			"NoChange": presenter.Table{

				Headers: []string{
					"Sku",
					"Name",
					"Price",
					"Cost Price",
					"Petrie",
					"Bunda",
					"Con",
					"Franklin",
					"Web",
					"Total",
					"BCID",
				},
				Rows: nochangeData,
			},
		}

		log.Println(data)
		return c.Render("partials/collapsable-table", data, "layout")
	})

	// f.Get("/categories", func(c *fiber.Ctx) error {
	// 	return c.Render("categories", fiber.Map{}, "layout")
	// })
}
