package report

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Report struct {
	Date   time.Time
	Source string
	Store  string
}

type ProductStockList struct {
	Report
	Lines []*ProductStockListLine
}

type ProductStockListLine struct {
	Sku      string `csv:"Barcode"`
	ProdName string `csv:"Product Name"`
	Price    Price  `csv:"Retail"`
	Qty      int    `csv:"Qty O/H"`
}

type ProductStockLists []*ProductStockList

func (psls ProductStockLists) Combine() *CombinedProductStockList {
	skuMap := make(map[string]*CombinedProductStockListLine)

	for _, psl := range psls {
		for _, line := range psl.Lines {
			sku := strings.TrimSpace(line.Sku)
			if _, exists := skuMap[sku]; !exists {
				skuMap[sku] = &CombinedProductStockListLine{
					Sku:      sku,
					ProdName: line.ProdName,
					Price:    line.Price,
				}
			}
			switch psl.Store {
			case "petrie":
				skuMap[sku].Petrie = line.Qty
			case "franklin":
				skuMap[sku].Franklin = line.Qty
			case "bunda":
				skuMap[sku].Bunda = line.Qty
			case "con":
				skuMap[sku].Con = line.Qty
			}
			skuMap[sku].Total = skuMap[sku].Total + roundNegativeToZero(line.Qty)
		}
	}

	var combinedLines []*CombinedProductStockListLine
	for _, line := range skuMap {
		combinedLines = append(combinedLines, line)
	}

	return &CombinedProductStockList{
		Report: Report{
			Date:   psls[0].Date,
			Source: psls[0].Source,
			Store:  "ms",
		},
		Lines: combinedLines,
	}
}

type CombinedProductStockList struct {
	Report
	Lines []*CombinedProductStockListLine
}

type CombinedProductStockListLine struct {
	Sku      string `csv:"Barcode"`
	ProdName string `csv:"Product Name"`
	Price    Price  `csv:"Retail"`
	Petrie   int    `csv:"Petrie"`
	Bunda    int    `csv:"Bunda"`
	Con      int    `csv:"Con"`
	Franklin int    `csv:"Franklin"`
	Web      int    `csv:"Web"`
	Total    int    `csv:"Total"`
}

func (r *CombinedProductStockList) ToSkuMap() map[string]*CombinedProductStockListLine {
	m := make(map[string]*CombinedProductStockListLine, len(r.Lines))
	for _, line := range r.Lines {
		m[line.Sku] = line
	}
	return m
}

type ProductFile struct {
	Report
	Lines []*ProductFileLine
}

type ProductFileLine struct {
	Id        string         `csv:"Id"`
	Sku       BigCommerceSku `csv:"Sku"`
	Name      string         `csv:"Name"`
	Price     Price          `csv:"Price"`
	Soh       int            `csv:"Soh"`
	IsVariant bool           `csv:"IsVariant"`
}

func (r *ProductFile) ToSkuMap() map[string]*ProductFileLine {
	m := make(map[string]*ProductFileLine, len(r.Lines))
	for _, line := range r.Lines {
		m[line.Sku.String()] = line
	}
	return m
}

func NewBigCommerceSku(s string) BigCommerceSku {
	return BigCommerceSku{string: s}
}

type BigCommerceSku struct {
	string
}

func (s BigCommerceSku) String() string {
	sku, err := cleanBigCommerceSku(s.string)
	if err != nil {
		return s.string
	}
	return sku
}

func (s BigCommerceSku) MarshalCSV() (string, error) {
	return fmt.Sprint(s), nil
}

func (s *BigCommerceSku) UnmarshalCSV(csv string) (err error) {
	sku, err := cleanBigCommerceSku(csv)
	if err != nil {
		log.Println(err)
		s.string = ""
		return nil
	}
	s.string = sku
	return nil
}

type Price struct {
	float64
}

func NewPrice(f float64) Price {
	return Price{float64: f}
}

func (p *Price) MarshalCSV() (string, error) {
	return fmt.Sprint(p), nil
}

func (p *Price) UnmarshalCSV(csv string) (err error) {
	p.float64, err = parseFloat(csv)
	return err
}

func (p *Price) Float64() float64 {
	return p.float64
}

func (p Price) String() string {
	return fmt.Sprintf("%.2f", p.float64)
}

// UTILS

func parseFloat(num string) (float64, error) {
	num = removeCommaFromNumber(num)

	return strconv.ParseFloat(strings.TrimSpace(num), 64)
}

func removeCommaFromNumber(num string) string {
	return strings.ReplaceAll(num, ",", "")
}

func roundNegativeToZero(n int) int {
	if n < 0 {
		return 0
	}
	return n
}

func cleanBigCommerceSku(sku string) (string, error) {
	pattern := `^\/{0,3}(\d+)(?:-[^-]*)?`
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	sku = strings.TrimSpace(sku)
	matches := re.FindStringSubmatch(sku)
	if len(matches) < 1 {
		return "", errors.New("Invalid sku: " + sku)
	}

	s := matches[1]

	for strings.HasPrefix(s, "0") {
		s = strings.TrimPrefix(s, "0")
	}

	return s, nil
}
