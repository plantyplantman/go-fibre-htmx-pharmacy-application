package main

import (
	"log"
	"time"

	"github.com/plantyplantman/bcapi/pkg/bigc"
)

func main() {
	date := time.Now().Format("060102")

	// var ZihanFilesPath = `C:\Users\admin\Develin Management Dropbox\Zihan\files\`
	// path := filepath.Join(ZihanFilesPath, date+`\`, date+`__web__pf.tsv`)

	path := date + `__web__pf.tsv`
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
