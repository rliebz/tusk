package appcli

import (
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
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			got, err := ioutil.ReadFile(tt.path)
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(got, tt.want) {
				t.Errorf("completions out of date between in-memory file and %q", tt.path)
			}
		})
	}

}
