package main

import (
	"os"
	"slices"
	"strings"

	"github.com/plantyplantman/bcapi/pkg/bigc"
)

func main() {
	run()
}

func run() error {
	c := bigc.MustGetClient()
	ps, err := c.GetAllProducts(map[string]string{"include": "images,variants"})
	if err != nil {
		return err
	}

	f, err := os.Create("weirdos.tsv")
	if err != nil {
		return err
	}
	defer f.Close()

	for _, p := range ps {
		if strings.HasPrefix(p.Sku, "//") || !p.IsVisible || slices.Contains(p.Categories, bigc.RETIRED_PRODUCTS) {
			var line = []string{p.Sku, p.Name}
			if p.IsVisible {
				line = append(line, "visible")
			}

			if p.InventoryLevel > 0 {
				line = append(line, "soh>0")
			}

			if !slices.Contains(p.Categories, bigc.RETIRED_PRODUCTS) {
				line = append(line, "!retired")
			}

			if !strings.HasPrefix(p.Sku, "//") {
				line = append(line, "!//")
			}
			if len(p.Images) == 0 || len(line) == 2 {
				continue
			}
			_, err := f.WriteString(strings.Join(line, "\t") + "\n")
			if err != nil {
				return err
			}
		} else if p.Sku == "" {
			for _, v := range p.Variants {
				if strings.HasPrefix(v.Sku, "//") || v.PurchasingDisabled {
					var line = []string{v.Sku, p.Name + " " + v.OptionValues[0].OptionDisplayName}
					if !v.PurchasingDisabled {
						line = append(line, "visible")
					}
					if v.InventoryLevel > 0 {
						line = append(line, "soh>0")
					}

					if !strings.HasPrefix(p.Sku, "//") {
						line = append(line, "!//")
					}

					if v.ImageURL == "" || len(line) == 2 {
						continue
					}
					_, err := f.WriteString(strings.Join(line, "\t") + "\n")
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}
