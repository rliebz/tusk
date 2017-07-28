package task

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// When defines the conditions for running a task.
type When struct {
	Exists []string `yaml:",omitempty"`
	OS     []string `yaml:",omitempty"`
	Test   []string `yaml:",omitempty"`
}

// Validate returns an error if any when clauses fail.
func (w *When) Validate() error {
	for _, f := range w.Exists {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			return fmt.Errorf("file %s does not exist", f)
		}
	}

	if err := validateOS(runtime.GOOS, w.OS); err != nil {
		return err
	}

	for _, test := range w.Test {
		if err := testCommand(test); err != nil {
			return fmt.Errorf("test failed: %s", test)
		}
	}

	return nil
}

func validateOS(os string, required []string) error {
	// Nothing specified means any OS is fine
	if len(required) == 0 {
		return nil
	}

	// Otherwise, at least one must match
	for _, r := range required {
		if os == normalizeOS(r) {
			return nil
		}
	}

	return fmt.Errorf("current OS %s not listed in %v", os, required)
}

func normalizeOS(os string) string {
	lower := strings.ToLower(os)

	for _, alt := range []string{"mac", "macos", "osx"} {
		if lower == alt {
			return "darwin"
		}
	}

	for _, alt := range []string{"win"} {
		if lower == alt {
			return "windows"
		}
	}

	return lower
}
