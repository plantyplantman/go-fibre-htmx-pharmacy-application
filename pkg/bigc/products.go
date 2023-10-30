package bigc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

func (p *NewProduct) Fill(c *AI_Client) (*NewProduct, error) {
	if p.Name == "" || p.Sku == "" || p.Price == 0 {
		return nil, errors.New("missing required fields: Name, Sku, Price")
	}

	name, err := c.GenerateProductNameFromMinfosName(p.Name, openai.GPT4)
	if err != nil {
		log.Println(err)
	} else {
		p.Name = name
	}
	p.Mpn = p.Sku
	p.Upc = p.Sku
	p.Gtin = p.Sku

	if p.MetaKeywords == nil || p.SearchKeywords == "" {
		keywords, err := c.GenerateSearchKeywordsFromProductName(p.Name, openai.GPT4)
		if err != nil {
			return nil, err
		}

		if p.MetaKeywords == nil {
			p.MetaKeywords = strings.Split(keywords, ",")
		}
		if p.SearchKeywords == "" {
			p.SearchKeywords = keywords
		}
	}

	if p.MetaDescription == "" {
		meta, err := c.GenerateMetadataFromProductName(p.Name, openai.GPT4)
		if err != nil {
			return nil, err
		}
		p.MetaDescription = meta
	}

	if p.Categories == nil {
		p.Categories = []int{NEW}
	}
	if p.Weight == 0 {
		p.Weight = 100
	}
	p.RelatedProducts = []int{0}
	p.Availability = "available"
	p.Condition = "New"
	p.InventoryTracking = "product"
	p.Type = "physical"
	p.IsVisible = false

	return p, nil
}

type Products []Product

func (ps Products) Export(path string) error {
	headers, e := GetStructFields(ps[0])
	if e != nil {
		return e
	}

	var content [][]string
	var tmp []string
	for _, p := range ps {
		values, e := GetStructValues(p)
		if e != nil {
			return e
		}
		for _, v := range values {
			tmp = append(tmp, fmt.Sprint(v))
		}
		content = append(content, tmp)
		tmp = []string{}
	}

	return WriteToTsv(path, headers, content)
}

func defaultProductUpdateOpts(og *Product) ProductUpdateOpts {
	return ProductUpdateOpts{
		ID:              og.ID,
		Name:            og.Name,
		Type:            og.Type,
		Sku:             og.Sku,
		Description:     og.Description,
		Weight:          og.Weight,
		Width:           og.Width,
		Depth:           og.Depth,
		Height:          og.Height,
		Price:           og.Price,
		CostPrice:       og.CostPrice,
		RetailPrice:     0,
		SalePrice:       og.SalePrice,
		InventoryLevel:  og.InventoryLevel,
		PageTitle:       og.PageTitle,
		MetaKeywords:    og.MetaKeywords,
		MetaDescription: og.MetaDescription,
		IsVisible:       og.IsVisible,
		Categories:      og.Categories,
	}
}

type ProductUpdateOptFn func(opt *ProductUpdateOpts) map[string]any

func WithUpdateProductSku(sku string) ProductUpdateOptFn {
	return func(opt *ProductUpdateOpts) map[string]any {
		opt.Sku = sku
		return map[string]any{"sku": sku}
	}
}

func WithUpdateProductPrice(price float64) ProductUpdateOptFn {
	return func(opt *ProductUpdateOpts) map[string]any {
		opt.Price = price
		return map[string]any{"price": price}
	}
}

func WithUpdateProductInventoryLevel(soh int) ProductUpdateOptFn {
	return func(opt *ProductUpdateOpts) map[string]any {
		opt.InventoryLevel = soh
		return map[string]any{"soh": soh}
	}
}

func WithUpdateProductIsVisible(isVisible bool) ProductUpdateOptFn {
	return func(opt *ProductUpdateOpts) map[string]any {
		opt.IsVisible = isVisible
		return map[string]any{"isVisible": isVisible}
	}
}

func WithUpdateProductSalePrice(salePrice float64) ProductUpdateOptFn {
	return func(opt *ProductUpdateOpts) map[string]any {
		opt.SalePrice = salePrice
		return map[string]any{"salePrice": salePrice}
	}
}

func WithUpdateProductCategoriesWithoutSaleIDs(ids []int) ProductUpdateOptFn {
	newIds := RemoveSaleCategories(ids)
	return func(opt *ProductUpdateOpts) map[string]any {
		opt.Categories = newIds
		return map[string]any{"categories": newIds}
	}
}
func WithUpdateProductCategories(catIDs []int) ProductUpdateOptFn {
	return func(opt *ProductUpdateOpts) map[string]any {
		opt.Categories = catIDs
		return map[string]any{"categories": catIDs}
	}
}
func WithUpdateProductCategoryIsRetired(b bool) ProductUpdateOptFn {
	return func(opt *ProductUpdateOpts) map[string]any {
		if b {
			return map[string]any{"categories": append(opt.Categories, RETIRED_PRODUCTS)}
		} else {
			return map[string]any{"categories": lo.Filter(opt.Categories, func(i int, _ int) bool {
				return i != RETIRED_PRODUCTS
			})}
		}
	}
}

func WithUpdateProductMetaKeywords(kw []string) ProductUpdateOptFn {
	return func(opt *ProductUpdateOpts) map[string]any {
		opt.MetaKeywords = kw
		return map[string]any{"metaKeywords": kw}
	}
}

func WithUpdateProductMetaDesc(desc string) ProductUpdateOptFn {
	return func(opt *ProductUpdateOpts) map[string]any {
		opt.MetaDescription = desc
		return map[string]any{"metaDescription": desc}
	}
}

func WithUpdateProductPageTitle(title string) ProductUpdateOptFn {
	return func(opt *ProductUpdateOpts) map[string]any {
		opt.PageTitle = title
		return map[string]any{"pageTitle": title}
	}
}

func WithUpdateProductCostPrice(costPrice float64) ProductUpdateOptFn {
	return func(opt *ProductUpdateOpts) map[string]any {
		opt.CostPrice = costPrice
		return map[string]any{"costPrice": costPrice}
	}
}

func NewUpdateProductReq(og *Product, optFuncs ...ProductUpdateOptFn) *ProductUpdateReq {
	o := defaultProductUpdateOpts(og)
	for _, fn := range optFuncs {
		fn(&o)
	}
	return &ProductUpdateReq{
		Id:  og.ID,
		Req: o,
	}
}

// == PRODUCTS ==

func (c *BigCommerceClient) GetProductById(id int) (*Product, error) {
	p, err := c.GetProducts(map[string]string{"id": fmt.Sprint(id), "include": "images,variants"})
	if err != nil {
		return nil, err
	}
	if len(p) < 1 {
		return nil, &ProductNotFoundError{}
	}
	return &p[0], nil
}
func (c *BigCommerceClient) GetProductFromSku(sku string) (Product, error) {
	p, e := c.GetProducts(map[string]string{"sku": sku, "include": "images,variants"})
	if e != nil {
		return Product{}, e
	}

	if len(p) < 1 {
		if len(sku) < 12 {
			return c.GetProductFromSku(fmt.Sprintf("%d%v", 0, sku))
		}
		if strings.HasPrefix(sku, "//") {
			return Product{}, &ProductNotFoundError{Sku: sku}
		}
		return c.GetProductFromSku(fmt.Sprintf("/%v", sku))
	}
	return p[0], nil
}

func (c *BigCommerceClient) GetProducts(params map[string]string) ([]Product, error) {
	url := GetUrl(c.BaseURL, "/catalog/products", params)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	data, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	var resp ProductsGetResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

func (c *BigCommerceClient) GetAllProducts(params map[string]string) ([]Product, error) {
	var retv []Product
	page := 1
	limit := 100
	for {
		pagination := map[string]string{
			"page":    fmt.Sprint(page),
			"limit":   fmt.Sprint(limit),
			"include": "variants",
		}
		maps.Copy(params, pagination)
		products, err := c.GetProducts(params)
		if err != nil {
			return nil, err
		}

		if len(products) == 0 {
			break
		}

		retv = append(retv, products...)
		page++
	}

	return retv, nil
}

func (c *BigCommerceClient) CreateProduct(productReq NewProduct) (*Product, error) {
	data, err := json.Marshal(productReq)
	if err != nil {
		return nil, err
	}
	url := GetUrl(c.BaseURL, "/catalog/products", map[string]string{})
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	respData, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var resp ProductsPostResponse
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}
	c.logger.WithFields(logrus.Fields{
		"created": "product",
		"product": resp.Data,
	}).Info("Created product")

	return &resp.Data, nil
}

func (c *BigCommerceClient) GetProductIDFromSku(sku string) (int, error) {
	p, e := c.GetProducts(map[string]string{"sku": sku, "include_fields": "sku"})
	if e != nil {
		return 0, e
	}
	if len(p) < 1 {
		return 0, errors.New(fmt.Sprintf("Product with Sku: %v not found", sku))
	}
	return p[0].ID, nil
}

func (c *BigCommerceClient) UpdateProduct(og *Product, optFuncs ...ProductUpdateOptFn) (Product, error) {
	o := defaultProductUpdateOpts(og)
	updateFields := map[string]any{}

	for _, fn := range optFuncs {
		tmp := fn(&o)
		for k, v := range tmp {
			updateFields[k] = v
		}
	}

	data, err := json.Marshal(o)
	if err != nil {
		return Product{}, err
	}

	url := GetUrl(c.BaseURL,
		fmt.Sprintf("/catalog/products/%d", og.ID),
		map[string]string{})
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(data))
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"update":   "product",
			"id":       og.ID,
			"fields":   updateFields,
			"previous": og,
		}).Error("Failed to update product " + err.Error())
		return Product{}, err
	}

	respData, err := c.doRequest(req)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"update":   "product",
			"id":       og.ID,
			"fields":   updateFields,
			"previous": og,
		}).Error("Failed to update product " + err.Error())
		return Product{}, err
	}

	var resp ProductsPutResponse
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"update":   "product",
			"id":       og.ID,
			"fields":   updateFields,
			"previous": og,
		}).Error("Failed to unmarshal ProductsPutResponse " + err.Error())
		return Product{}, err
	}

	c.logger.WithFields(logrus.Fields{
		"update":   "product",
		"id":       og.ID,
		"fields":   updateFields,
		"previous": og,
	}).Info("Updated product")

	return resp.Data, nil
}
