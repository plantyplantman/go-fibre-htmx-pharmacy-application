package bigc

import (
	"time"
)

type ProductsGetResponse struct {
	Data []Product `json:"data"`
	Meta Meta      `json:"meta"`
}

type ProductsPostResponse struct {
	Data Product `json:"data"`
	Meta Meta    `json:"meta"`
}

type ProductsPutResponse struct {
	Data Product `json:"data"`
	Meta Meta    `json:"meta"`
}

type CategoriesGetResponse struct {
	Data []Category `json:"data"`
	Meta Meta       `json:"meta"`
}

type CategoriesPostResponse struct {
	Data Category `json:"data"`
	Meta Meta     `json:"meta"`
}

type CategoriesPutResponse struct {
	Data Category `json:"data"`
	Meta Meta     `json:"meta"`
}

type VariantsGetResponse struct {
	Data []Variant `json:"data"`
	Meta Meta      `json:"meta"`
}

type VariantGetResponse struct {
	Data Variant `json:"data"`
	Meta Meta    `json:"meta"`
}

type UpdateVariantResp struct {
	Data Variant `json:"data"`
	Meta Meta    `json:"meta"`
}

type CustomFieldPostResponse struct {
	Data CustomField `json:"data"`
	Meta Meta        `json:"meta"`
}

type ProductUpdateReq struct {
	Id  int
	Req ProductUpdateOpts
}

type Meta struct {
	Pagination struct {
		Total       int `json:"total"`
		Count       int `json:"count"`
		PerPage     int `json:"per_page"`
		CurrentPage int `json:"current_page"`
		TotalPages  int `json:"total_pages"`
		Links       struct {
			Current string `json:"current"`
		} `json:"links"`
		TooMany bool `json:"too_many"`
	} `json:"pagination"`
}

type CustomURL struct {
	URL          string `json:"url"`
	IsCustomized bool   `json:"is_customized"`
}

type Category struct {
	ID                 int       `json:"id"`
	ParentID           int       `json:"parent_id"`
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	Views              int       `json:"views"`
	SortOrder          int       `json:"sort_order"`
	PageTitle          string    `json:"page_title"`
	MetaKeywords       []string  `json:"meta_keywords"`
	MetaDescription    string    `json:"meta_description"`
	LayoutFile         string    `json:"layout_file"`
	ImageURL           string    `json:"image_url"`
	IsVisible          bool      `json:"is_visible"`
	SearchKeywords     string    `json:"search_keywords"`
	DefaultProductSort string    `json:"default_product_sort"`
	CustomURL          CustomURL `json:"custom_url"`
}

type Variant struct {
	ID                        int            `json:"id"`
	ProductID                 int            `json:"product_id"`
	Sku                       string         `json:"sku"`
	SkuID                     int            `json:"sku_id"`
	Price                     float64        `json:"price"`
	CalculatedPrice           float64        `json:"calculated_price"`
	SalePrice                 float64        `json:"sale_price"`
	RetailPrice               float64        `json:"retail_price"`
	Weight                    float64        `json:"weight"`
	CalculatedWeight          float64        `json:"calculated_weight"`
	Width                     float64        `json:"width"`
	Height                    float64        `json:"height"`
	Depth                     float64        `json:"depth"`
	IsFreeShipping            bool           `json:"is_free_shipping"`
	FixedCostShippingPrice    float64        `json:"fixed_cost_shipping_price"`
	PurchasingDisabled        bool           `json:"purchasing_disabled"`
	PurchasingDisabledMessage string         `json:"purchasing_disabled_message"`
	ImageURL                  string         `json:"image_url"`
	CostPrice                 float64        `json:"cost_price"`
	Upc                       string         `json:"upc"`
	Mpn                       string         `json:"mpn"`
	Gtin                      string         `json:"gtin"`
	InventoryLevel            int            `json:"inventory_level"`
	InventoryWarningLevel     int            `json:"inventory_warning_level"`
	BinPickingNumber          string         `json:"bin_picking_number"`
	OptionValues              []OptionValues `json:"option_values"`
}

type OptionValues struct {
	ID                int    `json:"id"`
	Label             string `json:"label"`
	OptionID          int    `json:"option_id"`
	OptionDisplayName string `json:"option_display_name"`
}

type UpdateVariantOpt func(*Variant)

func WithUpdateVariantPurchasingDisabled(b bool) UpdateVariantOpt {
	return func(v *Variant) {
		v.PurchasingDisabled = b
	}
}
func WithUpdateVariantPrice(price float64) UpdateVariantOpt {
	return func(v *Variant) {
		v.Price = price
	}
}
func WithUpdateVariantSalePrice(price float64) UpdateVariantOpt {
	return func(v *Variant) {
		v.SalePrice = price
	}
}
func WithUpdateVariantRetailPrice(price float64) UpdateVariantOpt {
	return func(v *Variant) {
		v.RetailPrice = price
	}
}
func WithUpdateVariantCostPrice(price float64) UpdateVariantOpt {
	return func(v *Variant) {
		v.CostPrice = price
	}
}
func WithUpdateVariantInventoryLevel(soh int) UpdateVariantOpt {
	return func(v *Variant) {
		v.InventoryLevel = soh
	}
}
func WithUpdateVariantOptionDisplayName(s string) UpdateVariantOpt {
	return func(v *Variant) {
		v.OptionValues[0].OptionDisplayName = s
	}
}
func WithUpdateVariantSku(sku string) UpdateVariantOpt {
	return func(v *Variant) {
		v.Sku = sku
	}
}

type ProductUpdateOpts struct {
	ID              int      `json:"id"`
	Name            string   `json:"name"`
	Type            string   `json:"type"`
	Sku             string   `json:"sku"`
	Description     string   `json:"description"`
	Weight          float64  `json:"weight"`
	Width           float64  `json:"width"`
	Depth           float64  `json:"depth"`
	Height          float64  `json:"height"`
	Price           float64  `json:"price"`
	CostPrice       float64  `json:"cost_price"`
	RetailPrice     float64  `json:"retail_price"`
	SalePrice       float64  `json:"sale_price"`
	InventoryLevel  int      `json:"inventory_level"`
	PageTitle       string   `json:"page_title"`
	MetaKeywords    []string `json:"meta_keywords"`
	MetaDescription string   `json:"meta_description"`
	IsVisible       bool     `json:"is_visible"`
	Categories      []int    `json:"categories"`
}

type Product struct {
	ID                          int                `json:"id"`
	Name                        string             `json:"name"`
	Type                        string             `json:"type"`
	Sku                         string             `json:"sku"`
	Description                 string             `json:"description"`
	Weight                      float64            `json:"weight"`
	Width                       float64            `json:"width"`
	Depth                       float64            `json:"depth"`
	Height                      float64            `json:"height"`
	Price                       float64            `json:"price"`
	CostPrice                   float64            `json:"cost_price"`
	RetailPrice                 float64            `json:"retail_price"`
	SalePrice                   float64            `json:"sale_price"`
	MapPrice                    float64            `json:"map_price"`
	TaxClassID                  int                `json:"tax_class_id"`
	ProductTaxCode              string             `json:"product_tax_code"`
	CalculatedPrice             float64            `json:"calculated_price"`
	Categories                  []int              `json:"categories"`
	BrandID                     int                `json:"brand_id"`
	OptionSetID                 any                `json:"option_set_id"`
	OptionSetDisplay            string             `json:"option_set_display"`
	InventoryLevel              int                `json:"inventory_level"`
	InventoryWarningLevel       int                `json:"inventory_warning_level"`
	InventoryTracking           string             `json:"inventory_tracking"`
	ReviewsRatingSum            int                `json:"reviews_rating_sum"`
	ReviewsCount                int                `json:"reviews_count"`
	TotalSold                   int                `json:"total_sold"`
	FixedCostShippingPrice      float64            `json:"fixed_cost_shipping_price"`
	IsFreeShipping              bool               `json:"is_free_shipping"`
	IsVisible                   bool               `json:"is_visible"`
	IsFeatured                  bool               `json:"is_featured"`
	RelatedProducts             []int              `json:"related_products"`
	Warranty                    string             `json:"warranty"`
	BinPickingNumber            string             `json:"bin_picking_number"`
	LayoutFile                  string             `json:"layout_file"`
	Upc                         string             `json:"upc"`
	Mpn                         string             `json:"mpn"`
	Gtin                        string             `json:"gtin"`
	SearchKeywords              string             `json:"search_keywords"`
	Availability                string             `json:"availability"`
	AvailabilityDescription     string             `json:"availability_description"`
	GiftWrappingOptionsType     string             `json:"gift_wrapping_options_type"`
	GiftWrappingOptionsList     []any              `json:"gift_wrapping_options_list"`
	SortOrder                   int                `json:"sort_order"`
	Condition                   string             `json:"condition"`
	IsConditionShown            bool               `json:"is_condition_shown"`
	OrderQuantityMinimum        int                `json:"order_quantity_minimum"`
	OrderQuantityMaximum        int                `json:"order_quantity_maximum"`
	PageTitle                   string             `json:"page_title"`
	MetaKeywords                []string           `json:"meta_keywords"`
	MetaDescription             string             `json:"meta_description"`
	DateCreated                 time.Time          `json:"date_created"`
	DateModified                time.Time          `json:"date_modified"`
	ViewCount                   int                `json:"view_count"`
	PreorderReleaseDate         any                `json:"preorder_release_date"`
	PreorderMessage             string             `json:"preorder_message"`
	IsPreorderOnly              bool               `json:"is_preorder_only"`
	IsPriceHidden               bool               `json:"is_price_hidden"`
	PriceHiddenLabel            string             `json:"price_hidden_label"`
	CustomURL                   CustomURL          `json:"custom_url"`
	BaseVariantID               int                `json:"base_variant_id"`
	OpenGraphType               string             `json:"open_graph_type"`
	OpenGraphTitle              string             `json:"open_graph_title"`
	OpenGraphDescription        string             `json:"open_graph_description"`
	OpenGraphUseMetaDescription bool               `json:"open_graph_use_meta_description"`
	OpenGraphUseProductName     bool               `json:"open_graph_use_product_name"`
	OpenGraphUseImage           bool               `json:"open_graph_use_image"`
	Variants                    []Variant          `json:"variants"`
	BulkPricingRules            []BulkPricingRules `json:"bulk_pricing_rules"`
	Images                      []Image            `json:"images"`
	Videos                      []Video            `json:"videos"`
	CustomFields                []CustomField      `json:"custom_fields"`
}

type CustomField struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type BulkPricingRules struct {
	QuantityMin int    `json:"quantity_min"`
	QuantityMax int    `json:"quantity_max"`
	Type        string `json:"type"`
	Amount      int    `json:"amount"`
}

type Image struct {
	ImageFile    string    `json:"image_file"`
	IsThumbnail  bool      `json:"is_thumbnail"`
	SortOrder    int64     `json:"sort_order"`
	Description  string    `json:"description"`
	ImageURL     string    `json:"image_url"`
	ID           int       `json:"id"`
	ProductID    int       `json:"product_id"`
	DateModified time.Time `json:"date_modified"`
}

type Video struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	Type        string `json:"type"`
	VideoID     string `json:"video_id"`
	ID          int    `json:"id"`
	ProductID   int    `json:"product_id"`
	Length      string `json:"length"`
}

type Config struct {
	DefaultValue                string   `json:"default_value"`
	CheckedByDefault            bool     `json:"checked_by_default"`
	CheckboxLabel               string   `json:"checkbox_label"`
	DateLimited                 bool     `json:"date_limited"`
	DateLimitMode               string   `json:"date_limit_mode"`
	DateEarliestValue           string   `json:"date_earliest_value"`
	DateLatestValue             string   `json:"date_latest_value"`
	FileTypesMode               string   `json:"file_types_mode"`
	FileTypesSupported          []string `json:"file_types_supported"`
	FileTypesOther              []string `json:"file_types_other"`
	FileMaxSize                 int      `json:"file_max_size"`
	TextCharactersLimited       bool     `json:"text_characters_limited"`
	TextMinLength               int      `json:"text_min_length"`
	TextMaxLength               int      `json:"text_max_length"`
	TextLinesLimited            bool     `json:"text_lines_limited"`
	TextMaxLines                int      `json:"text_max_lines"`
	NumberLimited               bool     `json:"number_limited"`
	NumberLimitMode             string   `json:"number_limit_mode"`
	NumberLowestValue           int      `json:"number_lowest_value"`
	NumberHighestValue          int      `json:"number_highest_value"`
	NumberIntegersOnly          bool     `json:"number_integers_only"`
	ProductListAdjustsInventory bool     `json:"product_list_adjusts_inventory"`
	ProductListAdjustsPricing   bool     `json:"product_list_adjusts_pricing"`
	ProductListShippingCalc     string   `json:"product_list_shipping_calc"`
}

type NewProduct struct {
	Name                   string   `json:"name"`
	Type                   string   `json:"type"`
	Sku                    string   `json:"sku"`
	Description            string   `json:"description"`
	Weight                 float64  `json:"weight"`
	Width                  float64  `json:"width"`
	Depth                  float64  `json:"depth"`
	Height                 float64  `json:"height"`
	Price                  float64  `json:"price"`
	CostPrice              float64  `json:"cost_price"`
	RetailPrice            float64  `json:"retail_price"`
	SalePrice              float64  `json:"sale_price"`
	Categories             []int    `json:"categories"`
	BrandID                int      `json:"brand_id"`
	InventoryLevel         int      `json:"inventory_level"`
	InventoryWarningLevel  int      `json:"inventory_warning_level"`
	InventoryTracking      string   `json:"inventory_tracking"`
	FixedCostShippingPrice float64  `json:"fixed_cost_shipping_price"`
	IsFreeShipping         bool     `json:"is_free_shipping"`
	IsVisible              bool     `json:"is_visible"`
	IsFeatured             bool     `json:"is_featured"`
	RelatedProducts        []int    `json:"related_products"`
	BinPickingNumber       string   `json:"bin_picking_number"`
	Upc                    string   `json:"upc"`
	Mpn                    string   `json:"mpn"`
	Gtin                   string   `json:"gtin"`
	SearchKeywords         string   `json:"search_keywords"`
	Condition              string   `json:"condition"`
	IsConditionShown       bool     `json:"is_condition_shown"`
	PageTitle              string   `json:"page_title"`
	MetaKeywords           []string `json:"meta_keywords"`
	MetaDescription        string   `json:"meta_description"`
	Availability           string   `json:"availability"`
}
