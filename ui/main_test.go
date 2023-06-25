package ui

import (
	"bytes"
	"io"
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
)

func withStdout(l *Logger, out io.Writer) {
	l.Stdout = out
}

func withStderr(l *Logger, err io.Writer) {
	l.Stderr = err
}

type printTestCase struct {
	name            string
	setOutput       func(l *Logger, out io.Writer)
	printFunc       func(l *Logger)
	levelNoOutput   VerbosityLevel
	levelWithOutput VerbosityLevel
	expected        string
}

func testPrint(t *testing.T, tt printTestCase) {
	g := ghost.New(t)

	empty := new(bytes.Buffer)
	t.Cleanup(func() {
		g.Should(be.Zero(empty.String()))
	})

	logger := New()
	logger.Stderr = empty
	logger.Stdout = empty

	buf := new(bytes.Buffer)
	tt.setOutput(logger, buf)

	t.Run(tt.levelNoOutput.String(), func(t *testing.T) {
		g := ghost.New(t)

		logger.Verbosity = tt.levelNoOutput
		tt.printFunc(logger)
		g.Should(be.Zero(buf.String()))
	})

	buf.Reset()

	t.Run(tt.levelWithOutput.String(), func(t *testing.T) {
		g := ghost.New(t)

		logger.Verbosity = tt.levelWithOutput
		tt.printFunc(logger)
		g.Should(be.Equal(tt.expected, buf.String()))
	})
}

func TestVerbosityLevel_String(t *testing.T) {
	tests := []struct {
		level VerbosityLevel
		want  string
	}{
		{VerbosityLevelSilent, "Silent"},
		{VerbosityLevelQuiet, "Quiet"},
		{VerbosityLevelNormal, "Normal"},
		{VerbosityLevelVerbose, "Verbose"},
		{VerbosityLevel(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			g := ghost.New(t)

			g.Should(be.Equal(tt.want, tt.level.String()))
		})
	}
}
