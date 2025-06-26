package ui

import (
	"bytes"
	"io"
	"testing"

	"github.com/rliebz/ghost"
	"github.com/rliebz/ghost/be"
)

func withStdout(l *Logger, out io.Writer) {
	l.stdout = out
}

func withStderr(l *Logger, err io.Writer) {
	l.stderr = err
}

type printTestCase struct {
	name            string
	setOutput       func(l *Logger, out io.Writer)
	printFunc       func(l *Logger)
	levelNoOutput   Level
	levelWithOutput Level
	expected        string
}

func testPrint(t *testing.T, tt printTestCase) {
	g := ghost.New(t)

	empty := new(bytes.Buffer)
	t.Cleanup(func() {
		g.Should(be.Zero(empty.String()))
	})

	logger := New(Config{
		Stdout: empty,
		Stderr: empty,
	})

	buf := new(bytes.Buffer)
	tt.setOutput(logger, buf)

	t.Run(tt.levelNoOutput.String(), func(t *testing.T) {
		g := ghost.New(t)

		logger.level = tt.levelNoOutput
		tt.printFunc(logger)
		g.Should(be.Zero(buf.String()))
	})

	buf.Reset()

	t.Run(tt.levelWithOutput.String(), func(t *testing.T) {
		g := ghost.New(t)

		logger.level = tt.levelWithOutput
		tt.printFunc(logger)
		g.Should(be.Equal(buf.String(), tt.expected))
	})
}

func TestVerbosityLevel_String(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{LevelSilent, "Silent"},
		{LevelQuiet, "Quiet"},
		{LevelNormal, "Normal"},
		{LevelVerbose, "Verbose"},
		{Level(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			g := ghost.New(t)

			g.Should(be.Equal(tt.level.String(), tt.want))
		})
	}
}
