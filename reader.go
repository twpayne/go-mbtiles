package mbtiles

import (
	"database/sql"
	"net/http"
	"regexp"
	"strconv"
)

var (
	zxyRegexp = regexp.MustCompile(`\A([0-9]+)/([0-9]+)/([0-9]+)\z`)
)

// A Reader reads a tileset.
type Reader struct {
	db                 *sql.DB
	tileSelectStmt     *sql.Stmt
	metadataSelectStmt *sql.Stmt
}

// NewReader returns a new Reader.
func NewReader(dsn string) (*Reader, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	return NewReaderWithDB(db)
}

// NewReaderWithDB returns a new Reader initialized with a sql.Database.
// This is useful for instantiating alternative implementations of sqlite.
func NewReaderWithDB(db *sql.DB) (*Reader, error) {
	return &Reader{db: db}, nil
}

// Close releases all resources associated with r.
func (r *Reader) Close() error {
	var err error
	if r.tileSelectStmt != nil {
		if err2 := r.tileSelectStmt.Close(); err2 != nil {
			err = err2
		}
	}
	if r.db != nil {
		if err2 := r.db.Close(); err2 != nil {
			err = err2
		}
	}
	return err
}

// SelectTile returns the tile at (z, x, y).
func (r *Reader) SelectTile(z, x, y int) ([]byte, error) {
	if r.tileSelectStmt == nil {
		var err error
		r.tileSelectStmt, err = r.db.Prepare("SELECT tile_data FROM tiles WHERE zoom_level = ? AND tile_column = ? AND tile_row = ?;")
		if err != nil {
			return nil, err
		}
	}
	var tileData []byte
	err := r.tileSelectStmt.QueryRow(z, x, 1<<uint(z)-y-1).Scan(&tileData)
	return tileData, err
}

// SelectMetadata returns the metadata value for 'name'
func (r *Reader) SelectMetadata(name string) (string, error) {
	if r.tileSelectStmt == nil {
		var err error
		r.metadataSelectStmt, err = r.db.Prepare("SELECT value FROM metadata WHERE name = ?;")
		if err != nil {
			return "", err
		}
	}
	var value string
	err := r.metadataSelectStmt.QueryRow(name).Scan(&value)
	return value, err
}

// ServeHTTP implements http.Handler.
func (r *Reader) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	m := zxyRegexp.FindStringSubmatch(req.URL.Path)
	if m == nil {
		http.NotFound(w, req)
		return
	}
	z, _ := strconv.Atoi(m[1])
	x, _ := strconv.Atoi(m[2])
	y, _ := strconv.Atoi(m[3])
	tileData, err := r.SelectTile(z, x, y)
	if err != nil {
		http.NotFound(w, req)
		return
	}
	_, _ = w.Write(tileData)
}
