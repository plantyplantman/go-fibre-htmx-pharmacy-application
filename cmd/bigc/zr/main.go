package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/plantyplantman/bcapi/api/presenter"
	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/plantyplantman/bcapi/pkg/parser"
	"github.com/plantyplantman/bcapi/pkg/product"
	"github.com/samber/lo"
)

type Zr struct {
	Ls []Zrl
}

type Zrl struct {
	Id         string
	Sku        string
	Name       string
	SohWeb     int
	SohStores  int
	Price      float64
	PromoPrice float64
	IsVariant  bool
	IsVisible  bool
	IsDeleted  bool
}

func (zr *Zr) Export(path string) error {
	headers, err := GetStructFields(zr.Ls[0])
	if err != nil {
		return err
	}

	var content [][]string
	var tmp []string
	for _, p := range zr.Ls {
		values, err := GetStructValues(p)
		if err != nil {
			return err
		}
		for _, v := range values {
			tmp = append(tmp, fmt.Sprint(v))
		}
		content = append(content, tmp)
		tmp = []string{}
	}

	if err = WriteToTsv(path, headers, content); err != nil {
		return err
	}
	return nil
}

func main() {
	gen2()
}

func update() {
	var date = time.Now().Format("060102")
	path := filepath.Join(`C:\Users\admin\Develin Management Dropbox\Zihan\files\out\`, date, date+`__web__zr.tsv`)
	f, err := os.Open(path)
	if err != nil {
		log.Fatalln(err)
	}

	parser, err := parser.NewCsvParser(f)
	if err != nil {
		log.Fatalln(err)
	}

	var zr Zr
	if err = parser.Parse(&zr.Ls); err != nil {
		log.Fatalln(err)
	}

	c := bigc.MustGetClient()
	for _, zrl := range zr.Ls {
		isRetired := !zrl.IsVisible
		if zrl.IsDeleted && zrl.SohStores == 0 {
			isRetired = true
		}

		if zrl.IsVariant {
			ids := strings.Split(zrl.Id, "/")
			vid, err := strconv.Atoi(ids[1])
			if err != nil {
				log.Println(err)
				continue
			}
			pid, err := strconv.Atoi(ids[0])
			if err != nil {
				log.Println(err)
				continue
			}

			v, err := c.GetVariantById(vid, pid, map[string]string{})
			if err != nil {
				log.Println(err)
				continue
			}
			if isRetired {
				sku := v.Sku
				if !strings.HasPrefix(sku, "/") {
					sku = "/" + sku
				}
				if !strings.HasPrefix(sku, "//") {
					sku = "/" + sku
				}
				_, err = c.UpdateVariant(v,
					bigc.WithUpdateVariantPurchasingDisabled(true),
					bigc.WithUpdateVariantInventoryLevel(zrl.SohStores),
					bigc.WithUpdateVariantSku(sku))
				if err != nil {
					log.Println(err)
					continue
				}

			} else {
				_, err = c.UpdateVariant(v, bigc.WithUpdateVariantInventoryLevel(zrl.SohStores))
				if err != nil {
					log.Println(err)
					continue
				}
			}
		} else {
			id, err := strconv.Atoi(zrl.Id)
			if err != nil {
				log.Println(err)
				continue
			}
			p, err := c.GetProductById(id)
			if err != nil {
				log.Println(err)
				continue
			}
			if isRetired {
				sku := p.Sku
				if !strings.HasPrefix(sku, "/") {
					sku = "/" + sku
				}
				if !strings.HasPrefix(sku, "//") {
					sku = "/" + sku
				}
				_, err = c.UpdateProduct(p,
					bigc.WithUpdateProductIsVisible(false),
					bigc.WithUpdateProductInventoryLevel(zrl.SohStores),
					bigc.WithUpdateProductSku(sku),
					bigc.WithUpdateProductCategoryIsRetired(true),
					bigc.WithUpdateProductCategoriesWithoutSaleIDs(p.Categories))
				if err != nil {
					log.Println(err)
					continue
				}
			} else {
				_, err = c.UpdateProduct(p, bigc.WithUpdateProductInventoryLevel(zrl.SohStores))
				if err != nil {
					log.Println(err)
					continue
				}
			}
		}
	}
}

func gen2() error {
	var date = time.Now().Format("060102")
	path := filepath.Join(`C:\Users\admin\Develin Management Dropbox\Zihan\files\out\`, date, date+`__web__zr.tsv`)
	if err := os.MkdirAll(filepath.Dir(path), 0770); err != nil {
		log.Fatalln(err)
	}

	c := bigc.MustGetClient()
	service, err := product.NewDefaultService()
	if err != nil {
		log.Fatalln(err)
	}

	ps, err := c.GetAllProducts(map[string]string{"include": "variants,images"})
	if err != nil {
		log.Fatalln(err)
	}
	pIDMap := lo.Associate(ps, func(p bigc.Product) (string, bigc.Product) {
		return strconv.Itoa(p.ID), p
	})

	pps := lo.Filter(ps, func(p bigc.Product, _ int) bool {
		return p.Sku != "" && len(p.Images) > 0 && p.IsVisible && p.InventoryLevel == 0
	})

	var zr Zr
	for _, p := range pps {
		pp := presenter.Product{BCID: strconv.Itoa(p.ID)}
		err := service.FetchProduct(&pp)
		if err != nil {
			log.Println(err)
			continue
		}
		zr.Ls = append(zr.Ls, Zrl{
			Id:         strconv.Itoa(p.ID),
			Sku:        p.Sku,
			Name:       p.Name,
			SohWeb:     p.InventoryLevel,
			SohStores:  pp.StockInformation.Total,
			Price:      p.Price,
			PromoPrice: p.SalePrice,
			IsVariant:  false,
			IsVisible:  p.IsVisible,
			IsDeleted:  strings.HasPrefix(pp.ProdName, "/"),
		})
	}

	vps := lo.Filter(ps, func(p bigc.Product, _ int) bool {
		return p.Sku == "" && len(p.Variants) > 0 && p.IsVisible
	})

	vs := lo.FlatMap(vps, func(p bigc.Product, _ int) []bigc.Variant {
		return p.Variants
	})
	vs = lo.Filter(vs, func(v bigc.Variant, _ int) bool {
		return v.ImageURL != "" && v.InventoryLevel == 0 && !v.PurchasingDisabled
	})

	for _, v := range vs {
		pp := presenter.Product{BCID: fmt.Sprintf("%d/%d", v.ProductID, v.ID)}
		err := service.FetchProduct(&pp)
		if err != nil {
			log.Println(err)
			continue
		}
		var name string
		if len(v.OptionValues) > 0 {
			name = pIDMap[strconv.Itoa(v.ProductID)].Name + v.OptionValues[0].OptionDisplayName
		} else {
			name = pIDMap[strconv.Itoa(v.ProductID)].Name
		}
		zr.Ls = append(zr.Ls, Zrl{
			Id:         fmt.Sprintf("%d/%d", v.ProductID, v.ID),
			Sku:        v.Sku,
			Name:       name,
			SohWeb:     v.InventoryLevel,
			SohStores:  pp.StockInformation.Total,
			Price:      v.Price,
			PromoPrice: v.SalePrice,
			IsVariant:  true,
			IsVisible:  !v.PurchasingDisabled,
			IsDeleted:  strings.HasPrefix(pp.ProdName, "/"),
		})
	}
	return zr.Export(path)
}

func gen() {
	var date = time.Now().Format("060102")
	path := filepath.Join(`C:\Users\admin\Develin Management Dropbox\Zihan\files\out\`, date, date+`__web__zr.tsv`)
	if err := os.MkdirAll(filepath.Dir(path), 0770); err != nil {
		log.Fatalln(err)
	}

	service, err := product.NewDefaultService()
	if err != nil {
		log.Fatalln(err)
	}

	c := bigc.MustGetClient()
	ps, err := c.GetAllProducts(map[string]string{"include": "variants,images"})
	if err != nil {
		log.Fatalln(err)
	}

	eps, err := service.FetchProducts(product.WithSkus(
		lo.Map(ps, func(p bigc.Product, _ int) string {
			sku, err := bigc.CleanBigCommerceSku(p.Sku)
			if err != nil {
				log.Println(err, " sku:", p.Sku)
				return p.Sku
			}
			return sku
		})...,
	))
	if err != nil {
		log.Fatalln(err)
	}

	epsM := lo.Associate(eps, func(p *presenter.Product) (string, *presenter.Product) {
		return p.BCID, p
	})

	var zr Zr
	for _, p := range ps {
		if p.Sku == "" {
			for _, v := range p.Variants {
				if v.ImageURL == "" {
					continue
				}
				ep, ok := epsM[fmt.Sprintf("%d/%d", p.ID, v.ID)]
				if !ok {
					log.Println("Product not found in service: ", v.Sku)
				}
				soh := -1
				if ep != nil {
					soh = ep.StockInformation.Total
				}

				var name string
				if len(v.OptionValues) > 0 {
					name = p.Name + v.OptionValues[0].OptionDisplayName
				} else {
					name = p.Name
				}
				tmp := Zrl{
					Id:         fmt.Sprintf("%d/%d", p.ID, v.ID),
					Sku:        v.Sku,
					Name:       name,
					SohWeb:     v.InventoryLevel,
					SohStores:  soh,
					Price:      v.Price,
					PromoPrice: v.SalePrice,
					IsVariant:  true,
					IsVisible:  !v.PurchasingDisabled,
				}
				zr.Ls = append(zr.Ls, tmp)
			}
		} else {
			if len(p.Images) == 0 {
				continue
			}
			ep, ok := epsM[strconv.Itoa(p.ID)]
			if !ok {
				log.Println("Product not found in service:", p.Sku)
			}
			soh := -1
			if ep != nil {
				soh = ep.StockInformation.Total
			}
			zr.Ls = append(zr.Ls, Zrl{
				Id:         strconv.Itoa(p.ID),
				Sku:        p.Sku,
				Name:       p.Name,
				SohWeb:     p.InventoryLevel,
				SohStores:  soh,
				Price:      p.Price,
				PromoPrice: p.SalePrice,
				IsVariant:  false,
				IsVisible:  p.IsVisible,
			})
		}
	}

	if err := zr.Export(path); err != nil {
		log.Fatalln(err)
	}
}

func WriteToTsv(path string, headers []string, content [][]string) error {
	f, e := os.Create(path)
	if e != nil {
		fmt.Println(e)
	}
	w := csv.NewWriter(f)
	w.Comma = '\t'
	defer w.Flush()

	e = w.Write(headers)
	if e != nil {
		fmt.Println(e)
	}

	for _, row := range content {
		w.Write(row)
	}
	return e
}

func GetStructFields(s any) ([]string, error) {
	if s == nil {
		return nil, errors.New("NO FIELD FOR A NIL DIMWIT")
	}
	v := reflect.ValueOf(s)
	var r []string
	for i := 0; i < v.NumField(); i++ {
		r = append(r, v.Type().Field(i).Name)
	}
	return r, nil
}

func GetStructValues(s any) ([]reflect.Value, error) {
	if s == nil {
		return nil, errors.New("NO FIELD FOR A NIL DIMWIT")
	}
	v := reflect.ValueOf(s)
	var r []reflect.Value
	for i := 0; i < v.NumField(); i++ {
		r = append(r, v.Field(i))
	}
	return r, nil
}
