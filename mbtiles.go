// Package mbtiles implements an HTTP handler for map tiles in MBTiles format.
// See https://github.com/mapbox/mbtiles-spec.
package mbtiles

import (
	"database/sql"
	"net/http"
	"regexp"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

var zxyRegexp = regexp.MustCompile(`\A([0-9]+)/([0-9]+)/([0-9]+)\z`)

// A TileServer is an abstract tile server.
type TileServer struct {
	db   *sql.DB
	stmt *sql.Stmt
}

// NewTileServer returns a new TileServer that serves tiles from dsn.
func NewTileServer(dsn string) (*TileServer, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	stmt, err := db.Prepare("SELECT tile_data FROM tiles WHERE zoom_level = ? AND tile_column = ? AND tile_row = ?;")
	if err != nil {
		db.Close()
		return nil, err
	}
	return &TileServer{db, stmt}, nil
}

// Close releases all resources associated with t.
func (t *TileServer) Close() error {
	for _, err := range []error{
		t.stmt.Close(),
		t.db.Close(),
	} {
		if err != nil {
			return err
		}
	}
	return nil
}

// ServeHTTP implements http.Handler.
func (t *TileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m := zxyRegexp.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return
	}
	z, _ := strconv.Atoi(m[1])
	x, _ := strconv.Atoi(m[2])
	y, _ := strconv.Atoi(m[3])
	var tileData []byte
	if err := t.stmt.QueryRow(z, x, 1<<uint(z)-y-1).Scan(&tileData); err != nil {
		http.NotFound(w, r)
		return
	}
	w.Write(tileData)
}
