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
	s := mcservice.NewService(cfg)
	h, herr := s.Initialize()
	if herr != nil {
		log.Fatalf("could not initialize service: %s", herr)
	}
	http.Handle("/", h)
	err = http.ListenAndServe("localhost:8383", nil)
	if err != nil {
		log.Fatalf("could not listen: %s", err)
	}
}
