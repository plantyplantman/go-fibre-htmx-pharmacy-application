package main

import (
	"fmt"

	"github.com/plantyplantman/bcapi/pkg/bigc"
)

func main() {
	c := bigc.MustGetClient()

	cs, err := c.GetAllCategories(map[string]string{})
	if err != nil {
		panic(err)
	}

	for _, c := range cs {
		fmt.Println(c.ID, "\t", c.Name)
	}
}
