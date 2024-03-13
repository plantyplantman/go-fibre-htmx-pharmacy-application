package main

import (
	"log"
	"slices"
	"strconv"
	"strings"

	"github.com/plantyplantman/bcapi/api/presenter"
	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/plantyplantman/bcapi/pkg/product"
	"github.com/samber/lo"
)

type Blacklist struct {
	skus     []string
	brands   []string
	brandIDs map[int]struct{}
}

var blacklist = Blacklist{
	skus: []string{
		"8710447262696",   // Badedas Classic Vital Bath and Shower Gel 500ml
		"8710447262696-3", // Badedas Classic Vital Bath Gel 3 Pack of 500ml Bottles
		// covid
		"754590335515",  // V Check
		"6970277518482", // JusCheck x5
		"9309000566020", // TouchBio x2
		"6970277517843", // Alltest
		"6974408020240", // Fantest
		"9310059063316", // lyclear
		"9556258003290", //coco scalp
	},
	brands: []string{
		"Instant Smile",
		"Xylimelt",
		"Blephadex",
		"Osmo",
		"Plain",
		"Disprin",
		"Solprin",
		"Aspro",
		"Panadol",
		"Badedas",
		"Gaviscon",
		"Badedas",
		"Expiry",
		"Best Before",
	},
	brandIDs: map[int]struct{}{
		542: {}, // Flawless
	},
}

func (bl *Blacklist) ApplyForProducts(ps []bigc.Product) []bigc.Product {
	return lo.FilterMap(ps, func(item bigc.Product, index int) (bigc.Product, bool) {
		if strings.Contains(item.Sku, "-") {
			return item, false
		}

		sku, err := bigc.CleanBigCommerceSku(item.Sku)
		if err != nil {
			return item, false
		}

		if slices.Contains(bl.skus, sku) {
			return item, false
		}

		for _, b := range bl.brands {
			x, y := strings.ToUpper(strings.TrimSpace(item.Name)), strings.ToUpper(strings.TrimSpace(b))
			if strings.Contains(x, y) {
				return item, false
			}
		}

		if _, ok := bl.brandIDs[item.BrandID]; ok {
			return item, false
		}

		return item, true
	})
}

func (bl *Blacklist) ApplyForVariants(vs []bigc.Variant) []bigc.Variant {
	return lo.FilterMap(vs, func(item bigc.Variant, index int) (bigc.Variant, bool) {
		sku, err := bigc.CleanBigCommerceSku(item.Sku)
		if err != nil {
			return item, false
		}

		if slices.Contains(bl.skus, sku) {
			return item, false
		}

		return item, true
	})
}

func main() {
	run()
}

func run() {
	c := bigc.MustGetClient()

	service, err := product.NewDefaultService()
	if err != nil {
		log.Fatalln(err)
	}

	data, err := c.GetAllProducts(map[string]string{"is_visible": "true"})
	if err != nil {
		log.Fatalln(err)
	}

	ps := filterForActiveProducts(data)
	pids := getBCIDsForProducts(ps)

	ppsMap, err := fetchProductsFromServiceByBCIDs(service, pids)
	if err != nil {
		log.Fatalln(err)
	}

	updateProducts(ps, ppsMap, c)

	vs := filterForActiveVariants(data)
	vids := getBCIDsForVariants(vs)
	vpsMap, err := fetchVariantsFromServiceByBCIDs(service, vids)
	if err != nil {
		log.Fatalln(err)
	}
	updateVariants(vs, vpsMap, c)
}

func updateVariants(vs []bigc.Variant, vpsMap map[string]*presenter.Product, c *bigc.Client) {
	for _, v := range vs {
		if strings.HasPrefix(v.Sku, "//") || v.PurchasingDisabled {

			sku := v.Sku
			if !strings.HasPrefix(sku, "//") {
				if !strings.HasPrefix(sku, "/") {
					sku = "/" + sku
				}
				sku = "/" + sku
			}
			_, err := c.UpdateVariant(&v,
				bigc.WithUpdateVariantInventoryLevel(0),
				bigc.WithUpdateVariantSalePrice(0),
				bigc.WithUpdateVariantSku(sku),
			)
			if err != nil {
				log.Println(err)
			}
			continue
		}
		if vp, ok := vpsMap[strconv.Itoa(v.ProductID)+"/"+strconv.Itoa(v.ID)]; ok {
			updateFns := []bigc.UpdateVariantOpt{}

			// update price
			if vp.Price != v.Price {
				updateFns = append(updateFns, bigc.WithUpdateVariantPrice(vp.Price))
			}

			// update stock
			if vp.StockInformation.Total != v.InventoryLevel {
				updateFns = append(updateFns, bigc.WithUpdateVariantInventoryLevel(vp.StockInformation.Total))
			}

			// update sku
			if strings.HasPrefix(vp.ProdName, "\\") {
				sku := v.Sku
				if !strings.HasPrefix(sku, "/") {
					sku = "/" + sku
				}

				if vp.StockInformation.Total == 0 {
					if !strings.HasPrefix(sku, "//") {
						sku = "/" + sku
					}
					updateFns = append(updateFns,
						bigc.WithUpdateVariantPurchasingDisabled(true),
						bigc.WithUpdateVariantInventoryLevel(0))
				}

				if sku != v.Sku {
					updateFns = append(updateFns, bigc.WithUpdateVariantSku(sku))
				}
			}

			if len(updateFns) == 0 {
				continue
			}
			_, err := c.UpdateVariant(&v, updateFns...)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}

func updateProducts(ps []bigc.Product, ppsMap map[string]*presenter.Product, c *bigc.Client) {
	for _, p := range ps {
		if strings.HasPrefix(p.Sku, "//") || !p.IsVisible || slices.Contains(p.Categories, 1230) {
			sku := p.Sku
			if !strings.HasPrefix(sku, "//") {
				if !strings.HasPrefix(sku, "/") {
					sku = "/" + sku
				}
				sku = "/" + sku
			}
			_, err := c.UpdateProduct(&p,
				bigc.WithUpdateProductInventoryLevel(0),
				bigc.WithUpdateProductCategories(bigc.RemoveSaleCategories(p.Categories)),
				bigc.WithUpdateProductIsVisible(false),
				bigc.WithUpdateProductCategoryIsRetired(true),
				bigc.WithUpdateProductSku(sku))
			if err != nil {
				log.Println(err)
			}
			continue
		}
		if pp, ok := ppsMap[strconv.Itoa(p.ID)]; ok {
			updateFns := []bigc.ProductUpdateOptFn{}

			if pp.Price != p.Price {
				updateFns = append(updateFns, bigc.WithUpdateProductPrice(pp.Price))
			}

			if pp.StockInformation.Total != p.InventoryLevel {
				updateFns = append(updateFns, bigc.WithUpdateProductInventoryLevel(pp.StockInformation.Total))
			}

			if strings.HasPrefix(pp.ProdName, "\\") {
				sku := p.Sku
				if !strings.HasPrefix(sku, "/") {
					sku = "/" + sku
				}

				if pp.StockInformation.Total == 0 {
					sku = "/" + sku
					if !strings.HasPrefix(sku, "//") {
						sku = "/" + sku
					}
					updateFns = append(updateFns,
						bigc.WithUpdateProductIsVisible(false),
						bigc.WithUpdateProductCategoryIsRetired(true),
						bigc.WithUpdateProductInventoryLevel(0))
				}

				if sku != p.Sku {
					updateFns = append(updateFns, bigc.WithUpdateProductSku(sku))
				}
			}

			if len(updateFns) == 0 {
				continue
			}
			_, err := c.UpdateProduct(&p, updateFns...)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}

func fetchVariantsFromServiceByBCIDs(service product.Service, vids []string) (map[string]*presenter.Product, error) {
	pps, err := service.FetchProducts(product.WithBCIDs(vids...), product.WithStockInformation())
	if err != nil {
		return nil, err
	}

	return lo.Associate(pps, func(p *presenter.Product) (string, *presenter.Product) {
		return p.BCID, p
	}), nil
}

func fetchProductsFromServiceByBCIDs(service product.Service, pids []string) (map[string]*presenter.Product, error) {
	pps, err := service.FetchProducts(product.WithBCIDs(pids...), product.WithStockInformation())
	if err != nil {
		return nil, err
	}

	return lo.Associate(pps, func(p *presenter.Product) (string, *presenter.Product) {
		return p.BCID, p
	}), nil
}

func getBCIDsForProducts(ps []bigc.Product) []string {
	ids := lo.Map(ps, func(p bigc.Product, _ int) string {
		return strconv.Itoa(p.ID)
	})

	return lo.Uniq(lo.Filter(ids, func(item string, index int) bool {
		return item != ""
	}))
}

func getBCIDsForVariants(vs []bigc.Variant) []string {
	ids := lo.Map(vs, func(v bigc.Variant, _ int) string {
		return strconv.Itoa(v.ProductID) + "/" + strconv.Itoa(v.ID)
	})

	return lo.Uniq(lo.Filter(ids, func(item string, index int) bool {
		return item != ""
	}))
}

func filterForActiveProducts(ps []bigc.Product) []bigc.Product {
	return lo.Filter(blacklist.ApplyForProducts(ps), func(item bigc.Product, index int) bool {
		return item.Sku != "" && item.IsVisible && !strings.HasPrefix(item.Sku, "//") && !slices.Contains(item.Categories, 1230) && !strings.Contains(item.Sku, "-")
	})
}

func filterForActiveVariants(ps []bigc.Product) []bigc.Variant {
	var retv []bigc.Variant

	for _, p := range ps {
		for _, v := range p.Variants {
			if !v.PurchasingDisabled && !strings.HasPrefix(v.Sku, "//") {
				retv = append(retv, v)
			}
		}
	}

	return blacklist.ApplyForVariants(retv)
}
