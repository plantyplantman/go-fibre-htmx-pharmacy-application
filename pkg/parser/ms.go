package parser

import (
	"encoding/csv"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/plantyplantman/bcapi/pkg/report"
)

func ParseMultistoreInput(r *csv.Reader, date time.Time, source string) (map[string]*report.ProductRetailList, error) {
	var prlMap = map[string]*report.ProductRetailList{
		"new":    {Report: report.Report{Date: date, Source: source, Store: "new"}},
		"edited": {Report: report.Report{Date: date, Source: source, Store: "edited"}},
		"clean":  {Report: report.Report{Date: date, Source: source, Store: "clean"}},
	}
	curr := "new"

	for {
		line, err := r.Read()
		// fmt.Println(curr)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			if errors.Is(err, csv.ErrFieldCount) {
				if len(line) > 1 {
					switch strings.ToLower(strings.TrimSpace(line[1])) {
					case "products - new":
						curr = "new"
					case "products - edited":
						curr = "edited"
					case "products - deleted permanently":
						curr = "clean"
					case "products - clean up stock":
						curr = "clean"
					default:
						continue
					}
					continue
				}
			} else {
				return nil, err
			}
		}
		relevantCols := []int{2, 4, 6, 8, 10}

		relevantFields := getRelevantFields(line, relevantCols)
		if len(relevantFields) < 1 {
			continue
		}
		mnpn := relevantFields[0]
		sku := relevantFields[1]
		prodNo := relevantFields[2]
		prodName := relevantFields[3]
		priceStr := relevantFields[4]

		mnpnInt, err := parseInt(mnpn)
		if err != nil {
			mnpnInt = 0
		}

		sku = strings.TrimSpace(sku)

		prodNoInt, err := parseInt(prodNo)
		if err != nil {
			prodNoInt = 0
		}

		prodName = strings.TrimSpace(prodName)
		priceFloat, err := parseFloat(priceStr)
		if err != nil {
			priceFloat = 0
		}
		price := report.NewPrice(priceFloat)

		prlMap[curr].Lines = append(
			prlMap[curr].Lines,
			&report.ProductRetailListLine{
				Mnpn:     mnpnInt,
				Sku:      sku,
				ProdNo:   prodNoInt,
				ProdName: prodName,
				Price:    price,
			},
		)
	}

	// fmt.Println("len prodMap: ", len(prlMap))

	return prlMap, nil
}

func parseFloat(num string) (float64, error) {
	num = removeCommaFromNumber(num)

	return strconv.ParseFloat(strings.TrimSpace(num), 64)
}

func parseInt(num string) (int, error) {
	num = removeCommaFromNumber(num)
	return strconv.Atoi(strings.TrimSpace(num))
}

func removeCommaFromNumber(num string) string {
	return strings.ReplaceAll(num, ",", "")
}

func getRelevantFields(data []string, relevantCols []int) []string {
	fltd := make([]string, 0, len(data))
	for i, e := range data {
		if slices.Contains(relevantCols, i) {
			fltd = append(fltd, e)
		}
	}
	return fltd
}
