package sales

import (
	"fmt"
	"sort"
	"time"

	"github.com/plantyplantman/bcapi/pkg/entities"
	"github.com/samber/lo"
)

type Report struct {
	Start               time.Time
	End                 time.Time
	CustomerCount       int
	ProductsSold        int
	GrossSales          float64
	Discounts           float64
	GST                 float64
	NetSales            float64
	COGS                float64
	GPDollar            float64
	GPPercent           float64
	DollarPerCustomer   float64
	ProductsPerCustomer float64
}

type ReportFactory interface {
	GenerateReport(time.Time, time.Time) (*Report, error)
	GenerateReports(time.Time, time.Time, time.Duration) ([]*Report, error)
}

type reportFactory struct {
	service Service
}

func NewReportFactory(service Service) ReportFactory {
	return &reportFactory{
		service: service,
	}
}

func (rf *reportFactory) GenerateReport(from, to time.Time) (*Report, error) {
	if !from.Before(to) {
		return nil, fmt.Errorf("bad request\tfrom before to\tfrom: %v\tto: %v", from, to)
	}

	saleSlice, err := rf.service.GetSales(from, to)
	if err != nil {
		return nil, err
	}

	return saleSliceToReport(saleSlice), nil
}

func (rf *reportFactory) GenerateReports(from, to time.Time, interval time.Duration) ([]*Report, error) {
	if !from.Before(to) {
		return nil, fmt.Errorf("bad request\t`from` after `to`\tfrom: %v\tto: %v", from, to)
	}

	if interval > to.Sub(from) {
		return nil, fmt.Errorf("bad request\tinterval > from-to\tinterval: %v\tfrom: %v\tto: %v", interval, from, to)
	}

	sales, err := rf.service.GetSales(from, to)
	if err != nil {
		return nil, err
	}

	var reports []*Report
	for start := from; start.Before(to); start = start.Add(interval) {
		end := start.Add(interval)
		if end.After(to) {
			end = to
		}

		var intervalSales []*entities.Sale
		for _, sale := range sales {
			if sale.Date.After(start) && sale.Date.Before(end) {
				intervalSales = append(intervalSales, sale)
			}
		}

		report := saleSliceToReport(intervalSales)
		if report == nil {
			continue
		}
		reports = append(reports, report)
	}

	sort.Slice(reports, func(i, j int) bool {
		return reports[i].Start.Before(reports[j].Start)
	})

	return reports, nil
}

func saleSliceToReport(saleSlice []*entities.Sale) *Report {
	if saleSlice == nil {
		return nil
	}
	var dst = &Report{
		Start: lo.Reduce(saleSlice, func(rv time.Time, s *entities.Sale, _ int) time.Time {
			if s.Date.Before(rv) {
				return s.Date.Time
			}
			return rv
		}, time.Now().Add(time.Hour*24*365*100)),
		End: lo.Reduce(saleSlice, func(rv time.Time, s *entities.Sale, _ int) time.Time {
			if s.Date.After(rv) {
				return s.Date.Time
			}
			return rv
		}, saleSlice[0].Date.Time),
	}

	dst.CustomerCount = len(saleSlice)
	for _, sale := range saleSlice {
		dst.GrossSales += sale.Total

		dst.Discounts -= sale.Discount
		dst.GST -= sale.Gst

		for _, line := range sale.SaleLines {
			dst.ProductsSold += int(line.Quantity)
			dst.COGS += line.Cogs
			dst.NetSales += line.Total - line.Discount - line.PromoDiscount - line.Gst
		}
	}

	dst.GPDollar = dst.NetSales - dst.COGS
	dst.GPPercent = (dst.GPDollar / dst.GrossSales) * 100
	dst.DollarPerCustomer = dst.NetSales / float64(dst.CustomerCount)
	dst.ProductsPerCustomer = float64(dst.ProductsSold) / float64(dst.CustomerCount)
	return dst
}
