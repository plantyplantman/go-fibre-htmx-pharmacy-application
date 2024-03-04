package main

import (
	"fmt"
	"log"

	"github.com/plantyplantman/bcapi/pkg/bigc"
)

func main() {
	c := bigc.MustGetClient()

	ps, err := c.GetAllProducts(map[string]string{"is_visible": "true"})
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(delimit('\t', "sku", "name"))
	for _, p := range ps {
		if p.Sku == "" {
			for _, v := range p.Variants {
				if v.Gtin == "" {
					fmt.Println(delimit('\t', v.Sku, fmt.Sprintf("%s - %s", p.Name, v.OptionValues[0].OptionDisplayName)))
				}
			}
		} else {
			if p.Gtin == "" {
				fmt.Println(delimit('\t', p.Sku, p.Name))
			}
		}
	}
}

func delimit(delimiter rune, ss ...any) string {
	var s string
	for i, v := range ss {
		if i == 0 {
			s = fmt.Sprint(v)
			continue
		}
		s = s + string(delimiter) + fmt.Sprint(v)
	}
	return s
}
