package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/vharish836/middleman/config"
	"github.com/vharish836/middleman/mcservice"
)

func main() {
	f := flag.String("config", "middleman.toml", "Config file name")
	cfg, err := config.GetConfig(*f)
	if err != nil {
		log.Fatalf("could not load config: %s", err)
	}
	s,err := mcservice.NewService(cfg)
	if err != nil {
		log.Fatalf("could not initialize service: %s", err)
	}
	http.Handle("/", s.GetHandler())
	err = http.ListenAndServe("localhost:8383", nil)
	if err != nil {
		log.Fatalf("could not listen: %s", err)
	}
}
