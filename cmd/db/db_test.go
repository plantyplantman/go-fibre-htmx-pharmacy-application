package db_test

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/plantyplantman/bcapi/pkg/db"
	"github.com/plantyplantman/bcapi/pkg/entities"
	"github.com/plantyplantman/bcapi/pkg/env"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestMigrate(t *testing.T) {

	var connString = env.TEST_NEON
	if connString == "" {
		log.Fatalln("TEST_NEON_CONNECTION_STRING not set")
	}
	DB, err := gorm.Open(postgres.Open(connString), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	if err = db.Migrate(DB); err != nil {
		log.Fatalln(err)
	}
}

func TestInsertSale(t *testing.T) {
	// var connString = env.TEST_NEON
	// if connString == "" {
	// 	log.Fatalln("TEST_NEON_CONNECTION_STRING not set")
	// }
	// DB, err := gorm.Open(postgres.Open(connString), &gorm.Config{})
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	url := func(from string, to string) string {
		return fmt.Sprintf("https://localhost:44350/api/4302/sales?from=%s&to=%s", from, to)
	}

	// from=1999-12-31&to=2023-11-14

	totalChan := make(chan float64)

	getSales := func(date string) {
		resp, err := http.Get(url(date, date))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		var sales []entities.Sale
		if err := json.Unmarshal(body, &sales); err != nil {
			t.Fatal(err, string(body))
		}

		for _, sale := range sales {
			totalChan <- sale.Total
		}
	}

	from, err := time.Parse("2006-01-02", "1999-12-31")
	if err != nil {
		t.Fatal(err)
	}

	to, err := time.Parse("2006-01-02", "2023-11-14")
	if err != nil {
		t.Fatal(err)
	}

	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		go getSales(d.Format("2006-01-02"))
	}

	total := 0.0
	for range time.Tick(time.Millisecond * 100) {
		select {
		case t := <-totalChan:
			total += t
		default:
			if time.Now().After(to) {
				break
			}
		}
	}

	fmt.Println("Total sales: ", total)
	// DB.Create(&sales)

	// for _, sale := range sales {
	// 	fmt.Printf("%+v\n", sale)
	// 	// DB.FirstOrCreate(&entities.Sale{}, entities.Sale{SaleID: sale.SaleID})
	// }
}

func TestGet(t *testing.T) {
	db, err := gorm.Open(postgres.Open("postgres://plantyplantman:QA4rI2PmNbDS@ep-delicate-scene-07511672.ap-southeast-1.aws.neon.tech/test"), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}

	var products []entities.Product
	// if err := db.Where("prod_name LIKE ?", "%r%").Find(&products).Error; err != nil {
	// 	t.Fatal(err)
	// }
	// if len(products) == 0 {
	// 	t.Fatal("no products found")
	// }

	db.Preload("stock_informations").Find(&products, entities.Product{Sku: "3616303424237"})

	fmt.Println("len products: ", len(products))
	for _, p := range products {
		fmt.Printf("%+v\n", p)
	}

}
