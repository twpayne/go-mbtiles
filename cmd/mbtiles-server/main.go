package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "modernc.org/sqlite" // Register sqlite database driver.

	"github.com/twpayne/go-mbtiles"
)

var (
	indexHTML = template.Must(template.New("index.html").Parse(`<html>
	<head>
		<title>mbtiles-server</title>
		<link rel="stylesheet" href="https://openlayers.org/en/v4.2.0/css/ol.css" type="text/css">
		<script src="https://openlayers.org/en/v4.2.0/build/ol.js"></script>
	</head>
	<body>
		<div id="map" class="map"></div>
		<script>
			var map = new ol.Map({
				layers: [
					new ol.layer.Tile({
						source: new ol.source.XYZ({
							url: {{.TilePrefix}} + "{z}/{x}/{y}"
						})
					})
				],
				target: 'map',
				view: new ol.View({
					center: [0, 0],
					zoom: 0
				})
			});
		</script>
	</body>
</html>
`))
)

var (
	addr = flag.String("addr", "localhost:8080", "addr")
	dsn  = flag.String("dsn", "", "dsn")
)

type mapServer struct {
	TilePrefix string
}

func (ms *mapServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := indexHTML.Execute(w, ms); err != nil {
		log.Print(err)
	}
}

func run() error {
	flag.Parse()
	mbtr, err := mbtiles.NewReader("sqlite", *dsn)
	if err != nil {
		return err
	}
	defer func() {
		if err := mbtr.Close(); err != nil {
			log.Print(err)
		}
	}()
	r := mux.NewRouter()
	tilePrefix := "/" + filepath.Base(*dsn) + "/"
	ms := &mapServer{
		TilePrefix: tilePrefix,
	}
	r.PathPrefix(tilePrefix).Handler(http.StripPrefix(tilePrefix, mbtr))
	r.PathPrefix("/").Handler(http.StripPrefix("/", ms))
	http.Handle("/", handlers.LoggingHandler(os.Stdout, r))
	return http.ListenAndServe(*addr, nil)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
