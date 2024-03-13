package main

import (
	"io"
	"log"
	"os/exec"
)

func main() {
	path := `C:/Users/admin/Develin Management Dropbox/Zihan/files/in/240313/240313__web__pf.tsv`
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
			log.Println(msg)
			if msg == "ok\n" {
				break Loop
			}
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
