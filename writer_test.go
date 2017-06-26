package mbtiles_test

import (
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
