package bigc

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/exp/maps"
)

type Order struct {
	ID                     int    `json:"id"`
	CustomerID             int    `json:"customer_id"`
	DateCreated            string `json:"date_created"`
	DateModified           string `json:"date_modified"`
	DateShipped            string `json:"date_shipped"`
	StatusID               int    `json:"status_id"`
	Status                 string `json:"status"`
	SubtotalExTax          string `json:"subtotal_ex_tax"`
	SubtotalIncTax         string `json:"subtotal_inc_tax"`
	SubtotalTax            string `json:"subtotal_tax"`
	BaseShippingCost       string `json:"base_shipping_cost"`
	ShippingCostExTax      string `json:"shipping_cost_ex_tax"`
	ShippingCostIncTax     string `json:"shipping_cost_inc_tax"`
	ShippingCostTax        string `json:"shipping_cost_tax"`
	ShippingCostTaxClassID int    `json:"shipping_cost_tax_class_id"`
	BaseHandlingCost       string `json:"base_handling_cost"`
	HandlingCostExTax      string `json:"handling_cost_ex_tax"`
	HandlingCostIncTax     string `json:"handling_cost_inc_tax"`
	HandlingCostTax        string `json:"handling_cost_tax"`
	HandlingCostTaxClassID int    `json:"handling_cost_tax_class_id"`
	BaseWrappingCost       string `json:"base_wrapping_cost"`
	WrappingCostExTax      string `json:"wrapping_cost_ex_tax"`
	WrappingCostIncTax     string `json:"wrapping_cost_inc_tax"`
	WrappingCostTax        string `json:"wrapping_cost_tax"`
	WrappingCostTaxClassID int    `json:"wrapping_cost_tax_class_id"`
	TotalExTax             string `json:"total_ex_tax"`
	TotalIncTax            string `json:"total_inc_tax"`
	TotalTax               string `json:"total_tax"`
	ItemsTotal             int    `json:"items_total"`
	ItemsShipped           int    `json:"items_shipped"`
	PaymentMethod          string `json:"payment_method"`
	PaymentProviderID      string `json:"payment_provider_id"`
	PaymentStatus          string `json:"payment_status"`
	RefundedAmount         string `json:"refunded_amount"`
	OrderIsDigital         bool   `json:"order_is_digital"`
	StoreCreditAmount      string `json:"store_credit_amount"`
	GiftCertificateAmount  string `json:"gift_certificate_amount"`
	IPAddress              string `json:"ip_address"`
	IPAddressV6            string `json:"ip_address_v6"`
	GeoipCountry           string `json:"geoip_country"`
	GeoipCountryIso2       string `json:"geoip_country_iso2"`
	CurrencyID             int    `json:"currency_id"`
	CurrencyCode           string `json:"currency_code"`
	CurrencyExchangeRate   string `json:"currency_exchange_rate"`
	DefaultCurrencyID      int    `json:"default_currency_id"`
	DefaultCurrencyCode    string `json:"default_currency_code"`
	StaffNotes             string `json:"staff_notes"`
	CustomerMessage        string `json:"customer_message"`
	DiscountAmount         string `json:"discount_amount"`
	CouponDiscount         string `json:"coupon_discount"`
	ShippingAddressCount   int    `json:"shipping_address_count"`
	IsDeleted              bool   `json:"is_deleted"`
	EbayOrderID            string `json:"ebay_order_id"`
	CartID                 string `json:"cart_id"`
	BillingAddress         struct {
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		Company     string `json:"company"`
		Street1     string `json:"street_1"`
		Street2     string `json:"street_2"`
		City        string `json:"city"`
		State       string `json:"state"`
		Zip         string `json:"zip"`
		Country     string `json:"country"`
		CountryIso2 string `json:"country_iso2"`
		Phone       string `json:"phone"`
		Email       string `json:"email"`
		FormFields  []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"form_fields"`
	} `json:"billing_address"`
	IsEmailOptIn   bool   `json:"is_email_opt_in"`
	CreditCardType any    `json:"credit_card_type"`
	OrderSource    string `json:"order_source"`
	ChannelID      int    `json:"channel_id"`
	ExternalSource any    `json:"external_source"`
	Products       struct {
		URL      string `json:"url"`
		Resource string `json:"resource"`
	} `json:"products"`
	ShippingAddresses struct {
		URL      string `json:"url"`
		Resource string `json:"resource"`
	} `json:"shipping_addresses"`
	Coupons struct {
		URL      string `json:"url"`
		Resource string `json:"resource"`
	} `json:"coupons"`
	ExternalID                              any    `json:"external_id"`
	ExternalMerchantID                      any    `json:"external_merchant_id"`
	TaxProviderID                           string `json:"tax_provider_id"`
	StoreDefaultCurrencyCode                string `json:"store_default_currency_code"`
	StoreDefaultToTransactionalExchangeRate string `json:"store_default_to_transactional_exchange_rate"`
	CustomStatus                            string `json:"custom_status"`
	CustomerLocale                          string `json:"customer_locale"`
	ExternalOrderID                         string `json:"external_order_id"`
}

type Orders []struct{ Order }

type OrderProductsGetResp []struct {
	ID                   int    `json:"id"`
	OrderID              int    `json:"order_id"`
	ProductID            int    `json:"product_id"`
	OrderAddressID       int    `json:"order_address_id"`
	Name                 string `json:"name"`
	NameCustomer         string `json:"name_customer,omitempty"`
	NameMerchant         string `json:"name_merchant,omitempty"`
	Sku                  string `json:"sku"`
	Upc                  string `json:"upc,omitempty"`
	Type                 string `json:"type"`
	BasePrice            string `json:"base_price"`
	PriceExTax           string `json:"price_ex_tax"`
	PriceIncTax          string `json:"price_inc_tax"`
	PriceTax             string `json:"price_tax"`
	BaseTotal            string `json:"base_total"`
	TotalExTax           string `json:"total_ex_tax"`
	TotalIncTax          string `json:"total_inc_tax"`
	TotalTax             string `json:"total_tax"`
	Weight               string `json:"weight"`
	Quantity             int    `json:"quantity"`
	BaseCostPrice        string `json:"base_cost_price"`
	CostPriceIncTax      string `json:"cost_price_inc_tax"`
	CostPriceExTax       string `json:"cost_price_ex_tax"`
	CostPriceTax         string `json:"cost_price_tax"`
	IsRefunded           bool   `json:"is_refunded"`
	QuantityRefunded     int    `json:"quantity_refunded"`
	RefundAmount         string `json:"refund_amount"`
	ReturnID             int    `json:"return_id"`
	WrappingName         string `json:"wrapping_name"`
	BaseWrappingCost     string `json:"base_wrapping_cost"`
	WrappingCostExTax    string `json:"wrapping_cost_ex_tax"`
	WrappingCostIncTax   string `json:"wrapping_cost_inc_tax"`
	WrappingCostTax      string `json:"wrapping_cost_tax"`
	WrappingMessage      string `json:"wrapping_message"`
	QuantityShipped      int    `json:"quantity_shipped"`
	FixedShippingCost    string `json:"fixed_shipping_cost"`
	EbayItemID           string `json:"ebay_item_id"`
	EbayTransactionID    string `json:"ebay_transaction_id"`
	OptionSetID          int    `json:"option_set_id"`
	ParentOrderProductID any    `json:"parent_order_product_id"`
	IsBundledProduct     bool   `json:"is_bundled_product"`
	BinPickingNumber     string `json:"bin_picking_number"`
	ExternalID           any    `json:"external_id"`
	FulfillmentSource    string `json:"fulfillment_source"`
	Brand                string `json:"brand"`
	AppliedDiscounts     []struct {
		ID     string `json:"id"`
		Amount string `json:"amount"`
		Name   string `json:"name"`
		Code   any    `json:"code"`
		Target string `json:"target"`
	} `json:"applied_discounts"`
	ProductOptions []struct {
		ID                   int    `json:"id"`
		OptionID             int    `json:"option_id"`
		OrderProductID       int    `json:"order_product_id"`
		ProductOptionID      int    `json:"product_option_id"`
		DisplayName          string `json:"display_name"`
		DisplayNameCustomer  string `json:"display_name_customer,omitempty"`
		DisplayNameMerchant  string `json:"display_name_merchant,omitempty"`
		DisplayValue         string `json:"display_value"`
		DisplayValueCustomer string `json:"display_value_customer,omitempty"`
		DisplayValueMerchant string `json:"display_value_merchant,omitempty"`
		Value                string `json:"value"`
		Type                 string `json:"type"`
		Name                 string `json:"name"`
		DisplayStyle         string `json:"display_style"`
	} `json:"product_options"`
	ConfigurableFields    []any  `json:"configurable_fields"`
	GiftCertificateID     any    `json:"gift_certificate_id"`
	DiscountedTotalIncTax string `json:"discounted_total_inc_tax"`
	EventName             any    `json:"event_name,omitempty"`
	EventDate             any    `json:"event_date,omitempty"`
}

func (c *Client) GetOrder(orderId int) (*Order, error) {
	url := fmt.Sprintf("https://api.bigcommerce.com/stores/9wn8an8lno/v2/orders/%d", orderId)
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
	var resp Order
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetOrders(params map[string]string) (Orders, error) {
	url := GetUrl("https://api.bigcommerce.com/stores/9wn8an8lno/v2", "/orders", params)
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
	var resp Orders
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetAllOrders(params map[string]string) (Orders, error) {
	var retv Orders
	page := 1
	limit := 100
	for {

		pagination := map[string]string{
			"page":  fmt.Sprint(page),
			"limit": fmt.Sprint(limit),
		}
		maps.Copy(params, pagination)
		orders, err := c.GetOrders(params)
		if err != nil {
			return nil, err
		}

		if len(orders) == 0 {
			break
		}

		retv = append(retv, orders...)
		page++
	}

	return retv, nil
}

func (c *Client) GetOrderProducts(orderId int) (*OrderProductsGetResp, error) {
	url := fmt.Sprintf("https://api.bigcommerce.com/stores/9wn8an8lno/v2/orders/%d/products", orderId)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	data, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	var resp OrderProductsGetResp
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}
