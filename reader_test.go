package mbtiles_test

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alecthomas/assert/v2"
	_ "modernc.org/sqlite" // Register sqlite database driver.

	"github.com/twpayne/go-mbtiles"
)

func hexDecodeSHA256Sum(t *testing.T, s string) (sha256sum [sha256.Size]byte) {
	t.Helper()
	slice, err := hex.DecodeString(s)
	assert.NoError(t, err)
	copy(sha256sum[:], slice)
	return
}

func newReader(t *testing.T, dsn string) *mbtiles.Reader {
	t.Helper()
	r, err := mbtiles.NewReader("sqlite", dsn)
	assert.NoError(t, err)
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
		assert.NoError(t, err)
		assert.Equal(t, tc.sha256sum, sha256.Sum256(tileData))
	}
	for _, mbtr := range mbtrCache {
		assert.NoError(t, mbtr.Close())
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
		assert.NoError(t, err)
		assert.Equal(t, tc.value, value)
	}
	for _, mbtr := range mbtrCache {
		assert.NoError(t, mbtr.Close())
	}
}

func TestReader_ServeHTTP(t *testing.T) {
	mbtr := newReader(t, "testdata/openstreetmap.org.mbtiles")
	s := httptest.NewServer(http.StripPrefix("/", mbtr))
	defer s.Close()
	url := s.URL + "/0/0/0"
	res, err := http.Get(url)
	assert.NoError(t, err)
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	got, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	wantSHA256Sum := hexDecodeSHA256Sum(t, "075c660f81ba41146fda8610216a077b81bf5d8d102dbc893a57b7969e32ee88")
	assert.Equal(t, wantSHA256Sum, sha256.Sum256(got))
}
