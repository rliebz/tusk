package config

import "testing"

func TestParseFileFlag_no_args(t *testing.T) {
	args := []string{}
	_, passed := parseFileFlag(args)
	if passed {
		t.Fatalf("expected: %#v, got: %#v", false, passed)
	}
}

func TestParseFileFlag_short(t *testing.T) {
	filename := "file.yml"

	args := []string{
		"-f",
		filename,
	}
	tuskfile, passed := parseFileFlag(args)
	if !passed {
		t.Fatalf("expected: %#v, got: %#v", true, passed)
	}

	if tuskfile != filename {
		t.Fatalf("expected: %#v, got: %#v", filename, tuskfile)
	}
}

func TestParseFileFlag_long(t *testing.T) {
	filename := "file.yml"

	args := []string{
		"--file",
		filename,
	}
	tuskfile, passed := parseFileFlag(args)
	if !passed {
		t.Fatalf("expected: %#v, got: %#v", true, passed)
	}

	if tuskfile != filename {
		t.Fatalf("expected: %#v, got: %#v", filename, tuskfile)
	}
}

func TestParseFileFlag_repeated(t *testing.T) {
	preferred := "preferred.yml"

	args := []string{
		"-f",
		preferred,
		"-f",
		"other.yml",
	}
	tuskfile, passed := parseFileFlag(args)
	if !passed {
		t.Fatalf("expected: %#v, got: %#v", true, passed)
	}

	if tuskfile != preferred {
		t.Fatalf("expected: %#v, got: %#v", preferred, tuskfile)
	}
}
