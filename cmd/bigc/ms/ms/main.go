package main

import (
	"encoding/csv"
	"log"
	"os"
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
	fp := `C:\Users\admin\Develin Management Dropbox\Zihan\files\in\231105\231105 MS Web.TXT`
	date, err := time.Parse("060102", "231105")
	if err != nil {
		log.Fatalln(err)
	}
	f, err := os.Open(fp)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = '\t'
	r.LazyQuotes = true

	prls, err := parser.ParseMultistoreInput(r, date, "ms")
	if err != nil {
		log.Fatalln(err)
	}

	var connString = env.TEST_NEON
	if connString == "" {
		log.Fatalln("TEST_NEON_CONNECTION_STRING not set")
	}
	DB, err := gorm.Open(postgres.Open(connString), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}

	repo := product.NewRepository(DB)
	service := product.NewService(repo)
	c := bigc.MustGetClient()

	var nosr = report.DoMultistore(prls, service, c)
	of, err := os.OpenFile("231105__ms__not-on-site.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}

	err = gocsv.MarshalFile(&nosr, of)
	if err != nil {
		log.Fatalln(err)
	}
}
