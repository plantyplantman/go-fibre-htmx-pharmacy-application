package parser_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/plantyplantman/bcapi/pkg/parser"
	"github.com/plantyplantman/bcapi/pkg/report"
)

func TestParser__prlwgp(t *testing.T) {

	f, err := os.Open(`c:\Users\admin\Develin Management Dropbox\Zihan\files\in\231025\231025__petrie__prlwgp.TXT`)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	p, err := parser.NewParser(f)
	if err != nil {
		t.Fatal(err)
	}

	prl := report.ProductRetailList{
		Lines: []*report.ProductRetailListLine{},
	}
	if err = p.Parse(&prl.Lines); err != nil {
		t.Fatal(err)
	}

	for _, l := range prl.Lines {
		fmt.Printf("\n%+v", l)
	}

}

func TestParser__psl(t *testing.T) {
	f, err := os.Open(`c:\Users\admin\Develin Management Dropbox\Zihan\files\in\231025\231025__bunda__sts.TXT`)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	p, err := parser.NewParser(f)
	if err != nil {
		t.Fatal(err)
	}

	sts := report.ProductStockList{
		Lines: []*report.ProductStockListLine{},
	}
	if err = p.Parse(&sts.Lines); err != nil {
		t.Fatal(err)
	}

	for _, l := range sts.Lines {
		fmt.Printf("\n%+v", l)
	}
}

func TestParser__ms(t *testing.T) {
	f, err := os.Open(`C:\Users\admin\Develin Management Dropbox\Zihan\files\in\231018\231018__ms__web.TXT`)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	p, err := parser.NewParser(f, parser.IsMultistore(true))
	if err != nil {
		t.Fatal(err)
	}

	ms := map[string]*report.ProductRetailList{}
	if err = p.Parse(&ms); err != nil {
		t.Fatal(err)
	}

	if len(ms) != 3 {
		t.Fatalf("expected 3 types, got %d", len(ms))
	}

	for k := range ms {
		if len(ms[k].Lines) < 1 {
			t.Fatalf("expected at least 1 line, got %d", len(ms[k].Lines))
		}
		fmt.Printf("\n\nType: %s", k)
		fmt.Printf("\n\nLen: %v", len(ms[k].Lines))
		fmt.Printf("\nFirst: %+v", ms[k].Lines[0])
		fmt.Printf("\nLast: %+v", ms[k].Lines[len(ms[k].Lines)-1])
	}
}

func TestParser_pf(t *testing.T) {
	f, err := os.Open(`c:\Users\admin\Develin Management Dropbox\Zihan\files\in\231030\231030__web__pf.tsv`)
	if err != nil {
		t.Fatal(err)
	}
	p, err := parser.NewParser(f)
	if err != nil {
		t.Fatal(err)
	}
	r := report.ProductFile{}
	if err = p.Parse(&r.Lines); err != nil {
		t.Fatal(err)
	}
	if len(r.Lines) < 1 {
		t.Fatalf("expected at least 1 line, got %d", len(r.Lines))
	}

	for i, l := range r.Lines {
		if i == 0 {
			continue
		}
		if l.Id == "" {
			t.Fatalf("expected id to be set, got: %s\tLine: %d", l.Id, i)
		}
		if l.IsVariant {
			ids := strings.Split(l.Id, "/")
			if len(ids) != 2 {
				t.Fatalf("expected at 2 ids, got: %d\tLine: %d", len(ids), i)
			}
			if ids[0] == "" || ids[1] == "" {
				t.Fatalf("expected ids to be set, got: %s\tLine: %d", l.Id, i)
			}
		}
	}
}

func TestParser__nosr(t *testing.T) {
	f, err := os.Open(`C:\Users\admin\Develin Management Dropbox\Zihan\files\in\231009__MS__actions.txt`)
	if err != nil {
		t.Fatal(err)
	}

	p, err := parser.NewParser(f)
	if err != nil {
		t.Fatal(err)
	}

	nosr := report.NotOnSiteReport{}
	if err = p.Parse(&nosr); err != nil {
		t.Fatal(err)
	}

	if len(nosr) < 1 {
		t.Fatalf("expected at least 1 line, got %d", len(nosr))
	}

	for _, l := range nosr {
		fmt.Printf("\n%+v", l)
	}
}
