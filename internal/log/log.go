package log

import (
	"io"
	stdlog "log"
	"os"
)

const logFlags = stdlog.Ldate | stdlog.Ltime | stdlog.Lmicroseconds | stdlog.Lshortfile | stdlog.LUTC | stdlog.Lmsgprefix

func NewDebug(prefix string) *stdlog.Logger {
	// hint: replace io.Discard with the real writer when need verbose logging
	return stdlog.New(io.Discard, "[DEBUG] "+prefix, logFlags)
}

func NewInfo(prefix string) *stdlog.Logger {
	return stdlog.New(os.Stdout, "[INFO] "+prefix, logFlags)
}

func NewWarn(prefix string) *stdlog.Logger {
	return stdlog.New(os.Stdout, "[WARN] "+prefix, logFlags)
}

func NewError(prefix string) *stdlog.Logger {
	return stdlog.New(os.Stderr, "[ERROR] "+prefix, logFlags)
}
