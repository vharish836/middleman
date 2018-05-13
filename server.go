package main

import (
	"github.com/vharish836/middleman/boxer"
	"github.com/vharish836/middleman/handler"
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
	b,err := boxer.NewBoxer(&cfg.Boxer)
	if err != nil {
		log.Fatalf("could not initialize boxer: %s", err)
	}
	s, err := mcservice.NewService(&cfg.MultiChain, b)	
	if err != nil {
		log.Fatalf("could not initialize service: %s", err)
	}
	h := handler.NewHandler(s)
	http.Handle("/", h)
	err = http.ListenAndServe("localhost:8383", nil)
	if err != nil {
		log.Fatalf("could not listen: %s", err)
	}
}
