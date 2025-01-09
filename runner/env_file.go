package runner

import (
	"maps"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"

	"github.com/rliebz/tusk/marshal"
)

// EnvFile is a dotenv file that should be parsed during configuration start.
type EnvFile struct {
	Path     string `yaml:"path"`
	Required bool   `yaml:"required"`
}

// UnmarshalYAML allows a string to represent a required path.
func (f *EnvFile) UnmarshalYAML(unmarshal func(any) error) error {
	var path string
	pathCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&path) },
		Assign:    func() { *f = EnvFile{Path: path, Required: true} },
	}

	type envFileType EnvFile // Use new type to avoid recursion
	var envFileItem envFileType
	envFileCandidate := marshal.UnmarshalCandidate{
		Unmarshal: func() error { return unmarshal(&envFileItem) },
		Assign:    func() { *f = EnvFile(envFileItem) },
	}

	return marshal.UnmarshalOneOf(pathCandidate, envFileCandidate)
}

// loadEnvFiles sets env vars from a set of file configs. Values are not
// overridden.
//
// If no files are specified, it will load from an optional default of .env.
// If an empty list is specified, no files will be loaded.
func loadEnvFiles(dir string, envFiles []EnvFile) error {
	// An explicit [] is an obvious attempt to remove the default, so check only
	// for nilness.
	if envFiles == nil {
		envFiles = []EnvFile{{Path: ".env", Required: false}}
	}

	envMap := make(map[string]string)
	for _, envFile := range envFiles {
		m, err := readEnvFile(dir, envFile)
		if err != nil {
			return err
		}

		maps.Copy(envMap, m)
	}

	for k, v := range envMap {
		if _, ok := os.LookupEnv(k); !ok {
			if err := os.Setenv(k, v); err != nil {
				return err
			}
		}
	}

	return nil
}

func readEnvFile(dir string, envFile EnvFile) (map[string]string, error) {
	m, err := godotenv.Read(filepath.Join(dir, envFile.Path))
	switch {
	case !envFile.Required && os.IsNotExist(err):
	case err != nil:
		return nil, err
	}

	return m, nil
}
