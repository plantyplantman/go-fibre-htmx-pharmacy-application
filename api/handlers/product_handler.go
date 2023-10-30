package handlers

import (
	"errors"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/plantyplantman/bcapi/api/presenter"
	"github.com/plantyplantman/bcapi/pkg/product"
)

func GetProduct(service product.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sku := c.Params("sku")
		if sku == "" {
			c.Status(http.StatusBadRequest)
			return c.JSON(
				presenter.ProductErrorResponse(
					errors.New("please specify sku")))
		}
		var p = presenter.Product{Sku: sku}
		if err := service.FetchProduct(&p); err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.ProductErrorResponse(err))
		}

		return c.JSON(presenter.ProductSuccessResponse(&p))
	}
}

func GetProducts(service product.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var (
			ps  []*presenter.Product
			err error
		)
		if ps, err = service.FetchProducts(
			product.WithQuery(c),
			product.WithPaginate(c),
			product.WithOrderBy(c),
			product.WithStockInformation(),
			product.WithDeleted(c),
			product.WithStockedOnWeb(c),
		); err != nil {
			c.Status(http.StatusInternalServerError)
			return c.JSON(presenter.ProductErrorResponse(err))
		}

		return c.JSON(presenter.ProductsSuccessResponse(ps))
	}
}
