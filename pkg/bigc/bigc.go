package bigc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/plantyplantman/bcapi/pkg/env"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

type BigCommerceClient struct {
	BaseURL     string
	AccessToken string
	httpClient  *http.Client
	logger      *logrus.Logger
}

func initLogger() *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	return log
}

func NewClient(baseURL, accessToken string, logger *logrus.Logger) *BigCommerceClient {
	if logger == nil {
		logger = initLogger()
	}

	return &BigCommerceClient{
		BaseURL:     baseURL,
		AccessToken: accessToken,
		httpClient:  &http.Client{},
		logger:      logger,
	}
}

func MustGetClient() *BigCommerceClient {
	return NewClient("https://api.bigcommerce.com/stores/9wn8an8lno/v3", env.BIGCOMMERCE, nil)
}

func GetClient() (*BigCommerceClient, error) {
	return NewClient("https://api.bigcommerce.com/stores/9wn8an8lno/v3", env.BIGCOMMERCE, nil), nil
}

func (c *BigCommerceClient) doRequest(req *http.Request) ([]byte, error) {
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

func (c *BigCommerceClient) GetVariants(productID int, params map[string]string) ([]Variant, error) {
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

func (c *BigCommerceClient) GetVariantById(variantID int, productID int, params map[string]string) (*Variant, error) {
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

func (c *BigCommerceClient) UpdateVariant(v *Variant, opts ...UpdateVariantOpt) (*Variant, error) {

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

func (c *BigCommerceClient) CreateCustomField(productId int, cf *CustomField) (*CustomField, error) {
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

func (c *BigCommerceClient) AddShippingField(productId int) error {
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

func (c BigCommerceClient) DeleteProducts(product_ids []string) error {
	return &NotImplementedError{}
}

// == CATEGORIES ==
func (c BigCommerceClient) GetCategories(params map[string]string) ([]Category, error) {
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

func (c BigCommerceClient) CreateCategory(categoryReq Category) (*Category, error) {
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

func (c BigCommerceClient) UpdateCategory(catID int, categoryReq Category) (*Category, error) {
	data, err := json.Marshal(categoryReq)
	if err != nil {
		return nil, err
	}
	url := GetUrl(c.BaseURL, "/catalog/categories/"+fmt.Sprint(catID), map[string]string{})
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

func (c BigCommerceClient) ExportCategories(path string, categories []Category) error {
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

func (c BigCommerceClient) GetCategoryFromID(id int) (Category, error) {
	cs, e := c.GetCategories(map[string]string{"id": fmt.Sprint(id)})
	if e != nil || len(cs) == 0 {
		return Category{}, e
	}
	return cs[0], nil
}

func (c BigCommerceClient) GetAllVisibleCategoriesWOutDescriptions() ([]Category, error) {
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

func (c BigCommerceClient) GetAllCategories(params map[string]string) ([]Category, error) {
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
