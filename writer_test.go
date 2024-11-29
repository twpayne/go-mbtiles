package mbtiles_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
	_ "modernc.org/sqlite" // Register sqlite database driver.

	"github.com/twpayne/go-mbtiles"
)

func newWriter(t *testing.T) *mbtiles.Writer {
	dsn := filepath.Join(t.TempDir(), "mbtiles.db")
	w, err := mbtiles.NewWriter("sqlite", dsn)
	assert.NoError(t, err)
	return w
}

func TestWriter_CreateTiles(t *testing.T) {
	w := newWriter(t)
	assert.NoError(t, w.CreateTiles())
	assert.NoError(t, w.Close())
}

func TestWriter_CreateMetadata(t *testing.T) {
	w := newWriter(t)
	assert.NoError(t, w.CreateMetadata())
	assert.NoError(t, w.Close())
}

func TestWriter_InsertTile(t *testing.T) {
	w := newWriter(t)
	z, x, y, tileData := 0, 0, 0, []byte{0}
	assert.NoError(t, w.InsertTile(z, x, y, tileData))
	gotTileData, err := w.SelectTile(z, x, y)
	assert.NoError(t, err)
	assert.Equal(t, tileData, gotTileData)
	assert.NoError(t, w.Close())
}

func TestWriter_InsertDeleteMetadata(t *testing.T) {
	w := newWriter(t)
	name := "name"
	value := "foobarbaz"
	assert.NoError(t, w.InsertMetadata(name, value))
	gotValue, err := w.SelectMetadata(name)
	assert.NoError(t, err)
	assert.Equal(t, value, gotValue)
	assert.NoError(t, w.DeleteMetadata())
	_, err = w.SelectMetadata(name)
	assert.IsError(t, err, sql.ErrNoRows)
	assert.NoError(t, w.Close())
}

func TestWriter_ReplaceTile(t *testing.T) {
	w := newWriter(t)
	z, x, y, tileData1 := 0, 0, 0, []byte{0}
	assert.NoError(t, w.InsertTile(z, x, y, tileData1))
	gotTileData, err := w.SelectTile(z, x, y)
	assert.NoError(t, err)
	assert.Equal(t, tileData1, gotTileData)
	tileData2 := []byte{1}
	assert.NoError(t, w.InsertTile(z, x, y, tileData2))
	gotTileData, err = w.SelectTile(z, x, y)
	assert.NoError(t, err)
	assert.Equal(t, tileData2, gotTileData)
	assert.NoError(t, w.Close())
}

func TestWriter_BulkInsertTile(t *testing.T) {
	w := newWriter(t)
	tiles := []mbtiles.TileData{
		{
			Z:    0,
			X:    0,
			Y:    0,
			Data: []byte{0},
		},
		{
			Z:    6,
			X:    1,
			Y:    5,
			Data: []byte{0, 1, 2, 3, 4, 5},
		},
	}
	assert.NoError(t, w.BulkInsertTile(tiles))
	for _, tile := range tiles {
		gotTileData, err := w.SelectTile(tile.Z, tile.X, tile.Y)
		assert.NoError(t, err)
		assert.Equal(t, tile.Data, gotTileData)
	}
	assert.NoError(t, w.Close())
}
