package mbtiles

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func hexDecodeSHA256Sum(t *testing.T, s string) (sha256sum [sha256.Size]byte) {
	slice, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("hexDecode(..., %q) == %v, %v, want _, <nil>", s, slice, err)
	}
	copy(sha256sum[:], slice)
	return
}

func TestSelectTile(t *testing.T) {
	mbtCache := make(map[string]*T)
	for _, tc := range []struct {
		dsn       string
		z, x, y   int
		sha256sum [sha256.Size]byte
	}{
		{
			dsn:       "testdata/openstreetmap.org.mbtiles",
			z:         0,
			x:         0,
			y:         0,
			sha256sum: hexDecodeSHA256Sum(t, "075c660f81ba41146fda8610216a077b81bf5d8d102dbc893a57b7969e32ee88"),
		},
	} {
		mbt, ok := mbtCache[tc.dsn]
		if !ok {
			var err error
			mbt, err = New(tc.dsn)
			if err != nil {
				t.Errorf("New(%q) == %v, %v, want _, <nil>", tc.dsn, mbt, err)
				continue
			}
			mbtCache[tc.dsn] = mbt
		}
		tileData, err := mbt.SelectTile(tc.z, tc.x, tc.y)
		if err != nil {
			t.Errorf("New(%q).SelectTile(%d, %d, %d) == %v, %v, want _, <nil>", tc.dsn, tc.z, tc.x, tc.y)
			continue
		}
		if sha256sum := sha256.Sum256(tileData); sha256sum != tc.sha256sum {
			t.Errorf("New(%q).SelectTile(%d, %d, %d) tile data has SHA256 sum %s, want %s", tc.dsn, tc.z, tc.x, tc.y, hex.EncodeToString(sha256sum[:]), hex.EncodeToString(tc.sha256sum[:]))
		}
	}
	for dsn, mbt := range mbtCache {
		if err := mbt.Close(); err != nil {
			t.Errorf("New(%q).Close() == %v, want <nil>", dsn, err)
		}
	}
}
