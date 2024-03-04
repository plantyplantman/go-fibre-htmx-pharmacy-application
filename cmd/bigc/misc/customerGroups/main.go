package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/plantyplantman/bcapi/pkg/bigc"
)

func main() {
	f, err := os.Open(`C:\Users\admin\Develin Management Dropbox\Zihan\files\in\231114\231110__web__calmoseptine-orders.csv`)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	r := csv.NewReader(f)

	var customerIds = map[int]struct{}{}
	for {
		line, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalln(err)
		}

		id, err := strconv.Atoi(line[1])
		if err != nil {
			continue
		}
		customerIds[id] = struct{}{}
	}

	c := bigc.MustGetClient()

	for id := range customerIds {
		cuss, err := c.GetCustomers(map[string]string{"id:in": fmt.Sprint(id)})
		if err != nil {
			log.Println(err)
		}
		if len(cuss) != 1 {
			log.Println("len(cuss) != 1", len(cuss))
			continue
		}
		cus, err := c.UpdateCustomer(cuss[0], id, bigc.WithCustomerGroupID(6))
		if err != nil {
			log.Println(err)
			continue
		}

		if cus.CustomerGroupID != 6 {
			log.Println("failed to update customer with id: ", id)
			continue
		}
	}
}
