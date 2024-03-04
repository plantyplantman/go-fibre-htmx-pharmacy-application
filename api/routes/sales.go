package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/plantyplantman/bcapi/api/handlers"
	"github.com/plantyplantman/bcapi/pkg/sales"
)

func SalesRouter(f fiber.Router, s sales.Service) {
	f.Get("/sales", handlers.GetSales("https://localhost:44350"))
}
