package sales_test

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/plantyplantman/bcapi/pkg/entities"
	"github.com/plantyplantman/bcapi/pkg/sales"
)

func TestFactory2(t *testing.T) {
	s, err := sales.NewService(4302, "https://localhost:44350")
	if err != nil {
		t.Fatal(err)
	}

	f := sales.NewReportFactory(s)

	from, err := time.Parse("2006-01-02 15:04", "2022-11-25 12:00")
	if err != nil {
		t.Fatal(err)
	}

	to, err := time.Parse("2006-01-02 15:04", "2022-11-25 22:00")
	if err != nil {
		t.Fatal(err)
	}

	summaryReport, err := f.GenerateReport(from, to)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("SUMMARY REPORT: %+v\n", summaryReport)

	hourlyReport, err := f.GenerateReports(from, to, time.Hour*1)
	if err != nil {
		log.Fatalln(err)
	}

	for _, report := range hourlyReport {
		t.Logf("%+v\n", *report)
	}

}

func TestWtf(t *testing.T) {
	// wg := &sync.WaitGroup{}
	from, err := time.Parse("2006-01-02T15:04:05", "2022-11-25T08:00:00")
	if err != nil {
		t.Fatal(err)
	}
	to, err := time.Parse("2006-01-02T15:04:05", "2022-11-25T22:00:00")
	if err != nil {
		t.Fatal(err)
	}

	if !from.Before(to) {
		t.Fail()
	}

	var interval time.Duration = 1 * time.Hour

	wg := &sync.WaitGroup{}
	strCh := make(chan string, 1)
	defer close(strCh)
	for from.Before(to.Add(interval)) {
		fmt.Println(from.Format("2006-01-02T15:04:05"), to.Format("2006-01-02T15:04:05"))
		fromCopy := from

		wg.Add(1)
		go func(from, to time.Time) {
			time.Sleep(1 * time.Second)
			strCh <- fmt.Sprintf("from: %s\tto: %s\n", from.Format("2006-01-02T15:04:05"), to.Format("2006-01-02T15:04:05"))
			wg.Done()
		}(fromCopy, fromCopy.Add(interval))

		from = from.Add(interval)
	}

	wg.Wait()

	for str := range strCh {
		fmt.Printf("%s", str)
	}

}

func TestHi(t *testing.T) {
	from, _ := time.Parse("2006-01-02 15:04", "2022-11-25 08:00")
	to, _ := time.Parse("2006-01-02 15:04", "2022-11-25 22:00")
	URL, _ := url.Parse("https://localhost:44350/api/4302")

	URL = URL.JoinPath("sales")
	params := URL.Query()
	params.Add("from", from.Format("2006-01-02T15:04:05"))
	params.Add("to", to.Format("2006-01-02T15:04:05"))

	URL.RawQuery = params.Encode()

	t.Logf("URL: %s\n", URL.String())
}

func TestFactory(t *testing.T) {
	s, err := sales.NewService(4302, "https://localhost:44350")
	if err != nil {
		t.Fatal(err)
	}

	f := sales.NewReportFactory(s)

	from, err := time.Parse("2006-01-02 15:04", "2022-11-25 12:00")
	if err != nil {
		t.Fatal(err)
	}

	to, err := time.Parse("2006-01-02 15:04", "2022-11-25 22:00")
	if err != nil {
		t.Fatal(err)
	}

	reportChan := make(chan *sales.Report, 1)
	defer close(reportChan)
	errChan := make(chan error, 1)
	defer close(errChan)
	wg := &sync.WaitGroup{}

	var interval = time.Duration(time.Hour * 1)
	for from.Before(to.Add(interval)) {
		fromCopy := from
		wg.Add(1)
		go func(from, to time.Time) {
			report, err := f.GenerateReport(from, to)
			if err != nil {
				errChan <- err
			}
			reportChan <- report
			wg.Done()
		}(fromCopy, fromCopy.Add(interval))

		from = from.Add(time.Hour)
	}

	wg.Wait()

Loop:
	for {
		select {
		case err, ok := <-errChan:
			if ok {
				t.Fatal(err)
			}
		case report, ok := <-reportChan:
			if ok {
				t.Logf("%+v\n", report)
			} else {
				t.Fail()
			}
		default:
			if len(reportChan) == 0 && len(errChan) == 0 {
				break Loop
			}
		}
	}
}

func TestGetSales(t *testing.T) {

	s, err := sales.NewService(4302, "https://localhost:44350")
	if err != nil {
		t.Fatal(err)
	}

	from, err := time.Parse("2006-01-02", "2016-01-01")
	if err != nil {
		t.Fatal(err)
	}

	to, err := time.Parse("2006-01-02", "2016-01-03")
	if err != nil {
		t.Fatal(err)
	}

	sales, err := s.GetSales(from, to)
	if err != nil {
		t.Fatal(err)
	}
	for _, sale := range sales {
		fmt.Printf("sale.Date: %s", sale.Date)
	}
}

type Sales []*entities.Sale

func (s Sales) TopProductsByRevenue() []string {
	top := make(map[string]float64)
	for _, sale := range s {
		for _, item := range sale.SaleLines {
			top[item.ProductID+item.Name] += item.Total
		}
	}

	return mapToSortedSlice(top)
}

func mapToSortedSlice(m map[string]float64 /*, sortFunc*/) []string {

	type kv struct {
		key   string
		value float64
	}

	var ss []kv
	for k, v := range m {
		ss = append(ss, kv{k, v})
	}

	slices.SortFunc(ss, func(a, b kv) int {
		if a.value > b.value {
			return -1
		}
		if a.value < b.value {
			return 1
		}
		return 0
	})

	var topKeys []string
	for _, kv := range ss {
		topKeys = append(topKeys, kv.key)
	}
	return topKeys
}

func (s Sales) TopProductsByNumberSold() []string {
	top := make(map[string]float64)
	for _, sale := range s {
		for _, item := range sale.SaleLines {
			top[item.ProductID+item.Name] += item.Total
		}
	}

	return mapToSortedSlice(top)
}

func (s Sales) TotalRevenue() float64 {
	var total float64
	for _, sale := range s {
		total += sale.Total
	}
	return total
}

func TestGggg(t *testing.T) {
	from := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC)
	url := "https://localhost:44350/api/4302"
	resp, err := http.Get(fmt.Sprintf("%s/sales?from=%s&to=%s", url, from.Format("2006-01-02"), to.Format("2006-01-02")))
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("b: %s\n", b)

	var data []*entities.Sale
	err = json.Unmarshal(b, &data)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("data: %+v", data)
}
