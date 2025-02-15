package mbtiles

// FIXME use views

import (
	"context"
	"database/sql"
)

// A Writer writes a tileset. It should be created with NewWriter().
type Writer struct {
	Reader
	hasTiles       bool
	hasMetadata    bool
	tileInsertStmt *sql.Stmt
}

// TileData represents a single tile for writing to the mbtiles file.
type TileData struct {
	X    int    // X coordinate in ZXY format
	Y    int    // Y coordinate in ZXY format
	Z    int    // Z coordinate in ZXY format
	Data []byte // Tile data, must be gzip encoded for mbtiles
}

// Optimizations are options that can be turned on/off for writing mbtiles
// files, however they do have side-effects so best to research before using
// them.
type Optimizations struct {
	// Synchronous turns ON or OFF the statement PRAGMA synchronous = OFF.
	SynchronousOff bool
	// JournalModeMemory turns ON or OFF the statement PRAGMA journal_mode = MEMORY.
	JournalModeMemory bool
}

// NewWriter returns a new Writer.
func NewWriter(driverName, dsn string) (*Writer, error) {
	r, err := NewReader(driverName, dsn)
	if err != nil {
		return nil, err
	}
	return &Writer{Reader: *r}, nil
}

// NewWriterWithDB returns a new Writer initialized with a sql.Database.
// This is useful for instantiating alternative implementations of sqlite.
func NewWriterWithDB(db *sql.DB) (*Writer, error) {
	r, err := NewReaderWithDB(db)
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

// SetOptimizations can be used to turn on or off Optimization options.
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

// CreateTileIndex generates the standard index on the tiles table.
func (w *Writer) CreateTileIndex() error {
	if _, err := w.db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS tiles_index ON tiles (zoom_level, tile_column, tile_row);
	`); err != nil {
		return err
	}
	return nil
}

// DeleteTileIndex removes the tile index, useful for speeding up bulk inserts.
func (w *Writer) DeleteTileIndex() error {
	if _, err := w.db.Exec(`
		DROP INDEX IF EXISTS tiles_index;
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
		CREATE TABLE IF NOT EXISTS metadata (name TEXT, value TEXT, PRIMARY KEY (name));
	`); err != nil {
		return err
	}
	w.hasMetadata = true
	return nil
}

// InsertMetadata inserts a name, value row to the metadata store.
func (w *Writer) InsertMetadata(name, value string) error {
	if err := w.CreateMetadata(); err != nil {
		return err
	}
	_, err := w.db.Exec(`
		INSERT OR REPLACE INTO metadata (name, value) VALUES (?, ?);
	`, name, value)
	return err
}

// DeleteMetadata removes the metadata table, useful for resetting the metadata
// in the mbtiles file.
func (w *Writer) DeleteMetadata() error {
	if _, err := w.db.Exec(`
		DROP TABLE IF EXISTS metadata;
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
		w.tileInsertStmt, err = w.db.Prepare(`
			INSERT OR REPLACE INTO tiles (zoom_level, tile_column, tile_row, tile_data) VALUES (?, ?, ?, ?);
		`)
		if err != nil {
			return err
		}
	}
	_, err := w.tileInsertStmt.Exec(z, x, 1<<uint(z)-y-1, tileData)
	return err
}

// BulkInsertTile inserts multiple tiles at the coordinates provided (z, x, y).
// This can be faster because it reduces the number of transactions.
// By default, sqlite wraps each insert in a transaction.
func (w *Writer) BulkInsertTile(data []TileData) error {
	if err := w.CreateTiles(); err != nil {
		return err
	}
	tx, err := w.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	if w.tileInsertStmt == nil {
		var err error
		w.tileInsertStmt, err = w.db.Prepare(`
			INSERT OR REPLACE INTO tiles (zoom_level, tile_column, tile_row, tile_data) VALUES (?, ?, ?, ?);
		`)
		if err != nil {
			return err
		}
	}
	stmt := tx.Stmt(w.tileInsertStmt)
	defer stmt.Close()
	for _, d := range data {
		if _, err := stmt.Exec(d.Z, d.X, 1<<uint(d.Z)-d.Y-1, d.Data); err != nil {
			_ = tx.Rollback()
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

// SelectMetadata returns the metadata value for name.
func (w *Writer) SelectMetadata(name string) (string, error) {
	if err := w.CreateMetadata(); err != nil {
		return "", err
	}
	return w.Reader.SelectMetadata(name)
}
