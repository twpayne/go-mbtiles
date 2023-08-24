package mbtiles_test

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/twpayne/go-mbtiles"
)

func newWriter(t *testing.T, dsn string) *mbtiles.Writer {
	w, err := mbtiles.NewWriter(dsn)
	if err != nil {
		t.Fatalf("NewWriter(%q) == %v, %v, want _, <nil>", dsn, w, err)
	}
	return w
}

func TestWriter_CreateTiles(t *testing.T) {
	w := newWriter(t, ":memory:")
	if err := w.CreateTiles(); err != nil {
		t.Errorf("w.CreateTiles() == %v, want <nil>", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("w.Close() == %v, want <nil>", err)
	}
}

func TestWriter_CreateMetadata(t *testing.T) {
	w := newWriter(t, ":memory:")
	if err := w.CreateMetadata(); err != nil {
		t.Errorf("w.CreateMetadata() == %v, want <nil>", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("w.Close() == %v, want <nil>", err)
	}
}

func TestWriter_InsertTile(t *testing.T) {
	w := newWriter(t, ":memory:")
	z, x, y, tileData := 0, 0, 0, []byte{0}
	if err := w.InsertTile(z, x, y, tileData); err != nil {
		t.Errorf("w.InsertTile(%d, %d, %d, %v) == %v, want <nil>", z, x, y, tileData, err)
	}
	if gotTileData, err := w.SelectTile(z, x, y); err != nil || !reflect.DeepEqual(gotTileData, tileData) {
		t.Errorf("w.SelectTile(%d, %d, %d) == %v, %v, want %v, <nil>", z, x, y, gotTileData, err, tileData)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("w.Close() == %v, want <nil>", err)
	}
}

func TestWriter_InsertDeleteMetadata(t *testing.T) {
	w := newWriter(t, ":memory:")
	name := "name"
	value := "foobarbaz"
	if err := w.InsertMetadata(name, value); err != nil {
		t.Errorf("w.InsertMetadata(%s, %s) == %v, want <nil>", name, value, err)
	}
	if dbValue, err := w.SelectMetadata(name); err != nil || value != dbValue {
		t.Errorf("w.SelectMetadata(%s) == %s, %v, want %s, <nil>", name, dbValue, err, value)
	}
	if err := w.DeleteMetadata(); err != nil {
		t.Errorf("w.DeleteMetadata() == %v, want <nil>", err)
	}
	if dbValue, err := w.SelectMetadata(name); err == nil || !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("w.SelectMetadata(%s) == %s, %v, want %s, %v", name, dbValue, err, "", sql.ErrNoRows)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("w.Close() == %v, want <nil>", err)
	}
}

func TestWriter_ReplaceTile(t *testing.T) {
	w := newWriter(t, ":memory:")
	z, x, y, tileData1 := 0, 0, 0, []byte{0}
	if err := w.InsertTile(z, x, y, tileData1); err != nil {
		t.Errorf("w.InsertTile(%d, %d, %d, %v) == %v, want <nil>", z, x, y, tileData1, err)
	}
	if gotTileData, err := w.SelectTile(z, x, y); err != nil || !reflect.DeepEqual(gotTileData, tileData1) {
		t.Errorf("w.SelectTile(%d, %d, %d) == %v, %v, want %v, <nil>", z, x, y, gotTileData, err, tileData1)
	}
	tileData2 := []byte{1}
	if err := w.InsertTile(z, x, y, tileData2); err != nil {
		t.Errorf("w.InsertTile(%d, %d, %d, %v) == %v, want <nil>", z, x, y, tileData2, err)
	}
	if gotTileData, err := w.SelectTile(z, x, y); err != nil || !reflect.DeepEqual(gotTileData, tileData2) {
		t.Errorf("w.SelectTile(%d, %d, %d) == %v, %v, want %v, <nil>", z, x, y, gotTileData, err, tileData2)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("w.Close() == %v, want <nil>", err)
	}
}

func TestWriter_BulkInsertTile(t *testing.T) {
	w := newWriter(t, ":memory:")
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
	if err := w.BulkInsertTile(tiles); err != nil {
		t.Errorf("w.InsertTile(%v) == %v, want <nil>", tiles, err)
	}
	for _, tile := range tiles {
		if gotTileData, err := w.SelectTile(tile.Z, tile.X, tile.Y); err != nil || !reflect.DeepEqual(gotTileData, tile.Data) {
			t.Errorf("w.SelectTile(%d, %d, %d) == %v, %v, want %v, <nil>", tile.Z, tile.X, tile.Y, gotTileData, err, tile)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("w.Close() == %v, want <nil>", err)
	}
}
