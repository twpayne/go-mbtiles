package mbtiles

// MetadataJSON is the metadata required by the mbtiles spec to be present for vector
// tiles in the metadata table. This struct should be marshaled to a UTF-8 string
// and the value stored with key name `json`.
//
// See https://github.com/mapbox/mbtiles-spec/blob/master/1.3/spec.md#vector-tileset-metadata
type MetadataJSON struct {
	VectorLayers []MetadataJSONVectorLayer `json:"vector_layers"` // Defines the vector layers in this mbtiles file. Required for vector tilesets (mvt/pbf). Not necessary for raster image (png/webp/jpg/etc) mbtiles files.
}

// MetadataJSONVectorLayer contains information about each vector tile layer in this mbtiles file.
type MetadataJSONVectorLayer struct {
	ID          *string           `json:"id"`                    // The layer ID, which is referred to as the name of the layer in the Mapbox Vector Tile spec.
	Description *string           `json:"description,omitempty"` // A human-readable description of the layer's contents.
	MinZoom     *int              `json:"minzoom,omitempty"`     // The lowest zoom level whose tiles this layer appears in.
	MaxZoom     *int              `json:"maxzoom,omitempty"`     // The highest zoom level whose tiles this layer appears in.
	Fields      map[string]string `json:"fields"`                // Fields has keys and values are the names and types of attributes available in this layer. Each type MUST be the string "Number", "Boolean", or "String". Attributes whose type varies between features SHOULD be listed as "String".
}
