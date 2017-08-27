package ui

import (
	"log"
	"os"
)

var (
	stdout = log.New(os.Stdout, "", 0)
	stderr = log.New(os.Stderr, "", 0)
)

func println(l *log.Logger, v ...interface{}) {
	l.Println(v...)
	HasPrinted = true
}

func printf(l *log.Logger, format string, v ...interface{}) {
	l.Printf(format, v...)
	HasPrinted = true
}
