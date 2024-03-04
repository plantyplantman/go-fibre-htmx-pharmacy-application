package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/plantyplantman/bcapi/api/presenter"
	"github.com/plantyplantman/bcapi/pkg/bigc"
	"github.com/plantyplantman/bcapi/pkg/product"
)

func main() {
	date := time.Now().Format("060102")

	infile := `C:\Users\admin\Develin Management Dropbox\Zihan\files\in\240228\item_issues_406002091_2024-02-28T04_22_24.245Z_.csv`
	outfile := filepath.Join(`C:\Users\admin\Develin Management Dropbox\Zihan\files\out\`, date, date+`__web__fixed-secondary-images.tsv`)
	f, err := os.Open(infile)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.Comma = '\t'
	r.LazyQuotes = true

	var (
		skus  = make([]string, 0)
		count = 0
	)
	// skip header
	_, err = r.Read()
	if err != nil {
		log.Fatalln(err)
	}
	for {
		line, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln(err)
		}
		sku, err := bigc.CleanBigCommerceSku(line[0])
		if err != nil {
			log.Println(err)
			continue
		}
		skus = append(skus, sku)
		count++
	}
	if count != len(skus) {
		log.Fatalln("skus count mismatch")
	}

	service, err := product.NewDefaultService()
	if err != nil {
		log.Fatalln(err)
	}

	ps, err := service.FetchProducts(product.WithSkus(skus...))
	if err != nil {
		log.Fatalln(err)
	}

	c := bigc.MustGetClient()

	f, err = os.Create(outfile)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Comma = '\t'

	run(ps, c, w, func(w *csv.Writer, v *bigc.Variant) error {
		return w.Write([]string{v.Sku, v.ImageURL})
	}, func(w *csv.Writer, p *bigc.Product) error {
		return w.Write([]string{p.Sku, p.Variants[0].ImageURL})
	})

	// run(ps, c, w, func(w *csv.Writer, v *bigc.Variant) error {
	// 	// return w.Write([]string{v.Sku, v.ImageURL})
	// 	return nil
	// }, func(w *csv.Writer, p *bigc.Product) error {
	// 	urls := lo.FilterMap(p.Images, func(img bigc.Image, i int) (string, bool) {
	// 		return img.ImageURL, i != 0
	// 	})
	// 	var out = []string{p.Sku}
	// 	out = append(out, urls...)
	// 	return w.Write(out)
	// })
}

func run(ps []*presenter.Product, c *bigc.Client, w *csv.Writer, wFuncVariant func(w *csv.Writer, v *bigc.Variant) error, wFuncProduct func(w *csv.Writer, p *bigc.Product) error) {
	for _, p := range ps {
		ids := strings.Split(p.BCID, "/")
		if p.IsVariant && len(ids) == 2 {
			vid, err := strconv.Atoi(ids[1])
			if err != nil {
				log.Println(err)
				continue
			}
			pid, err := strconv.Atoi(ids[0])
			if err != nil {
				log.Println(err)
				continue
			}
			vp, err := c.GetVariantById(vid, pid, map[string]string{})
			if err != nil {
				log.Println(err)
				continue
			}
			err = wFuncVariant(w, vp)
			// fmt.Println(vp.Sku, "\t", vp.ImageURL)
			// err = w.Write([]string{vp.Sku, vp.ImageURL})
			if err != nil {
				log.Fatalln(err)
			}
		} else if !p.IsVariant && len(ids) == 1 && ids[0] != "" {
			pid, err := strconv.Atoi(ids[0])
			if err != nil {
				log.Println(err)
				continue
			}
			pp, err := c.GetProductById(pid)
			if err != nil {
				log.Println(err)
				continue
			}
			if len(pp.Variants) == 0 {
				continue
			}
			// fmt.Println(pp.Sku, "\t", pp.Variants[0].ImageURL)
			// err = w.Write([]string{pp.Sku, pp.Variants[0].ImageURL})
			err = wFuncProduct(w, pp)
			if err != nil {
				log.Fatalln(err)
			}
		} else {
			log.Printf("skipping %s with sku %s due to invalid BCID: %s\n", p.ProdName, p.Sku, p.BCID)
		}
	}
}
