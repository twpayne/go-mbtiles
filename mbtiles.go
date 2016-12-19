package mbtiles

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"regexp"
	"strconv"
)

var zxyRegexp = regexp.MustCompile(`\A([0-9]+)/([0-9]+)/([0-9]+)\z`)

type tileServer struct {
	db   *sql.DB
	stmt *sql.Stmt
}

func NewTileServer(dsn string) (*tileServer, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	stmt, err := db.Prepare("SELECT tile_data FROM tiles WHERE zoom_level = ? AND tile_column = ? AND tile_row = ?;")
	if err != nil {
		db.Close()
		return nil, err
	}
	return &tileServer{db, stmt}, nil
}

func (t *tileServer) Close() {
	t.stmt.Close()
	t.db.Close()
}

func (t *tileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
