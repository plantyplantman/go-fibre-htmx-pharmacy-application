package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/plantyplantman/bcapi/pkg/env"
	"github.com/plantyplantman/bcapi/pkg/parser"
	"github.com/plantyplantman/bcapi/pkg/product"
	"github.com/plantyplantman/bcapi/pkg/report"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	var dates = []string{"240306", "240312"}
	// var out = report.NotOnSiteReport{}
	for _, d := range dates {
		nosr, err := action(d)
		if err != nil {
			log.Fatalln(err)
		}

		op := fmt.Sprintf(`C:\Users\admin\Develin Management Dropbox\Zihan\files\out\%s\%s__MS-Web__not-on-site.csv`, d, d)
		if err = export(op, nosr); err != nil {
			log.Println(err)
		}
	}
}

func export(path string, nosr report.NotOnSiteReport) error {
	if err := os.MkdirAll(filepath.Dir(path), 0770); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	return gocsv.Marshal(nosr, f)
}

func action(datestr string) (report.NotOnSiteReport, error) {
	date, err := time.Parse("060102", datestr)
	if err != nil {
		return nil, err
	}
	fp := fmt.Sprintf(`C:\Users\admin\Develin Management Dropbox\Zihan\files\in\%s\%s MS Web.TXT`, datestr, datestr)
	f, err := os.Open(fp)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = '\t'
	r.LazyQuotes = true

	prls, err := parser.ParseMultistoreInput(r, date, "ms")
	if err != nil {
		return nil, err
	}

	var connString = env.TEST_NEON
	if connString == "" {
		log.Fatalln("TEST_NEON_CONNECTION_STRING not set")
	}
	DB, err := gorm.Open(postgres.Open(connString), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	repo := product.NewRepository(DB)
	service := product.NewService(repo)
	c := bigc.MustGetClient()

	var nosr = report.DoMultistore(prls, service, c)

	return nosr, nil
}
