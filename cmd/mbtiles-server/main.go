package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/handlers"
	"github.com/twpayne/go-mbtiles"
)

var (
	addr = flag.String("addr", "localhost:8080", "addr")
	dsn  = flag.String("dsn", "", "dsn")
)

func run() error {
	flag.Parse()
	mbt, err := mbtiles.New(*dsn)
	if err != nil {
		return err
	}
	defer mbt.Close()
	mbtPrefix := "/" + filepath.Base(*dsn) + "/"
	http.Handle(mbtPrefix, handlers.LoggingHandler(os.Stdout, http.StripPrefix(mbtPrefix, mbt)))
	return http.ListenAndServe(*addr, nil)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
