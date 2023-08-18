# go-mbtiles

[![Build
Status](https://travis-ci.org/twpayne/go-mbtiles.svg?branch=master)](https://travis-ci.org/twpayne/go-mbtiles)
[![GoDoc](https://godoc.org/github.com/twpayne/go-mbtiles?status.svg)](https://godoc.org/github.com/twpayne/go-mbtiles)
[![Report
Card](https://goreportcard.com/badge/github.com/twpayne/go-mbtiles)](https://goreportcard.com/report/github.com/twpayne/go-mbtiles)

Package `mbtiles` reads and writes files in the [MBTiles
format](https://github.com/mapbox/mbtiles-spec).

## Running the http server

```
go install "github.com/twpayne/go-mbtiles/cmd/mbtiles-server"
go run "github.com/twpayne/go-mbtiles/cmd/mbtiles-server" -addr localhost:9091 -dsn ./testdata/openstreetmap.org.mbtiles
```

## Reading mbtiles files

MBTiles files are SQLite databases, and opened using a DSN string.

Create a reader and read a tile with:

```golang
reader, err := mbtiles.NewReader("./testdata/openstreetmap.org.mbtiles")
if err != nil {
	panic(err)
}
tile, err := reader.SelectTile(0, 0, 0)
if err != nil {
	if errors.Is(err, sql.ErrNoRows) {
		fmt.Printf("tile doesn't exist: %d, %d, %d\n", 0, 0, 0)
		return
	} else {
		panic(err)
	}
}
fmt.Printf("tile data: %+v\n", tile)
```

Note that SQLite will happily open a non-existant file to read without
throwing an error. It is recommend that you check for existance of the file
first, if you're using a file path as your DSN. For example:

```golang
if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
	panic(fmt.Sprintf("mbtiles file doesn't exist (%s): %v", filename, err))
}
```

The `Reader` type includes a `ServeHTTP` function. You can use it to create a
reader for your mbtiles.

## Writing mbtiles files

This package supports writing files in the [MBTiles
v1.3](https://github.com/mapbox/mbtiles-spec/blob/master/1.3/spec.md) format.

The library will not fill out all the required fields in the MBTiles
specification, instead it provides the tools to be compliant.

Of note:
* The caller is responsible for populating the correct metadata into the
  metadata persuant to the spec.
* The caller is responsible for gzip'ing the tile data before calling
  `InsertTile` or `BulkInsertTile`. The spec requires tiles to be compressed
  with gzip. How the caller implements the compression is outside the scope of
  this package.
* For the `json` key in the metadata table, helper types are provided in this
  package as `mbtiles.MetadataJson`. This type can be marshaled to a string
  and inserted into the metadata table for spec compliance for vector MBTiles
  files.
* go-mbtiles will invert the Y coordinate to TMS to be compliant with the
  mbtiles spec.
* go-mbtiles will create a metadata table if it doesn't exist, the first time
  `InsertMetadata` is called.
* go-mbtiles will create a tiles table if it doesn't exist, the first time
  `InsertTile` or `BulkInsertTile` is called.


### Performance Tips

MBTiles files are SQLite databases. To performantly bulk insert a large number
of rows into the database, certain optimizations may be necessary to improve
write performance.

SQLite is a a single writer database, which means only one write can occur at
a time. The database will otherwise be locked. Because of this, any write that
is not in a transaction will automatically be wrapped in a transaction, which
is slow. Consider the `BulkInsertTile` command, which will wrap all of the
inserts in a single transaction.

There are other optimizations exposed through the `Writer` interface. You
should understand their implications for your use case before turning them on.
* `JournalModeMemory` switches journaling from disk to memory. In bulk import
  scenarios, this is likely a very safe performance optimization to turn on.
* `SynchronousOff` allows SQLite to continue processing as soon as data is
  handed off to the operating system to be written (instead of wait for
  confirmation that the write was successful). This is likely safe for bulk
  writes, but likely will result in a corrupted database if the process is
  interrupted or computer loses power.

The performance improvements in go-mbtiles are motivated by the research in
this [StackOverflow post about SQLite INSERT
performance](https://stackoverflow.com/questions/1711631/improve-insert-per-second-performance-of-sqlite).

## License

BSD-2-Clause in [LICENCE](./LICENSE).

## Contributors
* Tom Payne (@twpayne)
* Joe Polastre (@polastre), FlightAware
