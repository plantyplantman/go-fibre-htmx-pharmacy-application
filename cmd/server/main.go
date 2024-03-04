package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"github.com/plantyplantman/bcapi/api/routes"
	"github.com/plantyplantman/bcapi/pkg/env"
	"github.com/plantyplantman/bcapi/pkg/product"
	"github.com/plantyplantman/bcapi/pkg/sales"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	var (
		db             *gorm.DB
		repo           product.Repository
		productService product.Service
		err            error
	)
	if db, err = initDB(); err != nil {
		log.Fatal(err)
	}

	repo = product.NewRepository(db)
	productService = product.NewService(repo)
	salesService, err := sales.NewService(4302, "https://localhost:44350")
	if err != nil {
		log.Fatalln(err)
	}

	engine := html.New(`.\public\views\`, ".tpl.html")
	engine.AddFunc("add1", func(a int) int {
		return a + 1
	})

	f := fiber.New(fiber.Config{
		Views: engine,
	})
	f.Use(cors.New())
	f.Use(recover.New())
	f.Use(cache.New())

	f.Static("/", "./public")
	routes.AppRouter(f, productService)

	api := f.Group("/api")
	routes.ProductRouter(api, productService)
	routes.SalesRouter(api, salesService)

	log.Fatal(f.Listen(":8080"))
}

func initDB() (*gorm.DB, error) {
	if env.NEON == "" {
		log.Fatalln("NEON_CONNECTION_STRING not set")
	}
	return gorm.Open(postgres.Open(env.TEST_NEON), &gorm.Config{})
}
