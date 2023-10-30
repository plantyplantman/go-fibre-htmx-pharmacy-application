package main

import (
	"log"
	"os"

	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/plantyplantman/bcapi/pkg/env"
	"github.com/plantyplantman/bcapi/pkg/parser"
	"github.com/plantyplantman/bcapi/pkg/product"
	"github.com/plantyplantman/bcapi/pkg/report"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	f, err := os.Open(`C:\Users\admin\Develin Management Dropbox\Zihan\files\in\231018-231025__MS__actions.txt`)
	if err != nil {
		log.Fatal(err)
	}

	p, err := parser.NewParser(f)
	if err != nil {
		log.Fatal(err)
	}

	nosr := report.NotOnSiteReport{}
	if err = p.Parse(&nosr); err != nil {
		log.Fatal(err)
	}

	if len(nosr) < 1 {
		log.Fatalf("expected at least 1 line, got %d", len(nosr))
	}

	var connString = env.TEST_NEON
	if connString == "" {
		log.Fatalln("TEST_NEON_CONNECTION_STRING not set")
	}
	DB, err := gorm.Open(postgres.Open(connString), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}

	r := product.NewRepository(DB)
	s := product.NewService(r)
	c := bigc.MustGetClient()
	ai_c, err := bigc.GetOpenAiClient()
	if err != nil {
		log.Fatal(err)
	}
	nosr.Action(s, c, &ai_c)
}
