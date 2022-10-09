package main

import (
	"flag"
	"log"

	"github.com/xegea/webhook_client/pkg/config"
	"github.com/xegea/webhook_client/pkg/server"
)

func main() {
	env := flag.String("env", ".env", ".env path")
	url := flag.String("url", "http:/localhost:8080", "url to connect. ie. http:/localhost:8080")
	flag.Parse()

	cfg, err := config.LoadConfig(env)
	if err != nil {
		log.Fatalf("unable to load config: %+v", err)
	}

	svr := server.NewServer(
		*cfg,
	)

	log.Fatal(svr.Start(*url))
}
