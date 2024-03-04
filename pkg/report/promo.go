package report

import "encoding/xml"

type Campaigns struct {
	XMLName  xml.Name `xml:"Campaigns" json:"campaigns,omitempty"`
	Text     string   `xml:",chardata" json:"text,omitempty"`
	Campaign Campaign `xml:"Campaign"`
}

type Campaign struct {
	Text         string `xml:",chardata" json:"text,omitempty"`
	CampaignName string `xml:"CampaignName"`
	Offers       Offers `xml:"Offers"`
}

type Offers struct {
	Text  string  `xml:",chardata" json:"text,omitempty"`
	Offer []Offer `xml:"Offer"`
}

type Offer struct {
	Text                string   `xml:",chardata" json:"text,omitempty"`
	OfferCode           string   `xml:"OfferCode"`
	OfferName           string   `xml:"OfferName"`
	OfferDesc           string   `xml:"OfferDesc"`
	OfferType           string   `xml:"OfferType"`
	IsLoyalty           string   `xml:"IsLoyalty"`
	StartDate           string   `xml:"StartDate"`
	EndDate             string   `xml:"EndDate"`
	ToDelete            string   `xml:"ToDelete"`
	MultipleRedemptions string   `xml:"MultipleRedemptions"`
	DollarOffDisc       string   `xml:"DollarOffDisc"`
	PercentOffDisc      string   `xml:"PercentOffDisc"`
	DollarThreshold     string   `xml:"DollarThreshold"`
	MultiBuyXQty        string   `xml:"MultiBuyXQty"`
	MultiBuyYQty        string   `xml:"MultiBuyYQty"`
	MultiBuyZQty        string   `xml:"MultiBuyZQty"`
	MultiBuyYDollar     string   `xml:"MultiBuyYDollar"`
	MultiBuyXDollar     string   `xml:"MultiBuyXDollar"`
	MultiBuyZDollar     string   `xml:"MultiBuyZDollar"`
	DivideDiscount      string   `xml:"DivideDiscount"`
	QtyThreshold        string   `xml:"QtyThreshold"`
	IsDiscountable      string   `xml:"IsDiscountable"`
	Products            Products `xml:"Products"`
}

type Products struct {
	Text    string    `xml:",chardata" json:"text,omitempty"`
	Product []Product `xml:"Product"`
}

type Product struct {
	Text           string  `xml:",chardata" json:"text,omitempty"`
	EAN            string  `xml:"EAN"`
	MNPN           string  `xml:"MNPN"`
	ProductName    string  `xml:"ProductName"`
	Cost           string  `xml:"Cost"`
	OfferPrice     float64 `xml:"OfferPrice"`
	QualifyingItem string  `xml:"QualifyingItem"`
	DiscountItem   string  `xml:"DiscountItem"`
	ProjSalesQty   string  `xml:"ProjSalesQty"`
}
