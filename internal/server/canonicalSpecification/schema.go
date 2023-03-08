package server

import (
	"bytes"
	"embed"
	"encoding/json"
	"json-schema-validation/lib/tkt"
)

//go:embed schemas/*
var schemas embed.FS

type CanonicalSchema struct {
	Version string
	Schema  string
}

func InitializeCanonicalSchema() *CanonicalSchema {
	//for now version will be hardcoded
	version := "schemas/example.json"
	data, err := schemas.ReadFile(version)
	tkt.CheckErr(err)

	buf := bytes.Buffer{}
	err = json.Compact(&buf, data)
	tkt.CheckErr(err)
	schema := buf.String()

	return &CanonicalSchema{
		Version: version,
		Schema:  schema,
	}

}
