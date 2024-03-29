package main

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/plantyplantman/bcapi/pkg/bigc"
)

func main() {
	date := time.Now().Format("060102")
	path := filepath.Join(`C:/Users/admin/Develin Management Dropbox/Zihan/files/in/`, date, date+`__web__pf.tsv`)
	if err := os.MkdirAll(filepath.Dir(path), 0770); err != nil {
		log.Fatalln(err)
	}
	c := bigc.MustGetClient()

	var (
		pf  *bigc.ProductFile
		err error
	)
	log.Println("Getting product file...")
	if pf, err = c.GetProductFile(); err != nil {
		log.Fatalln(err)
	}

	if err := pf.Export(path); err != nil {
		log.Fatalln(err)
	}
	log.Println("Exported product file to", path)

	log.Println("Executing Python...")
	cmd := exec.Command("python3", "cmd\\ftp\\main.py", "upload", path)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalln(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalln(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	msgs := make(chan string)
	defer close(msgs)
	errs := make(chan string)
	defer close(errs)
	done := make(chan struct{})

	go func() {
		for {
			resp := make([]byte, 1024)
			n, err := stdout.Read(resp)
			if err != nil {
				if err == io.EOF {
					done <- struct{}{}
					return
				}
				log.Fatal(err)
			}
			resp = resp[:n]
			msgs <- string(resp)
		}
	}()

	go func() {
		for {
			resp := make([]byte, 1024)
			n, err := stderr.Read(resp)
			if err != nil {
				if err == io.EOF {
					done <- struct{}{}
					return
				}
				log.Fatal(err)
			}
			resp = resp[:n]
			errs <- string(resp)
		}
	}()

Loop:
	for {
		select {
		case msg := <-msgs:
			if msg == "ok\n" {
				log.Println("Python script finished successfully.")
				log.Println("File uploaded to FTP server.")
				break Loop
			}
			log.Println(msg)
		case err := <-errs:
			log.Println(err)
			break Loop
		case <-done:
			break Loop
		}
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}
