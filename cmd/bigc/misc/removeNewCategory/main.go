package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/samber/lo"
)

func main() {

	path := `C:\Users\admin\Develin Management Dropbox\Zihan\files\in\240221\240221__PRODUCTS-IN-NEW-CATEGORY.csv`
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	bcids := make([]int, 0)
	r := csv.NewReader(f)

	// skip header
	r.Read()
	for {
		record, err := r.Read()
		if err != nil {
			break
		}
		id, err := strconv.Atoi(record[1])
		if err != nil {
			log.Fatalln(err)
		}
		bcids = append(bcids, id)
	}

	bc := bigc.MustGetClient()

	for _, id := range bcids {
		p, err := bc.GetProductById(id)
		if err != nil {
			log.Println(err)
		}

		p.Categories = lo.Filter(p.Categories, func(c int, _ int) bool {
			return c != 1041
		})

		_, err = bc.UpdateProduct(p, bigc.WithUpdateProductCategories(p.Categories))
		if err != nil {
			log.Println(err)
		}
		fmt.Println("removed category for product", id)
	}

}
