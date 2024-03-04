package bigc

import (
	"fmt"
	"strings"
)

type ProductFile struct {
	Lines []ProductFileLine
}

type ProductFileLine struct {
	Id        string
	Sku       string
	Name      string
	Price     float64
	Soh       int
	IsVariant bool
	ImageURL  string
}

func (pf *ProductFile) Export(path string) error {
	headers, err := GetStructFields(pf.Lines[0])
	if err != nil {
		return err
	}

	var content [][]string
	var tmp []string
	for _, p := range pf.Lines {
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
		if sku == "" {
			for _, v := range p.Variants {
				var name string
				if len(v.OptionValues) > 0 {
					name = p.Name + v.OptionValues[0].OptionDisplayName
				} else {
					name = p.Name
				}
				tmp := ProductFileLine{
					Id:        fmt.Sprintf("%d/%d", p.ID, v.ID),
					Name:      name,
					Sku:       v.Sku,
					Price:     v.Price,
					Soh:       v.InventoryLevel,
					IsVariant: true,
					ImageURL:  v.ImageURL,
				}
				retv.Lines = append(retv.Lines, tmp)
			}
		} else {
			imageUrl := ""
			if len(p.Images) > 0 {
				imageUrl = p.Images[0].ImageURL
			}
			tmp := ProductFileLine{
				Id:        fmt.Sprint(p.ID),
				Name:      p.Name,
				Sku:       p.Sku,
				Price:     p.Price,
				Soh:       p.InventoryLevel,
				IsVariant: false,
				ImageURL:  imageUrl,
			}
			retv.Lines = append(retv.Lines, tmp)
		}
	}
	return &retv, nil
}
