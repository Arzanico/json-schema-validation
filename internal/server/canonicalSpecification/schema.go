package server

import (
	"embed"
	"json-schema-validation/lib/tkt"
)

//go:embed schemas/*
var schemas embed.FS

type CanonicalSchema struct {
	Version string
	Schema  []byte
}

func InitializeCanonicalSchema() *CanonicalSchema {
	//Version refers to the schema file that will be loaded.
	//For now, version will be hardcoded, in future it could be a parameter
	//version := "schemas/example.json"
	version := "schemas/canonical_v2.json"
	data, err := schemas.ReadFile(version)
	tkt.CheckErr(err)

	/*buf := bytes.Buffer{}
	err = json.Compact(&buf, data)
	tkt.CheckErr(err)
	schema := buf.String()*/

	return &CanonicalSchema{
		Version: version,
		Schema:  data,
	}
}
