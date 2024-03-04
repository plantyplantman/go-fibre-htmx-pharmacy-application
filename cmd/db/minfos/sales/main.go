package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/plantyplantman/bcapi/pkg/entities"
	"github.com/samber/lo"
)

func main() {

	url := func(from string, to string) string {
		return fmt.Sprintf("https://localhost:44350/api/4302/sales?from=%s&to=%s", from, to)
	}

	// from=1999-12-31&to=2023-11-14

	from, err := time.Parse("2006-01-02", "1999-12-31")
	if err != nil {
		log.Fatal(err)
	}

	// to, err := time.Parse("2006-01-02", "2023-11-14")
	to, err := time.Parse("2006-01-02", "2000-01-31")
	if err != nil {
		log.Fatal(err)
	}

	var dt = from
	sales := []entities.Sale{}
	for dt.Before(to) {
		date := dt.Format("2006-01-02")
		fmt.Println("Getting sales for: ", date)
		resp, err := http.Get(url(date, date))
		if err != nil {
			log.Fatal(err)
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		if resp.StatusCode != 200 {
			log.Fatal("Status code: ", resp.StatusCode, string(body))
		}
		var s []entities.Sale
		if err := json.Unmarshal(body, &s); err != nil {
			log.Fatalln(err)
		}
		sales = append(sales, s...)
		dt = dt.AddDate(0, 0, 1)
	}

	fmt.Println("Total sales: ", len(sales))
	fmt.Println("Total lines sold: ", lo.Reduce(sales, func(agg int, s entities.Sale, _ int) int {
		return agg + len(s.SaleLines)
	}, 0))
}
