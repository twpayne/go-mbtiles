package mbtiles

// FIXME use views

import (
	"database/sql"
)

// A Writer writes a tileset.
type Writer struct {
	Reader
	hasTiles       bool
	tileInsertStmt *sql.Stmt
}

// NewWriter returns a new Writer.
func NewWriter(dsn string) (*Writer, error) {
	r, err := NewReader(dsn)
	if err != nil {
		return nil, err
	}
	return &Writer{Reader: *r}, nil
}

// Close releases all resources with w.
func (w *Writer) Close() error {
	var err error
	if w.tileInsertStmt != nil {
		if err2 := w.tileInsertStmt.Close(); err2 != nil {
			err = err2
		}
	}
	if err2 := w.Reader.Close(); err2 != nil {
		err = err2
	}
	return err
}

// CreateTiles creates the tiles view if it does not already exist.
func (w *Writer) CreateTiles() error {
	if w.hasTiles {
		return nil
	}
	if _, err := w.db.Exec(`
		BEGIN TRANSACTION;
		CREATE TABLE IF NOT EXISTS tiles (
			zoom_level INT NOT NULL,
			tile_column INT NOT NULL,
			tile_row INT NOT NULL,
			tile_data BLOB NOT NULL
		);
		CREATE UNIQUE INDEX IF NOT EXISTS tiles_index ON tiles (zoom_level, tile_column, tile_row);
		COMMIT;
	`); err != nil {
		return err
	}
	w.hasTiles = true
	return nil
}

// InsertTile inserts a tile at (z, x, y).
func (w *Writer) InsertTile(z, x, y int, tileData []byte) error {
	if err := w.CreateTiles(); err != nil {
		return err
	}
	if w.tileInsertStmt == nil {
		var err error
		w.tileInsertStmt, err = w.db.Prepare("INSERT OR REPLACE INTO tiles (zoom_level, tile_column, tile_row, tile_data) VALUES (?, ?, ?, ?);")
		if err != nil {
			return err
		}
	}
	_, err := w.tileInsertStmt.Exec(z, x, y, tileData)
	return err
}

// SelectTile returns the tile at (z, x, y).
func (w *Writer) SelectTile(z, x, y int) ([]byte, error) {
	if err := w.CreateTiles(); err != nil {
		return nil, err
	}
	return w.Reader.SelectTile(z, x, y)
}
