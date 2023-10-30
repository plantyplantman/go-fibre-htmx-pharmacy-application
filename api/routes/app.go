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

		// skus := []string{}
		// for k := range rM {
		// 	for _, l := range rM[k].Lines {
		// 		skus = append(skus, l.Sku)
		// 	}
		// }
		// ps, err := service.FetchProducts(product.WithStockInformation(), product.WithSkus(skus...))
		// if err != nil {
		// 	log.Println(err)
		// 	return err
		// }

		data := fiber.Map{
			"New":    rM["new"].ToTable(),
			"Edited": rM["edited"].ToTable(),
			"Clean":  rM["clean"].ToTable(),
		}
		log.Println(data)
		return c.Render("partials/collapsable-table", data, "layout")
	})

	// f.Get("/categories", func(c *fiber.Ctx) error {
	// 	return c.Render("categories", fiber.Map{}, "layout")
	// })
}
