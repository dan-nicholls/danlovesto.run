package log

import (
	"sync"
	"log"
	"io"
)

type Logger interface { 
	Infof(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

type StdLogger struct {
	mu sync.Mutex
	std *log.Logger
}

func NewLogger(out io.Writer, flag int) *StdLogger {
	return &StdLogger{
		std: log.New(out, "", flag),
	}
}

func (l *StdLogger) log(prefix string, format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.std.Printf(prefix+" "+format, v...)
}

func (l *StdLogger) Infof(format string, v ...interface{}) {
	l.log("[INFO]", format, v...)
}

func (l *StdLogger) Errorf(format string, v ...interface{}) {
	l.log("[ERROR]", format, v...)
}
