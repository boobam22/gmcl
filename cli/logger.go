package cli

import (
	"io"
	"log"
	"os"
)

type Logger struct {
	File  *os.File
	debug *log.Logger
	info  *log.Logger
	error *log.Logger
}

func NewLogger(path string) (*Logger, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	return &Logger{
		File:  file,
		debug: log.New(file, "[DEBUG] ", log.LstdFlags),
		info:  log.New(io.MultiWriter(file, os.Stdout), "[INFO ] ", log.LstdFlags),
		error: log.New(io.MultiWriter(file, os.Stderr), "[ERROR] ", log.LstdFlags),
	}, nil
}

func (l *Logger) Close() error {
	if l.File != nil {
		return l.File.Close()
	}
	return nil
}

func (l *Logger) Debug(msg string) { l.debug.Println(msg) }
func (l *Logger) Info(msg string)  { l.info.Println(msg) }
func (l *Logger) Error(msg string) { l.error.Println(msg) }

func (l *Logger) Debugf(format string, args ...any) { l.debug.Printf(format, args...) }
func (l *Logger) Infof(format string, args ...any)  { l.info.Printf(format, args...) }
func (l *Logger) Errorf(format string, args ...any) { l.error.Printf(format, args...) }
