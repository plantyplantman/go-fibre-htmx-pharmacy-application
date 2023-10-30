package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/plantyplantman/bcapi/api/handlers"
	"github.com/plantyplantman/bcapi/pkg/product"
)

func ProductRouter(f fiber.Router, service product.Service) {
	f.Get("/", func(c *fiber.Ctx) error {
		return c.Send([]byte("TPN DEV API"))
	})
	f.Get("/product/:sku", handlers.GetProduct(service))
	f.Get("/products", handlers.GetProducts(service))
}
