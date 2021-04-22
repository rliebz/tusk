package appcli

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCompletionsUpToDate(t *testing.T) {
	tests := []struct {
		shell string
		path  string
		want  []byte
	}{
		{"zsh", "../completion/_tusk", []byte(rawZshCompletion)},
		{"fish", "../completion/tusk.fish", []byte(rawFishCompletion)},
		{"bash", "../completion/tusk-completion.bash", []byte(rawBashCompletion)},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			contents, err := ioutil.ReadFile(tt.path)
			if err != nil {
				t.Fatal(err)
			}

			// Ignore windows line endings
			got := bytes.ReplaceAll(contents, []byte("\r\n"), []byte("\n"))

			if !cmp.Equal(got, tt.want) {
				t.Errorf("completions out of date between in-memory file and %q", tt.path)
			}
		})
	}
}
