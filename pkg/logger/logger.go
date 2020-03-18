package logger

import (
	"fmt"
	"io"

	"github.com/fatih/color"
)

type Logger struct {
	out      io.Writer
	loglevel int
}

func NewLogger(out io.Writer, loglevel int) Logger {
	return Logger{
		out:      out,
		loglevel: loglevel,
	}
}

func (l Logger) Debug(msg string, args ...interface{}) {
	if l.loglevel == 0 {
		return
	}
	if msg == "" {
		fmt.Fprintln(l.out, "")
		return
	}

	fmt.Fprintln(l.out, fmt.Sprintf(msg, args...))
}

func (l Logger) Info(msg string, args ...interface{}) {
	if msg == "" {
		fmt.Fprintln(l.out, "")
		return
	}

	c := color.New(color.FgHiCyan)
	c.Fprintln(l.out, fmt.Sprintf(msg, args...))
}

func (l Logger) Error(err error) {
	c := color.New(color.FgHiRed)
	c.Fprintln(l.out, fmt.Sprintf("%#v", err))
}

func (l Logger) Instructions(msg string, args ...interface{}) {
	white := color.New(color.FgHiWhite)
	white.Fprintln(l.out, "")
	white.Fprintln(l.out, fmt.Sprintf(msg, args...))
}
