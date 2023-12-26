package log

import (
	"io"
	stdlog "log"
	"os"
)

const logFlags = stdlog.Ldate | stdlog.Ltime | stdlog.Lmicroseconds | stdlog.Lshortfile | stdlog.LUTC | stdlog.Lmsgprefix

// copies methods of standard library log
type Printer interface {
	Print(v ...any)
	Printf(format string, v ...any)
	Println(v ...any)
}

type Logger interface {
	Debug() Printer
	Info() Printer
	Warn() Printer
	Error() Printer
	Prefix(prefix string) Logger
}

type logger struct {
	debug, info, warn, error *stdlog.Logger
}

func NewDefaultLogger() Logger {
	return &logger{
		// io.Discard is magic value, which prevents arguments from being formatted to a string
		debug: stdlog.New(io.Discard, "DEBUG ", logFlags),
		info:  stdlog.New(os.Stdout, "INFO ", logFlags),
		warn:  stdlog.New(os.Stdout, "WARN ", logFlags),
		error: stdlog.New(os.Stderr, "ERROR ", logFlags),
	}
}

func (l *logger) Debug() Printer { return l.debug }

func (l *logger) Error() Printer { return l.error }

func (l *logger) Info() Printer { return l.info }

func (l *logger) Warn() Printer { return l.warn }

func (l *logger) Prefix(prefix string) Logger {
	return &logger{
		debug: stdlog.New(l.debug.Writer(), l.debug.Prefix()+prefix+" ", l.debug.Flags()),
		info:  stdlog.New(l.info.Writer(), l.info.Prefix()+prefix+" ", l.info.Flags()),
		warn:  stdlog.New(l.warn.Writer(), l.warn.Prefix()+prefix+" ", l.warn.Flags()),
		error: stdlog.New(l.error.Writer(), l.error.Prefix()+prefix+" ", l.error.Flags()),
	}
}
