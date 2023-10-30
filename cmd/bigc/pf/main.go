package main

import (
	"log"
	"time"

	"github.com/plantyplantman/bcapi/pkg/bigc"
)

func main() {
	date := time.Now().Format("060102")
	path := "/mnt/c/Users/admin/Develin Management Dropbox/Zihan/files/in/" + date + "/" + date + "__web__pf.tsv"
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
}
