package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/plantyplantman/bcapi/api/presenter"
	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/plantyplantman/bcapi/pkg/env"
	"github.com/plantyplantman/bcapi/pkg/product"
	"github.com/samber/lo"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	f, err := os.Open(`C:\Users\admin\Desktop\231107__ms-web__promos-deleted.txt`)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = '\t'
	r.LazyQuotes = true

	skus := make([]string, 0)
	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln(err)
		}

		skus = append(skus, line[0])
	}

	var connString = env.TEST_NEON
	if connString == "" {
		log.Fatalln("TEST_NEON_CONNECTION_STRING not set")
	}
	DB, err := gorm.Open(postgres.Open(connString), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	repo := product.NewRepository(DB)
	service := product.NewService(repo)
	c := bigc.MustGetClient()

	ps, err := service.FetchProducts(product.WithSkus(skus...))
	if err != nil {
		log.Fatalln(err)
	}

	pBCIDs := lo.FilterMap(ps, func(p *presenter.Product, _ int) (string, bool) {
		tokens := strings.Split(p.BCID, "/")
		return tokens[0], len(tokens) == 1 && tokens[0] != ""
	})

	var bcPs []bigc.Product
	pBCIDs2 := splitSlice(pBCIDs, 100)
	for _, ids := range pBCIDs2 {
		tmp, err := c.GetAllProducts(
			map[string]string{
				"id:in": "[" + lo.Reduce(
					ids, func(agg string, item string, _ int) string {
						return agg + "," + item
					}, "") + "]"})
		if err != nil {
			log.Fatalln(err)
		}
		bcPs = append(bcPs, tmp...)
	}

	for _, p := range bcPs {
		sku := p.Sku
		if !strings.HasPrefix(sku, "/") {
			sku = "/" + sku
		}
		if !strings.HasPrefix(sku, "//") {
			sku = "/" + sku
		}
		_, err = c.UpdateProduct(&p,
			bigc.WithUpdateProductCategoriesWithoutSaleIDs(p.Categories),
			bigc.WithUpdateProductSalePrice(0),
			bigc.WithUpdateProductIsVisible(false),
			bigc.WithUpdateProductCategoryIsRetired(true),
			bigc.WithUpdateProductSku(sku),
			bigc.WithUpdateProductInventoryLevel(0),
		)
		if err != nil {
			log.Println(err)
			continue
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
		v, err := c.GetVariantById(bcid.vId, bcid.pId, map[string]string{})
		sku := v.Sku
		if !strings.HasPrefix(sku, "/") {
			sku = "/" + sku
		}
		if !strings.HasPrefix(sku, "//") {
			sku = "/" + sku
		}

		if err != nil {
			log.Fatalln(err)
		}
		_, err = c.UpdateVariant(v,
			bigc.WithUpdateVariantSalePrice(0),
			bigc.WithUpdateVariantPurchasingDisabled(true),
			bigc.WithUpdateVariantSku(sku),
			bigc.WithUpdateVariantInventoryLevel(0))
		if err != nil {
			log.Println(err)
			continue
		}
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
