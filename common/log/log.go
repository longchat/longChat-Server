package log

import (
	"errors"
	slog "log"
	"sync"

	"github.com/YueHonghui/rfw"
)

var ERROR *slog.Logger
var WARN *slog.Logger
var INFO *slog.Logger
var ACCESS *slog.Logger

type bufLogger struct {
	writer       *rfw.Rfw
	writerClosed bool
	writerlock   sync.RWMutex
	channel      chan []byte
}

var lwriter *bufLogger
var awriter *bufLogger

func InitLogger(logpath string, alogpath string, lchansize int, achansize int) error {
	var err error
	lwriter = &bufLogger{
		channel:      make(chan []byte, lchansize),
		writerClosed: false,
	}
	lwriter.writer, err = rfw.New(logpath)
	if err != nil {
		return err
	}
	awriter = &bufLogger{
		channel:      make(chan []byte, achansize),
		writerClosed: false,
	}
	awriter.writer, err = rfw.New(alogpath)
	if err != nil {
		return err
	}
	ERROR = slog.New(lwriter, "[ERR] ", slog.Ldate|slog.Ltime|slog.Lshortfile)
	WARN = slog.New(lwriter, "[WRN] ", slog.Ldate|slog.Ltime|slog.Lshortfile)
	INFO = slog.New(lwriter, "[INF] ", slog.Ldate|slog.Ltime|slog.Lshortfile)
	ACCESS = slog.New(awriter, "", slog.Ldate|slog.Ltime)
	lwriter.Writing(1)
	awriter.Writing(1)
	return nil
}

func (w *bufLogger) Write(p []byte) (int, error) {
	if !w.writerClosed {
		w.writerlock.RLock()
		if w.writerClosed {
			w.writerlock.RUnlock()
			return 0, errors.New("logger is closed")
		}
		line := make([]byte, len(p))
		copy(line, p)
		w.channel <- line
		w.writerlock.RUnlock()
		return len(line), nil
	}
	return 0, errors.New("logger is closed")
}

func (w *bufLogger) Writing(nworker int) {
	for i := 0; i < nworker; i++ {
		go func() {
			for !w.writerClosed {
				l, more := <-w.channel
				if !more || w.writerClosed {
					break
				} else {
					w.writerlock.RLock()
					if w.writerClosed {
						w.writerlock.RUnlock()
						break
					}
					w.writer.Write(l)
					w.writerlock.RUnlock()
				}
			}
		}()
	}
}

func (w *bufLogger) Close() {
	w.writerlock.Lock()
	defer w.writerlock.Unlock()
	w.writer.Close()
	w.writerClosed = true
	close(w.channel)
}

func FiniLogger() {
	awriter.Close()
	lwriter.Close()
}
