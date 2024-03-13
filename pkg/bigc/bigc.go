package bigc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/plantyplantman/bcapi/pkg/env"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

type Client struct {
	BaseURL     string
	AccessToken string
	httpClient  *http.Client
	logger      *logrus.Logger
	MaxWorkers  int
}

func initLogger() *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	return log
}

func NewClient(baseURL, accessToken string, logger *logrus.Logger) *Client {
	if logger == nil {
		logger = initLogger()
	}

	return &Client{
		BaseURL:     baseURL,
		AccessToken: accessToken,
		httpClient:  &http.Client{},
		logger:      logger,
		MaxWorkers:  50,
	}
}

func MustGetClient() *Client {
	return NewClient("https://api.bigcommerce.com/stores/9wn8an8lno/v3", env.BIGCOMMERCE, nil)
}

func GetClient() (*Client, error) {
	return NewClient("https://api.bigcommerce.com/stores/9wn8an8lno/v3", env.BIGCOMMERCE, nil), nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Add("X-Auth-Token", c.AccessToken)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 204 {
		return nil, nil
	} else if resp.StatusCode != 200 {
		return nil, &RequestError{resp.StatusCode, errors.New("request error"), body}
	}
	return body, nil
}

func (c *Client) DeleteProduct(productID int) error {
	url := GetUrl(c.BaseURL, fmt.Sprintf("/catalog/products/%d", productID), map[string]string{})
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req)
	return err
}

func (c *Client) DeleteProducts(productIDs []int) chan error {
	totalTasks := len(productIDs)
	semaphore := make(chan struct{}, c.MaxWorkers)
	errCh := make(chan error, totalTasks)

	for i := 0; i < totalTasks; i++ {
		semaphore <- struct{}{}
		go func(i int) {
			defer func() { <-semaphore }()
			if err := c.DeleteProduct(productIDs[i]); err != nil {
				errCh <- err
			}
		}(i)
	}
	// Wait for all goroutines to finish
	for i := 0; i < c.MaxWorkers; i++ {
		semaphore <- struct{}{}
	}

	close(semaphore)
	close(errCh)

	return errCh
}

func (c *Client) GetVariants(productID int, params map[string]string) ([]Variant, error) {
	url := GetUrl(c.BaseURL, fmt.Sprintf("/catalog/products/%d/variants", productID), params)
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

	var resp VariantsGetResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

func (c *Client) DeleteVariant(productID int, variantID int) error {
	url := GetUrl(c.BaseURL, fmt.Sprintf("/catalog/products/%d/variants/%d", productID, variantID), map[string]string{})
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req)
	return err
}

func (c *Client) DeleteVariants(ids []lo.Tuple2[int, int]) chan error {
	totalTasks := len(ids)
	semaphore := make(chan struct{}, c.MaxWorkers)
	errCh := make(chan error, totalTasks)

	for i := 0; i < totalTasks; i++ {
		semaphore <- struct{}{}
		go func(i int) {
			defer func() { <-semaphore }()
			if err := c.DeleteVariant(ids[i].A, ids[i].B); err != nil {
				errCh <- err
			}
		}(i)
	}
	// Wait for all goroutines to finish
	for i := 0; i < c.MaxWorkers; i++ {
		semaphore <- struct{}{}
	}

	close(semaphore)
	close(errCh)

	return errCh
}

func (c *Client) GetVariantById(variantID int, productID int, params map[string]string) (*Variant, error) {
	url := GetUrl(c.BaseURL, fmt.Sprintf("/catalog/products/%d/variants/%d", productID, variantID), params)
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

	var resp VariantGetResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

func (c *Client) UpdateVariant(v *Variant, opts ...UpdateVariantOpt) (*Variant, error) {

	for _, f := range opts {
		f(v)
	}

	var (
		data []byte

		url string
		r   *http.Request

		err error
	)

	data, err = json.Marshal(v)
	if err != nil {
		return nil, err
	}

	url = GetUrl(c.BaseURL,
		fmt.Sprintf("/catalog/products/%d/variants/%d", v.ProductID, v.ID),
		map[string]string{})
	r, err = http.NewRequest("PUT", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	respData, err := c.doRequest(r)
	if err != nil {
		return nil, err
	}

	var resp UpdateVariantResp
	if err = json.Unmarshal(respData, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c *Client) CreateCustomField(productId int, cf *CustomField) (*CustomField, error) {
	data, err := json.Marshal(cf)
	if err != nil {
		return &CustomField{}, err
	}

	url := GetUrl(c.BaseURL, fmt.Sprintf("/catalog/products/%d/custom-fields", productId), map[string]string{})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return &CustomField{}, err
	}

	respData, err := c.doRequest(req)
	if err != nil {
		return &CustomField{}, err
	}

	var resp CustomFieldPostResponse
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return &CustomField{}, err
	}
	return &resp.Data, nil
}

func (c *Client) AddShippingField(productId int) error {
	shippingField := CustomField{
		Name:  "Shipping",
		Value: "Ships in 1-7 business days",
	}

	resp, err := c.CreateCustomField(productId, &shippingField)
	if err != nil {
		return err
	}

	if resp.Name != shippingField.Name || resp.Value != shippingField.Value {
		return errors.New("failed to add shipping field")
	}
	return nil
}

// == CATEGORIES ==
func (c Client) GetCategories(params map[string]string) ([]Category, error) {
	url := GetUrl(c.BaseURL, "/catalog/categories", params)
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

	var resp CategoriesGetResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (c Client) CreateCategory(categoryReq Category) (*Category, error) {
	data, err := json.Marshal(categoryReq)
	if err != nil {
		return nil, err
	}
	url := GetUrl(c.BaseURL, "/catalog/categories", map[string]string{})
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	respData, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var resp CategoriesPostResponse
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

func WithUpdateDesc(desc string) func(*Category) *Category {
	return func(cat *Category) *Category {
		cat.Description = desc
		return cat
	}
}

func WithUpdateMetaDesc(meta string) func(*Category) *Category {
	return func(cat *Category) *Category {
		cat.MetaDescription = meta
		return cat
	}
}

func WithUpdateMetaKeywords(keywords []string) func(*Category) *Category {
	return func(cat *Category) *Category {
		cat.MetaKeywords = keywords
		return cat
	}
}

func (c Client) UpdateCategory(cat *Category, opts ...func(*Category) *Category) (*Category, error) {
	for _, o := range opts {
		o(cat)
	}
	data, err := json.Marshal(cat)
	if err != nil {
		return nil, err
	}
	url := GetUrl(c.BaseURL, "/catalog/categories/"+fmt.Sprint(cat.ID), map[string]string{})
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	respData, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	var resp CategoriesPutResponse

	if err = json.Unmarshal(respData, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

func (c Client) ExportCategories(path string, categories []Category) error {
	if categories == nil {
		return errors.New("can't export empty category list")
	}
	if path == "" {
		return errors.New("no path specified")
	}

	headers, e := GetStructFields(categories[0])
	if e != nil {
		return e
	}

	var content [][]string
	var tmp []string
	for _, p := range categories {
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

func (c Client) GetCategoryFromID(id int) (Category, error) {
	cs, e := c.GetCategories(map[string]string{"id": fmt.Sprint(id)})
	if e != nil || len(cs) == 0 {
		return Category{}, e
	}
	return cs[0], nil
}

func (c Client) GetAllVisibleCategoriesWOutDescriptions() ([]Category, error) {
	var retv []Category

	cats, e := c.GetAllCategories(map[string]string{"is_visible": "true"})
	if e != nil {
		return nil, e
	}

	for _, cat := range cats {
		if cat.Description == "" {
			retv = append(retv, cat)
		}
	}
	return retv, nil
}

func (c Client) GetAllCategories(params map[string]string) ([]Category, error) {
	page := 1
	limit := 100
	var retv []Category

	for {
		pagination := map[string]string{
			"page":  fmt.Sprint(page),
			"limit": fmt.Sprint(limit),
		}
		maps.Copy(params, pagination)
		cats, e := c.GetCategories(params)
		if e != nil {
			return nil, e
		}
		if len(cats) == 0 {
			break
		}

		retv = append(retv, cats...)
		page++
	}
	return retv, nil
}
