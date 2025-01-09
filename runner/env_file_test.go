package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
	"gopkg.in/yaml.v2"
)

func TestEnvFile_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name string
		data string
		want EnvFile
	}{
		{
			name: "empty",
			data: ``,
			want: EnvFile{},
		},
		{
			name: "string",
			data: `"dot.env"`,
			want: EnvFile{
				Path:     "dot.env",
				Required: true,
			},
		},
		{
			name: "object required",
			data: `{path: dot.env, required: true}`,
			want: EnvFile{
				Path:     "dot.env",
				Required: true,
			},
		},
		{
			name: "object not required",
			data: `{path: dot.env, required: false}`,
			want: EnvFile{
				Path:     "dot.env",
				Required: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := ghost.New(t)

			var envFile EnvFile
			err := yaml.UnmarshalStrict([]byte(tt.data), &envFile)
			g.NoError(err)

			g.Should(be.Equal(envFile, tt.want))
		})
	}
}

func Test_loadEnvFiles(t *testing.T) {
	t.Run("default used", func(t *testing.T) {
		g := ghost.New(t)
		stashEnv(t)
		tmpdir := useTempDir(t)

		t.Setenv("BAZ", "bazvalue")

		// Just use a little bit of all the fancy syntax
		data := []byte(`
# comment
FOO=foovalue
export BAR=barvalue
BAZ=newvalue
QUUX=${FOO}
`)

		err := os.WriteFile(filepath.Join(tmpdir, ".env"), data, 0o644)
		g.NoError(err)

		err = loadEnvFiles(tmpdir, nil)
		g.NoError(err)

		g.Should(be.All(
			be.Equal(os.Getenv("FOO"), "foovalue"),
			be.Equal(os.Getenv("BAR"), "barvalue"),
			be.Equal(os.Getenv("BAZ"), "bazvalue"),
			be.Equal(os.Getenv("QUUX"), "foovalue"),
		))
	})

	t.Run("default not found", func(t *testing.T) {
		g := ghost.New(t)
		stashEnv(t)
		tmpdir := useTempDir(t)

		err := loadEnvFiles(tmpdir, nil)
		g.NoError(err)
	})

	t.Run("dev null", func(t *testing.T) {
		g := ghost.New(t)
		stashEnv(t)
		tmpdir := useTempDir(t)

		err := loadEnvFiles(tmpdir, []EnvFile{{Path: "/dev/null"}})
		g.NoError(err)
	})

	t.Run("empty list", func(t *testing.T) {
		g := ghost.New(t)
		stashEnv(t)
		tmpdir := useTempDir(t)

		t.Setenv("BAZ", "bazvalue")

		err := os.WriteFile(filepath.Join(tmpdir, ".env"), []byte(`FOO=foovalue`), 0o644)
		g.NoError(err)

		err = loadEnvFiles(tmpdir, []EnvFile{})
		g.NoError(err)

		g.Should(be.Zero(os.Getenv("FOO")))
	})

	t.Run("required found", func(t *testing.T) {
		g := ghost.New(t)
		stashEnv(t)
		tmpdir := useTempDir(t)

		err := os.WriteFile(filepath.Join(tmpdir, ".env"), []byte("FOO=foovalue"), 0o644)
		g.NoError(err)

		err = loadEnvFiles(tmpdir, []EnvFile{
			{Path: ".env", Required: true},
		})
		g.NoError(err)

		g.Should(be.Equal(os.Getenv("FOO"), "foovalue"))
	})

	t.Run("required not found", func(t *testing.T) {
		g := ghost.New(t)
		stashEnv(t)
		tmpdir := useTempDir(t)

		err := loadEnvFiles(tmpdir, []EnvFile{
			{Path: ".env", Required: true},
		})
		g.Should(be.ErrorIs(err, os.ErrNotExist))
		g.Should(be.True(os.IsNotExist(err)))
	})

	t.Run("directory respected", func(t *testing.T) {
		g := ghost.New(t)
		stashEnv(t)

		// Write the file we plan to use
		tmpdir := useTempDir(t)
		err := os.WriteFile(filepath.Join(tmpdir, ".env"), []byte("FOO=foovalue"), 0o644)
		g.NoError(err)

		// Navigate to a directory where the .env file is NOT located
		useTempDir(t)

		err = loadEnvFiles(tmpdir, []EnvFile{
			{Path: ".env", Required: true},
		})
		g.NoError(err)

		g.Should(be.Equal(os.Getenv("FOO"), "foovalue"))
	})

	t.Run("overrides earlier values", func(t *testing.T) {
		g := ghost.New(t)
		stashEnv(t)
		tmpdir := useTempDir(t)

		err := os.WriteFile(filepath.Join(tmpdir, "1.env"), []byte("FOO=one"), 0o644)
		g.NoError(err)
		err = os.WriteFile(filepath.Join(tmpdir, "2.env"), []byte("FOO=two"), 0o644)
		g.NoError(err)
		err = os.WriteFile(filepath.Join(tmpdir, "3.env"), []byte("BAR=three"), 0o644)
		g.NoError(err)

		err = loadEnvFiles(tmpdir, []EnvFile{
			{Path: "1.env"},
			{Path: "2.env"},
			{Path: "3.env"},
		})
		g.NoError(err)

		g.Should(be.Equal(os.Getenv("FOO"), "two"))
		g.Should(be.Equal(os.Getenv("BAR"), "three"))
	})
}
