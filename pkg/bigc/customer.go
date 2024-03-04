package bigc

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type CustomerGetResponse struct {
	Data []Customer `json:"data,omitempty"`
	Meta Meta       `json:"meta,omitempty"`
}

type Addresses struct {
	FirstName       string `json:"first_name,omitempty"`
	LastName        string `json:"last_name,omitempty"`
	Address1        string `json:"address1,omitempty"`
	Address2        string `json:"address2,omitempty"`
	City            string `json:"city,omitempty"`
	StateOrProvince string `json:"state_or_province,omitempty"`
	PostalCode      string `json:"postal_code,omitempty"`
	CountryCode     string `json:"country_code,omitempty"`
	Phone           string `json:"phone,omitempty"`
	AddressType     string `json:"address_type,omitempty"`
	CustomerID      int    `json:"customer_id,omitempty"`
	ID              int    `json:"id,omitempty"`
	Country         string `json:"country,omitempty"`
}
type StoreCreditAmounts struct {
	Amount float64 `json:"amount,omitempty"`
}
type Customer struct {
	Email                                   string               `json:"email,omitempty"`
	FirstName                               string               `json:"first_name,omitempty"`
	LastName                                string               `json:"last_name,omitempty"`
	Company                                 string               `json:"company,omitempty"`
	Phone                                   string               `json:"phone,omitempty"`
	Notes                                   string               `json:"notes,omitempty"`
	TaxExemptCategory                       string               `json:"tax_exempt_category,omitempty"`
	CustomerGroupID                         int                  `json:"customer_group_id,omitempty"`
	Addresses                               []Addresses          `json:"addresses,omitempty"`
	StoreCreditAmounts                      []StoreCreditAmounts `json:"store_credit_amounts,omitempty"`
	AcceptsProductReviewAbandonedCartEmails bool                 `json:"accepts_product_review_abandoned_cart_emails,omitempty"`
	ChannelIds                              []int                `json:"channel_ids,omitempty"`
	ShopperProfileID                        string               `json:"shopper_profile_id,omitempty"`
	SegmentIds                              []string             `json:"segment_ids,omitempty"`
}
type Links struct {
	Previous string `json:"previous,omitempty"`
	Current  string `json:"current,omitempty"`
	Next     string `json:"next,omitempty"`
}
type Pagination struct {
	Total       int   `json:"total,omitempty"`
	Count       int   `json:"count,omitempty"`
	PerPage     int   `json:"per_page,omitempty"`
	CurrentPage int   `json:"current_page,omitempty"`
	TotalPages  int   `json:"total_pages,omitempty"`
	Links       Links `json:"links,omitempty"`
}

func (c *Client) GetCustomers(params map[string]string) ([]Customer, error) {
	var customers []Customer
	url := GetUrl(c.BaseURL, "/customers", params)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return customers, err
	}

	data, err := c.doRequest(req)
	if err != nil {
		return customers, err
	}

	if data == nil {
		return customers, nil
	}

	var resp CustomerGetResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return customers, err
	}

	return resp.Data, nil
}

func (c *Client) UpdateCustomer(customer Customer, id int, opts ...CustomerUpdateOpt) (*Customer, error) {
	for _, o := range opts {
		o(&customer)
	}

	type CustomerUpdate struct {
		Customer
		Id int `json:"id,omitempty"`
	}

	data, err := json.Marshal([]CustomerUpdate{{customer, id}})
	if err != nil {
		return nil, err
	}
	url := GetUrl(c.BaseURL, "/customers", map[string]string{})
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	respData, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var resp CustomerGetResponse
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, nil
	}

	return &resp.Data[0], nil
}

type CustomerUpdateOpt func(*Customer)

func WithCustomerGroupID(id int) CustomerUpdateOpt {
	return func(c *Customer) {
		c.CustomerGroupID = id
	}
}
