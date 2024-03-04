package report_test

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/plantyplantman/bcapi/pkg/parser"
	"github.com/plantyplantman/bcapi/pkg/product"
	"github.com/plantyplantman/bcapi/pkg/report"
)

func TestNotOnSite(t *testing.T) {
	dt, err := time.Parse("060102", "231211")
	if err != nil {
		t.Fatal(err)
	}

	var ZihanFilesPath = `C:\Users\admin\Develin Management Dropbox\Zihan\files\`
	var inPath = filepath.Join(ZihanFilesPath, `in\`+`231211`, `231211 MS Web.TXT`)
	f, err := os.Open(inPath)
	if err != nil {
		t.Fatal(err)
	}

	r := csv.NewReader(f)
	r.Comma = '\t'
	r.LazyQuotes = true

	prls, err := parser.ParseMultistoreInput(r, dt, inPath)
	if err != nil {
		t.Fatal(err)
	}

	service, err := product.NewDefaultService()
	if err != nil {
		t.Fatal(err)
	}
	var nosr = report.NotOnSiteReport{}
	for _, prl := range prls {
		tmp, err := prl.NotOnSite(service)
		if err != nil {
			t.Fatal(err)
		}
		nosr = append(nosr, tmp...)
	}

	for _, l := range nosr {
		t.Log(l)
	}
}
