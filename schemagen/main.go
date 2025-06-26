package main

import (
	"encoding/json"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	schemaFile, err := os.Open("./tusk.schema.yaml")
	if err != nil {
		return err
	}
	defer schemaFile.Close() //nolint:errcheck

	var schemaData any
	err = yaml.NewDecoder(schemaFile).Decode(&schemaData)
	if err != nil {
		return err
	}

	outfile, err := os.Create("./tusk.schema.json")
	if err != nil {
		return err
	}
	defer outfile.Close() //nolint:errcheck

	enc := json.NewEncoder(outfile)
	enc.SetIndent("", "\t")

	err = enc.Encode(schemaData)
	if err != nil {
		return err
	}

	return outfile.Close()
}
