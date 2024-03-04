package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/plantyplantman/bcapi/pkg/bigc"
)

func main() {
	date := time.Now().Format("060102")
	path := filepath.Join(`C:\Users\admin\Develin Management Dropbox\Zihan\files\in\`, date, date+`__web__pf.tsv`)
	if err := os.MkdirAll(filepath.Dir(path), 0770); err != nil {
		log.Fatalln(err)
	}
	c := bigc.MustGetClient()

	var (
		pf  *bigc.ProductFile
		err error
	)
	if pf, err = c.GetProductFile(); err != nil {
		log.Fatalln(err)
	}

	if err := pf.Export(path); err != nil {
		log.Fatalln(err)
	}

	log.Println("Exported product file to", path)
}
