package db_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/plantyplantman/bcapi/pkg/entities"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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
