package entities

import (
	"time"

	"gorm.io/gorm"
)

type Sale struct {
	gorm.Model
	CustomerCode  string     `json:"CustomerCode"`
	Date          csharpDate `json:"Date"`
	Discount      float64    `json:"Discount"`
	Gst           float64    `json:"Gst"`
	LastUpdated   string     `json:"LastUpdated"`
	MinfosSiteID  int        `json:"MinfosSiteId"`
	PromoDiscount float64    `json:"PromoDiscount"`
	SaleID        int        `json:"SaleId"`
	SaleLines     []SaleLine `json:"SaleLines"`
	SaleTypeID    int        `json:"SaleTypeId"`
	Total         float64    `json:"Total"`
}

type SaleLine struct {
	gorm.Model
	Cogs            float64 `json:"Cogs"`
	Discount        float64 `json:"Discount"`
	Gst             float64 `json:"Gst"`
	LineID          int     `json:"LineId"`
	MinfosSiteID    int     `json:"MinfosSiteId"`
	Mnpn            string  `json:"Mnpn"`
	Name            string  `json:"Name"`
	ProductID       string  `json:"ProductId"`
	PromoDiscount   float64 `json:"PromoDiscount"`
	Quantity        float64 `json:"Quantity"`
	SaleProductType int     `json:"SaleProductType"`
	Total           float64 `json:"Total"`
	SaleID          int     `json:"SaleId"`
}

type csharpDate struct {
	time.Time
}

func (c *csharpDate) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}

	var err error
	c.Time, err = time.Parse(`"2006-01-02T15:04:05"`, string(b))
	return err
}

func (c *csharpDate) MarshalJSON() ([]byte, error) {
	return []byte(c.Time.Format(`"2006-01-02T15:04:05"`)), nil
}
