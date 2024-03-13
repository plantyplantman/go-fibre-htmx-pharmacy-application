package bigc

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gocarina/gocsv"
	"github.com/samber/lo"
)

type ProductFile struct {
	Lines []ProductFileLine
}

type ProductFileLine struct {
	Id         string
	Sku        string
	Name       string
	Price      float64
	SalePrice  float64
	Soh        int
	IsVariant  bool
	ImageURL   string
	Categories []int
}

func (pf *ProductFile) Export(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	gocsv.SetCSVWriter(func(w io.Writer) *gocsv.SafeCSVWriter {
		csvw := csv.NewWriter(w)
		csvw.Comma = '\t'
		return gocsv.NewSafeCSVWriter(csvw)
	})
	return gocsv.MarshalFile(&pf.Lines, f)
}

func (pf *ProductFile) GetLineBySku(sku string) (*ProductFileLine, error) {
	sku = strings.TrimSpace(sku)
	for _, l := range pf.Lines {
		if l.Sku == "" || l.Sku == "/" {
			continue
		}
		lsku, err := CleanBigCommerceSku(l.Sku)
		if err != nil {
			if strings.HasPrefix(err.Error(), "Invalid sku") {
				continue
			}
			return nil, err
		}

		if strings.HasPrefix(lsku, sku) || strings.HasSuffix(lsku, sku) {
			return &l, nil
		}
	}
	return nil, &ProductNotFoundError{
		Sku:    sku,
		Source: "Product File",
	}
}

func (c *Client) GetProductFile() (*ProductFile, error) {
	ps, err := c.GetAllProducts(map[string]string{"include": "variants,images"})
	if err != nil {
		return nil, err
	}

	var retv ProductFile
	for _, p := range ps {
		sku := strings.TrimSpace(p.Sku)
		if strings.Contains(sku, "-") || strings.Contains(sku, "Best Before") || strings.Contains(sku, "Expiry") {
			continue
		}
		cats := lo.Keys(lo.Associate(p.Categories, func(c int) (int, struct{}) {
			return c, struct{}{}
		}))
		if len(p.Variants) > 1 {
			for _, v := range p.Variants {
				var name string
				if len(v.OptionValues) > 0 {
					name = p.Name + v.OptionValues[0].OptionDisplayName
				} else {
					name = p.Name
				}
				var imageUrl string
				if strings.TrimSpace(v.ImageURL) == "" && len(p.Variants) > 0 {
					imageUrl = p.Variants[0].ImageURL
				} else {
					imageUrl = v.ImageURL
				}

				tmp := ProductFileLine{
					Id:         fmt.Sprintf("%d/%d", p.ID, v.ID),
					Name:       name,
					Sku:        v.Sku,
					Price:      v.Price,
					SalePrice:  v.SalePrice,
					Soh:        v.InventoryLevel,
					IsVariant:  true,
					ImageURL:   imageUrl,
					Categories: cats,
				}
				retv.Lines = append(retv.Lines, tmp)
			}
		} else {
			imageUrl := ""
			if len(p.Variants) > 0 {
				imageUrl = p.Variants[0].ImageURL
			}
			tmp := ProductFileLine{
				Id:         fmt.Sprint(p.ID),
				Name:       p.Name,
				Sku:        p.Sku,
				Price:      p.Price,
				SalePrice:  p.SalePrice,
				Soh:        p.InventoryLevel,
				IsVariant:  false,
				ImageURL:   imageUrl,
				Categories: cats,
			}
			retv.Lines = append(retv.Lines, tmp)
		}
	}
	return &retv, nil
}
