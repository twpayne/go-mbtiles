package mbtiles_test

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/twpayne/go-mbtiles"
)

func hexDecodeSHA256Sum(t *testing.T, s string) (sha256sum [sha256.Size]byte) {
	slice, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("hexDecode(..., %q) == %v, %v, want _, <nil>", s, slice, err)
	}
	copy(sha256sum[:], slice)
	return
}

func newReader(t *testing.T, dsn string) *mbtiles.Reader {
	r, err := mbtiles.NewReader(dsn)
	if err != nil {
		t.Fatalf("mbtiles.NewReader(%q) == %v, %v, want _, <nil>", dsn, r, err)
	}
	return r
}

func TestReader_SelectTile(t *testing.T) {
	mbtrCache := make(map[string]*mbtiles.Reader)
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
		mbtr, ok := mbtrCache[tc.dsn]
		if !ok {
			mbtr = newReader(t, tc.dsn)
			mbtrCache[tc.dsn] = mbtr
		}
		tileData, err := mbtr.SelectTile(tc.z, tc.x, tc.y)
		if err != nil {
			t.Errorf("mbtiles.NewReader(%q).SelectTile(%d, %d, %d) == %v, %v, want _, <nil>", tc.dsn, tc.z, tc.x, tc.y, tileData, err)
			continue
		}
		if sha256sum := sha256.Sum256(tileData); sha256sum != tc.sha256sum {
			t.Errorf("mbtiles.NewReader(%q).SelectTile(%d, %d, %d) tile data has SHA256 sum %s, want %s", tc.dsn, tc.z, tc.x, tc.y, hex.EncodeToString(sha256sum[:]), hex.EncodeToString(tc.sha256sum[:]))
		}
	}
	for dsn, mbtr := range mbtrCache {
		if err := mbtr.Close(); err != nil {
			t.Errorf("mbtiles.NewReader(%q).Close() == %v, want <nil>", dsn, err)
		}
	}
}

func TestReader_SelectMetadata(t *testing.T) {
	mbtrCache := make(map[string]*mbtiles.Reader)
	for _, tc := range []struct {
		dsn   string
		name  string
		value string
	}{
		{
			dsn:   "testdata/openstreetmap.org.mbtiles",
			name:  "name",
			value: "testdata",
		},
	} {
		mbtr, ok := mbtrCache[tc.dsn]
		if !ok {
			mbtr = newReader(t, tc.dsn)
			mbtrCache[tc.dsn] = mbtr
		}
		value, err := mbtr.SelectMetadata(tc.name)
		if err != nil {
			t.Errorf("mbtiles.NewReader(%q).SelectMetadata(%s) == %v, %v, want _, <nil>", tc.dsn, tc.name, tc.value, err)
			continue
		}
		if value != tc.value {
			t.Errorf("mbtiles.NewReader(%q).SelectMetadata(%s) = %v != %v", tc.dsn, tc.name, tc.value, value)
		}
	}
	for dsn, mbtr := range mbtrCache {
		if err := mbtr.Close(); err != nil {
			t.Errorf("mbtiles.NewReader(%q).Close() == %v, want <nil>", dsn, err)
		}
	}
}

func TestReader_ServeHTTP(t *testing.T) {
	mbtr := newReader(t, "testdata/openstreetmap.org.mbtiles")
	s := httptest.NewServer(http.StripPrefix("/", mbtr))
	defer s.Close()
	url := s.URL + "/0/0/0"
	res, err := http.Get(url)
	if err != nil {
		t.Errorf("http.Get(%q) == %v, %v, want _, <nil>", url, res, err)
	}
	if got, want := res.StatusCode, http.StatusOK; got != want {
		t.Errorf("res.StatusCode() == %d, want %d", got, want)
	}
	got, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("ioutil.ReadAll(_) == %v, %v, want _, <nil>", got, err)
	}
	wantSHA256Sum := hexDecodeSHA256Sum(t, "075c660f81ba41146fda8610216a077b81bf5d8d102dbc893a57b7969e32ee88")
	if gotSHA256Sum := sha256.Sum256(got); gotSHA256Sum != wantSHA256Sum {
		t.Errorf("got SHA256 sum %s, want %v", gotSHA256Sum, wantSHA256Sum)
	}
}
