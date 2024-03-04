package main

import (
	"fmt"

	"github.com/plantyplantman/bcapi/pkg/bigc"
)

func main() {
	bigc := bigc.MustGetClient()

	p, err := bigc.GetProductFromSku("5010724537503")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", p.BrandID)

}
