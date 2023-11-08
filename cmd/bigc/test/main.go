package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/plantyplantman/bcapi/pkg/parser"
	"github.com/plantyplantman/bcapi/pkg/report"
	"github.com/samber/lo"
)

func main() {
	f, err := os.Open(`C:\Users\admin\Desktop\hi`)
	if err != nil {
		log.Panicln(err)
	}
	defer f.Close()
	csvr := csv.NewReader(f)

	skus := make([]string, 0)
	for {
		l, err := csvr.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Panicln(err)
		}
		skus = append(skus, strings.TrimSpace(l[0]))
	}

	f, err = os.Open(`C:\Users\admin\Desktop\out.TXT`)
	if err != nil {
		log.Panicln(err)
	}
	defer f.Close()
	parser, err := parser.NewCsvParser(f)
	if err != nil {
		log.Panicln(err)
	}
	var sts = report.ProductStockList{}
	if err = parser.Parse(&sts.Lines); err != nil {
		log.Panicln(err)
	}

	if len(sts.Lines) == 0 {
		log.Panicln("no lines")
	}

	var stslMap = lo.Associate(sts.Lines, func(l *report.ProductStockListLine) (string, *report.ProductStockListLine) {
		return strings.TrimSpace(l.Sku), l
	})

	for _, sku := range skus {
		if l, ok := stslMap[sku]; ok {
			fmt.Println(l.ProdName, l.Qty)
		} else {
			fmt.Println("not found")
		}
	}
}
