// Package mbtiles handles the MBTiles tileset format.
// See https://github.com/mapbox/mbtiles-spec.
package mbtiles

import (
	"database/sql"
	"net/http"
	"regexp"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

var (
	zxyRegexp = regexp.MustCompile(`\A([0-9]+)/([0-9]+)/([0-9]+)\z`)
)

// A T is an MBTiles tileset.
type T struct {
	db             *sql.DB
	tileSelectStmt *sql.Stmt
}

// New returns a new T.
func New(dsn string) (*T, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	return &T{db: db}, nil
}

// Close releases all resources associated with t.
func (t *T) Close() error {
	var err error
	if t.tileSelectStmt != nil {
		if err2 := t.tileSelectStmt.Close(); err2 != nil {
			err = err2
		}
	}
	if t.db != nil {
		if err2 := t.db.Close(); err2 != nil {
			err = err2
		}
	}
	return err
}

// SelectTile returns the tile at (z, x, y).
func (t *T) SelectTile(z, x, y int) ([]byte, error) {
	if t.tileSelectStmt == nil {
		var err error
		t.tileSelectStmt, err = t.db.Prepare("SELECT tile_data FROM tiles WHERE zoom_level = ? AND tile_column = ? AND tile_row = ?;")
		if err != nil {
			return nil, err
		}
	}
	var tileData []byte
	err := t.tileSelectStmt.QueryRow(z, x, 1<<uint(z)-y-1).Scan(&tileData)
	return tileData, err
}

// ServeHTTP implements http.Handler.
func (t *T) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m := zxyRegexp.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return
	}
	z, _ := strconv.Atoi(m[1])
	x, _ := strconv.Atoi(m[2])
	y, _ := strconv.Atoi(m[3])
	tileData, err := t.SelectTile(z, x, y)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Write(tileData)
}
