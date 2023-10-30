package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/plantyplantman/bcapi/pkg/db"
	"github.com/plantyplantman/bcapi/pkg/entities"
	"github.com/plantyplantman/bcapi/pkg/env"
	"github.com/plantyplantman/bcapi/pkg/parser"
	"github.com/plantyplantman/bcapi/pkg/report"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	var connString = env.TEST_NEON
	if connString == "" {
		log.Fatalln("TEST_NEON_CONNECTION_STRING not set")
	}

	// var ZihanFilesPath = `/mnt/c/Users/admin/Develin Management Dropbox/Zihan/files/`
	// var psls report.ProductStockLists = parseProductStockList(filepath.Join(ZihanFilesPath, `in/231025`))
	// prl := mustParseProductRetailList(filepath.Join(ZihanFilesPath, `in/231025/231025__petrie__prlwgp.TXT`))
	// pf := mustParseProductFile(filepath.Join(ZihanFilesPath, `in/231025/231025__web__pf.tsv`))

	var ZihanFilesPath = `c:\Users\admin\Develin Management Dropbox\Zihan\files\`
	var psls report.ProductStockLists = parseProductStockList(filepath.Join(ZihanFilesPath, `in\231030`))
	if len(psls) != 4 {
		log.Fatal("wrong number of psls, expected 4, got", len(psls))
	}
	prl := mustParseProductRetailList(filepath.Join(ZihanFilesPath, `in\231030\231030__petrie__prlwgp.TXT`))
	pf := mustParseProductFile(filepath.Join(ZihanFilesPath, `in\231030\231030__web__pf.tsv`))

	products := NewProducts(psls, prl, pf)

	DB, err := gorm.Open(postgres.Open(connString), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	if err = db.Migrate(DB); err != nil {
		log.Fatalln(err)
	}
	if err = db.Seed(DB, products); err != nil {
		log.Fatalln(err)
	}

	p := &entities.Product{Sku: "1234"}
	if err := DB.First(p); err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("%+v", p)
}

func NewProducts(psls []*report.ProductStockList, prl *report.ProductRetailList, pf *report.ProductFile) []*entities.Product {
	var psls2 report.ProductStockLists = psls
	cpsl := psls2.Combine()

	prlM := prl.ToSkuMap()
	pfM := pf.ToSkuMap()

	var products []*entities.Product
	for _, l := range cpsl.Lines {
		var sku = strings.TrimSpace(l.Sku)

		var (
			webSoh    = 0
			onWeb     = 0
			isVariant = false
			bcid      = ""
		)
		if _, ok := pfM[sku]; ok {
			webSoh = pfM[sku].Soh
			onWeb = 1
			isVariant = pfM[sku].IsVariant
			bcid = pfM[sku].Id
		}

		var (
			price     = 0.0
			costPrice = 0.0
		)
		if _, ok := prlM[sku]; ok {
			price = prlM[sku].Price.Float64()
			costPrice = prlM[sku].Cost.Float64()
		}

		p := &entities.Product{
			Sku:       sku,
			Name:      l.ProdName,
			Price:     price,
			CostPrice: costPrice,
			OnWeb:     onWeb,
			IsVariant: isVariant,
			BCID:      bcid,
			StockInformations: []entities.StockInformation{
				{
					Location: "petrie",
					Soh:      l.Petrie,
				},
				{
					Location: "franklin",
					Soh:      l.Franklin,
				},
				{
					Location: "bunda",
					Soh:      l.Bunda,
				},
				{
					Location: "con",
					Soh:      l.Con,
				},
				{
					Location: "web",
					Soh:      webSoh,
				},
			},
		}

		products = append(products, p)
	}
	return products
}

func mustParseProductFile(path string) *report.ProductFile {
	var (
		f   *os.File
		err error
	)
	if f, err = os.Open(path); err != nil {
		log.Fatal(err)
	}

	var p parser.Parser
	if p, err = parser.NewParser(f); err != nil {
		log.Fatal(err)
	}

	storeRegex := `(\d{6})__(.*?)__(.*?).(.*?)`
	re := regexp.MustCompile(storeRegex)

	var date time.Time
	if date, err = time.Parse("060102", re.FindStringSubmatch(path)[1]); err != nil {
		log.Fatal(err)
	}

	var pf = report.ProductFile{
		Report: report.Report{
			Date:   date,
			Source: path,
			Store:  "web",
		},
		Lines: []*report.ProductFileLine{},
	}

	if err = p.Parse(&pf.Lines); err != nil {
		log.Fatal(err)
	}

	return &pf
}

func mustParseProductRetailList(path string) *report.ProductRetailList {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	storeRegex := `(\d{6})__(.*?)__(.*?).(.*?)`
	re := regexp.MustCompile(storeRegex)

	var date time.Time
	if date, err = time.Parse("060102", re.FindStringSubmatch(path)[1]); err != nil {
		log.Fatal(err)
	}

	p, err := parser.NewParser(f)
	if err != nil {
		log.Fatal(err)
	}

	prl := report.ProductRetailList{
		Report: report.Report{
			Date:   date,
			Source: path,
			Store:  re.FindStringSubmatch(path)[2],
		},
		Lines: []*report.ProductRetailListLine{},
	}

	if err := p.Parse(&prl.Lines); err != nil {
		log.Fatal(err)
	}

	return &prl
}

func parseProductStockList(path string) []*report.ProductStockList {
	var (
		fs  []string
		err error
	)
	globQ := filepath.Join(path, `*sts.TXT`)
	if fs, err = filepath.Glob(globQ); err != nil {
		log.Fatal(err)
	}
	if len(fs) == 0 {
		log.Fatal("no sts files found")
	}

	var date time.Time
	if date, err = maybeDate(path); err != nil {
		log.Fatal(err)
	}

	storeRegex := `(\d{6})__(.*?)__(.*?).(.*?)`
	re := regexp.MustCompile(storeRegex)

	var psls = make([]*report.ProductStockList, len(fs))
	for i := range fs {
		psls[i] = &report.ProductStockList{
			Report: report.Report{
				Date:   date,
				Source: fs[i],
				Store:  re.FindStringSubmatch(fs[i])[2],
			},
			Lines: []*report.ProductStockListLine{}}
	}

	wg := sync.WaitGroup{}
	for i := range fs {
		wg.Add(1)

		pslch := make(chan *report.ProductStockList, len(psls))
		defer close(pslch)
		pslche := make(chan error, 1)
		defer close(pslche)

		go func(i int) {
			var file *os.File
			file, err := os.Open(fs[i])
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			var p parser.Parser
			if p, err = parser.NewParser(file); err != nil {
				log.Fatal(err)
			}

			rv := psls[i]
			if err = p.Parse(&rv.Lines); err != nil {
				pslche <- err
				return
			}
			pslch <- rv
		}(i)

		select {
		case err := <-pslche:
			log.Fatal(err)
		case <-pslch:
			// psls = append(psls, prl)
			wg.Done()
		}
		wg.Wait()
	}

	return psls
}

func maybeDate(p string) (time.Time, error) {
	base := filepath.Base(p)
	return time.Parse("060102", base)
}
