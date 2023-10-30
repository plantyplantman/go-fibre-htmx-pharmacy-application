package presenter

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/plantyplantman/bcapi/pkg/entities"
)

type Products []*Product

func (ps Products) ToTable(page, limit int) *fiber.Map {
	rows := make([][]string, 0)
	for _, p := range ps {
		rows = append(rows, p.ToTableRow())
	}
	return &fiber.Map{
		"Headers": []string{
			"Sku",
			"Name",
			"Price",
			"Cost Price",
			"Petrie",
			"Bunda",
			"Con",
			"Franklin",
			"Web",
			"Total",
			"BCID",
		},
		"Rows":  rows,
		"Page":  page,
		"Limit": limit,
	}
}

type Product struct {
	Sku             string           `json:"sku"`
	ProdName        string           `json:"product_name"`
	Price           float64          `json:"price"`
	CostPrice       float64          `json:"cost_price"`
	OnWeb           int              `json:"on_web"`
	IsVariant       bool             `json:"is_variant"`
	BCID            string           `json:"bcid"`
	StockInfomation StockInformation `json:"stock_information"`
}

type StockInformation struct {
	Petrie   int `json:"petrie"`
	Bunda    int `json:"bunda"`
	Con      int `json:"con"`
	Franklin int `json:"franklin"`
	Web      int `json:"web"`
	Total    int `json:"total"`
}

func (p *Product) FromEntity(ep *entities.Product) {
	p.Sku = ep.Sku
	p.ProdName = ep.Name
	p.Price = ep.Price
	p.CostPrice = ep.CostPrice
	p.OnWeb = ep.OnWeb
	p.BCID = ep.BCID
	p.IsVariant = ep.IsVariant

	for _, si := range ep.StockInformations {
		switch si.Location {
		case "petrie":
			p.StockInfomation.Petrie = si.Soh
		case "bunda":
			p.StockInfomation.Bunda = si.Soh
		case "con":
			p.StockInfomation.Con = si.Soh
		case "franklin":
			p.StockInfomation.Franklin = si.Soh
		case "web":
			p.StockInfomation.Web = si.Soh
		}
	}
}

func (p *Product) ToEntity() *entities.Product {
	return &entities.Product{
		Sku:       p.Sku,
		Name:      p.ProdName,
		Price:     p.Price,
		CostPrice: p.CostPrice,
		OnWeb:     p.OnWeb,
		IsVariant: p.IsVariant,
		BCID:      p.BCID,
		StockInformations: []entities.StockInformation{
			{
				Location: "petrie",
				Soh:      p.StockInfomation.Petrie,
			},
			{
				Location: "bunda",
				Soh:      p.StockInfomation.Bunda,
			},
			{
				Location: "con",
				Soh:      p.StockInfomation.Con,
			},
			{
				Location: "franklin",
				Soh:      p.StockInfomation.Franklin,
			},
			{
				Location: "web",
				Soh:      p.StockInfomation.Web,
			},
		},
	}
}

func (p *Product) ToTableRow() []string {
	return []string{
		p.Sku,
		p.ProdName,
		strconv.FormatFloat(p.Price, 'f', 2, 64),
		strconv.FormatFloat(p.CostPrice, 'f', 2, 64),
		strconv.Itoa(p.StockInfomation.Petrie),
		strconv.Itoa(p.StockInfomation.Bunda),
		strconv.Itoa(p.StockInfomation.Con),
		strconv.Itoa(p.StockInfomation.Franklin),
		strconv.Itoa(p.StockInfomation.Web),
		strconv.Itoa(p.StockInfomation.Total),
		p.BCID,
	}
}

func ProductSuccessResponse(data *Product) *fiber.Map {
	p := Product{
		Sku:       data.Sku,
		ProdName:  data.ProdName,
		Price:     data.Price,
		CostPrice: data.CostPrice,
		OnWeb:     data.OnWeb,
		IsVariant: data.IsVariant,
		BCID:      data.BCID,
		StockInfomation: StockInformation{
			Petrie:   data.StockInfomation.Petrie,
			Bunda:    data.StockInfomation.Bunda,
			Con:      data.StockInfomation.Con,
			Franklin: data.StockInfomation.Franklin,
			Web:      data.StockInfomation.Web,
			Total:    data.StockInfomation.Total,
		},
	}

	return &fiber.Map{
		"status": true,
		"data":   p,
		"error":  nil,
	}
}

func ProductsSuccessResponse(data []*Product) *fiber.Map {
	ps := []Product{}
	for _, p := range data {
		ps = append(ps, Product{
			Sku:       p.Sku,
			ProdName:  p.ProdName,
			Price:     p.Price,
			CostPrice: p.CostPrice,
			OnWeb:     p.OnWeb,
			IsVariant: p.IsVariant,
			BCID:      p.BCID,
			StockInfomation: StockInformation{
				Petrie:   p.StockInfomation.Petrie,
				Bunda:    p.StockInfomation.Bunda,
				Con:      p.StockInfomation.Con,
				Franklin: p.StockInfomation.Franklin,
				Web:      p.StockInfomation.Web,
				Total:    p.StockInfomation.Total,
			},
		})
	}

	return &fiber.Map{
		"status": true,
		"data":   ps,
		"error":  nil,
	}
}

func ProductErrorResponse(err error) *fiber.Map {
	return &fiber.Map{
		"status": false,
		"data":   "",
		"error":  err.Error(),
	}
}
