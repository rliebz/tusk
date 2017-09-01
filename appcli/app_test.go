package appcli

import (
	"reflect"
	"testing"
)

func TestNewFlagApp(t *testing.T) {
	cfgText := []byte(`options:
  foo:
    short: f
    default: foovalue

tasks:
  mytask:
    run:
      - command: echo ${foo}
`)

	flagApp, err := newFlagApp(cfgText)
	if err != nil {
		t.Fatalf(
			"newFlagApp():\nconfig: `%s`\nunexpected err: %s",
			string(cfgText), err,
		)
	}

	args := []string{"tusk", "mytask", "--foo", "other"}
	if err = flagApp.Run(args); err != nil {
		t.Fatalf(
			"flagApp.Run():\nconfig: `%s`\nunexpected err: %s",
			string(cfgText), err,
		)
	}

	actual, ok := flagApp.Metadata["flagsPassed"].(map[string]string)
	if !ok {
		t.Fatalf(
			"flagApp.Metadata:\nconfig: `%s`\nMetadata did not contain flagsPassed",
			string(cfgText),
		)
	}

	expected := map[string]string{
		"foo": "other",
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf(
			"flagApp.Metadata for args(%s):\n expected: %v\nactual: %v",
			args, expected, actual,
		)
	}
}
