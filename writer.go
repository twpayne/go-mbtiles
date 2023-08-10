package mbtiles

// FIXME use views

import (
	"context"
	"database/sql"
)

// A Writer writes a tileset.
type Writer struct {
	Reader
	hasTiles       bool
	hasMetadata    bool
	tileInsertStmt *sql.Stmt
}

type Optimizations struct {
	// Synchronous turns ON or OFF the statement PRAGMA synchronous = OFF
	SynchronousOff bool
	// JournalModeMemory turns ON or OFF the statement PRAGMA journal_mode = MEMORY
	JournalModeMemory bool
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

// SetOptimizations can be used to turn on or off Optimization options
func (w *Writer) SetOptimizations(opts Optimizations) error {
	if opts.SynchronousOff {
		if _, err := w.db.Exec("PRAGMA synchronous = OFF"); err != nil {
			return err
		}
	}
	if opts.JournalModeMemory {
		if _, err := w.db.Exec("PRAGMA journal_mode = MEMORY"); err != nil {
			return err
		}
	}
	return nil
}

// CreateTiles creates the tiles table if it does not already exist.
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
			tile_data BLOB NOT NULL,
			PRIMARY KEY (zoom_level, tile_column, tile_row)
		);
		CREATE UNIQUE INDEX IF NOT EXISTS tiles_index ON tiles (zoom_level, tile_column, tile_row);
		COMMIT;
	`); err != nil {
		return err
	}
	w.hasTiles = true
	return nil
}

// CreateTileIndex generates the standard index on the tiles table
func (w *Writer) CreateTileIndex() error {
	if _, err := w.db.Exec(`
		BEGIN TRANSACTION;
		CREATE UNIQUE INDEX IF NOT EXISTS tiles_index ON tiles (zoom_level, tile_column, tile_row);
		COMMIT;
	`); err != nil {
		return err
	}
	return nil
}

// DeleteTileIndex removes the tile index, useful for speeding up bulk inserts
func (w *Writer) DeleteTileIndex() error {
	if _, err := w.db.Exec(`
		BEGIN TRANSACTION;
		DROP INDEX IF EXISTS tiles_index;
		COMMIT;
	`); err != nil {
		return err
	}
	return nil
}

// CreateMetadata creates the metadata table if it does not already exist.
func (w *Writer) CreateMetadata() error {
	if w.hasMetadata {
		return nil
	}
	if _, err := w.db.Exec(`
		BEGIN TRANSACTION;
		CREATE TABLE IF NOT EXISTS metadata (name TEXT, value TEXT, PRIMARY KEY (name));
		COMMIT;
	`); err != nil {
		return err
	}
	w.hasMetadata = true
	return nil
}

// InsertMetadata inserts a name, value row to the metadata store
func (w *Writer) InsertMetadata(name string, value string) error {
	if err := w.CreateMetadata(); err != nil {
		return err
	}
	_, err := w.db.Exec("INSERT OR REPLACE INTO metadata (name, value) VALUES (?, ?);", name, value)
	return err
}

// DeleteMetadata removes the metadata table, useful for resetting the metadata in the mbtiles file
func (w *Writer) DeleteMetadata() error {
	if _, err := w.db.Exec(`
		BEGIN TRANSACTION;
		DELETE FROM metadata;
		COMMIT;
	`); err != nil {
		return err
	}
	w.hasMetadata = false
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

// BulkInsertTile inserts multiple tiles at the coordinates provided (z, x, y).
// This can be faster because it reduces the number of transactions.
// By default, sqlite wraps each insert in a transaction.
func (w *Writer) BulkInsertTile(z, x, y []int, tileData [][]byte) error {
	if err := w.CreateTiles(); err != nil {
		return err
	}
	tx, err := w.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	if w.tileInsertStmt == nil {
		var err error
		w.tileInsertStmt, err = w.db.Prepare("INSERT OR REPLACE INTO tiles (zoom_level, tile_column, tile_row, tile_data) VALUES (?, ?, ?, ?);")
		if err != nil {
			return err
		}
	}
	stmt := tx.Stmt(w.tileInsertStmt)
	for i, _ := range z {
		if _, err := stmt.Exec(z[i], x[i], y[i], tileData[i]); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// SelectTile returns the tile at (z, x, y).
func (w *Writer) SelectTile(z, x, y int) ([]byte, error) {
	if err := w.CreateTiles(); err != nil {
		return nil, err
	}
	return w.Reader.SelectTile(z, x, y)
}
