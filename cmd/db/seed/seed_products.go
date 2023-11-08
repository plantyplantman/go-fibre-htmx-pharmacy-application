package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/plantyplantman/bcapi/api/presenter"
	"github.com/plantyplantman/bcapi/pkg/entities"
	"github.com/plantyplantman/bcapi/pkg/env"
	"github.com/plantyplantman/bcapi/pkg/parser"
	"github.com/plantyplantman/bcapi/pkg/report"
)

func main() {
	var connString = env.TEST_NEON
	if connString == "" {
		log.Fatalln("TEST_NEON_CONNECTION_STRING not set")
	}

	var date = time.Now().Format("060102")
	var ZihanFilesPath = `C:\Users\admin\Develin Management Dropbox\Zihan\files\`
	var inPath = filepath.Join(ZihanFilesPath, `in\`+date)
	var outPath = filepath.Join(ZihanFilesPath, `out\`+date)
	outFilePath := filepath.Join(outPath, date+`__ms-web__pf.tsv`)

	var psls report.ProductStockLists = mustParseProductStockList(inPath)
	prl := mustParseProductRetailList(filepath.Join(inPath, date+`__petrie__prlwgp.TXT`))
	pf := mustParseProductFile(filepath.Join(inPath, date+`__web__pf.tsv`))

	var products = NewProducts(psls, prl, pf)

	// DB, err := gorm.Open(postgres.Open(connString), &gorm.Config{})
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// if err = db.Migrate(DB); err != nil {
	// 	log.Fatalln(err)
	// }
	// if err = db.Seed(DB, products); err != nil {
	// 	log.Fatalln(err)
	// }

	err := export(products, outFilePath)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("exported to", outFilePath)
}

func export(products []*entities.Product, path string) error {
	var pps presenter.Products
	for _, p := range products {
		var pp = &presenter.Product{}
		pp.FromEntity(p)
		pps = append(pps, pp)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0770); err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	gocsv.SetCSVWriter(func(out io.Writer) *gocsv.SafeCSVWriter {
		writer := csv.NewWriter(out)
		writer.Comma = '\t'
		return gocsv.NewSafeCSVWriter(writer)
	})

	err = gocsv.MarshalFile(&pps, f)
	if err != nil {
		log.Fatalln(err)
	}

	return nil
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
	if p, err = parser.NewCsvParser(f); err != nil {
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

	p, err := parser.NewCsvParser(f)
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

func mustParseProductStockList(path string) []*report.ProductStockList {
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
	if len(fs) != 4 {
		log.Fatal("wrong number of sts files, expected 4, got", len(fs))
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
			if p, err = parser.NewCsvParser(file); err != nil {
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
