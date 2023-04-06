package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/rliebz/ghost"
	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

func TestJSONSchema(t *testing.T) {
	g := ghost.New(t)

	compiler := jsonschema.NewCompiler()

	schemaFile, err := os.Open("tusk.schema.yaml")
	g.NoError(err)
	t.Cleanup(func() { g.NoError(schemaFile.Close()) })

	var schemaData any
	err = yaml.NewDecoder(schemaFile).Decode(&schemaData)
	g.NoError(err)

	jsonSchema, err := json.Marshal(schemaData)
	g.NoError(err)

	err = compiler.AddResource("tusk.schema.json", bytes.NewReader(jsonSchema))
	g.NoError(err)

	schema, err := compiler.Compile("tusk.schema.json")
	g.NoError(err)

	tuskFile, err := os.Open("tusk.yml")
	g.NoError(err)
	t.Cleanup(func() { g.NoError(tuskFile.Close()) })

	var v any
	err = yaml.NewDecoder(tuskFile).Decode(&v)
	g.NoError(err)

	err = schema.Validate(v)
	var vde *jsonschema.ValidationError
	if errors.As(err, &vde) {
		b, err := json.MarshalIndent(vde.DetailedOutput(), "", "  ")
		g.NoError(err)

		t.Fatal(string(b))
	}
}
