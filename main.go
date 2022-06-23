package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/tftp-server/tftp"
)

var (
	address = flag.String("a", "127.0.0.1:69", "listen address")
	payload = flag.String("p", "payload.jpeg", "file to serve to clients")
)

func main() {
	flag.Parse()

	if _, err := os.Stat(*payload); errors.Is(err, os.ErrNotExist) {
		log.Fatalf("File '%s' does not exist", *payload)
	}

	p, err := ioutil.ReadFile("payload.jpeg")
	if err != nil {
		log.Fatal(err)
	}

	s := tftp.Server{Payload: p}
	log.Fatal(s.ListenAndServer(*address))
}
