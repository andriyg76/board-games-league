package glog

import (
	"fmt"
	"io"
	"log"
	"os"
)

type Output interface {
	Printf(format string, a ...interface{})
}

type TraceLogger interface {
	Trace(format string, a ...interface{})
	TraceLogger() Output
	IsTrace() bool
}

type DebugLogger interface {
	Debug(format string, a ...interface{})
	IsDebug() bool
	DebugLogger() Output
}

type InfoLogger interface {
	Info(format string, a ...interface{})
	IsInfo() bool
}

type WarnLogger interface {
	Warn(format string, a ...interface{})
	IsWarn() bool
}

type ErrorLogger interface {
	Error(format string, a ...interface{}) error
	IsError() bool
}

type Logger interface {
	DebugLogger
	TraceLogger
	WarnLogger
	InfoLogger
	ErrorLogger

	// Log formats according to a format specifier
	Log(LogLevel LogLevel, format string, a ...interface{})
	IsEnabled(logLevel LogLevel) bool
	GetOutput(LogLevel LogLevel) Output

	Panic(format string, a ...interface{})
	Fatal(format string, a ...interface{})
}

type logger struct {
	logLevel LogLevel
	out      *log.Logger
	err      *log.Logger
	fatalf   func(format string, a ...interface{})
}

func (l logger) IsDebug() bool {
	return l.IsEnabled(DEBUG)
}

func (l logger) IsTrace() bool {
	return l.IsEnabled(TRACE)
}

func (l logger) IsWarn() bool {
	return l.IsEnabled(WARN)
}

func (l logger) IsInfo() bool {
	return l.IsEnabled(INFO)
}

func (l logger) IsError() bool {
	return l.IsEnabled(ERROR)
}

func (l logger) IsEnabled(logLevel LogLevel) bool {
	return logLevel.weight >= l.logLevel.weight
}

func create(logLevel LogLevel) logger {
	return logger{
		logLevel: logLevel,
		err:      _stderr,
		out:      _stdout,
		fatalf:   _stderr.Fatalf,
	}
}

func createWithWriters(outWriter, errWriter io.Writer, logLevel LogLevel) logger {
	if outWriter == nil {
		outWriter = io.Discard
	}
	if errWriter == nil {
		errWriter = io.Discard
	}

	out := log.New(outWriter, "", log.LstdFlags)
	err := log.New(errWriter, "", log.LstdFlags)

	return logger{
		logLevel: logLevel,
		err:      err,
		out:      out,
		fatalf:   err.Fatalf,
	}
}

func createFileLogger(file string, level ...LogLevel) (logger, error) {
	logLevel := INFO
	if len(level) > 0 {
		logLevel = level[0]
	}

	var instance = logger{
		logLevel: logLevel,
	}
	openFile, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return instance, _default.Error("Error creating file %s output: %s", file, err)
	}
	w := log.New(openFile, "", log.LstdFlags)
	instance.err = w
	instance.out = w
	instance.fatalf = w.Fatalf
	return instance, nil
}

func Create(logLevel LogLevel) Logger {
	return create(logLevel)
}

var _stdout = log.New(os.Stdout, "", log.LstdFlags)
var _stderr = log.New(os.Stderr, "", log.LstdFlags)

type dumbLogger struct{}

func (d dumbLogger) Printf(format string, objs ...interface{}) {}

var dumbLoggerInstance = dumbLogger{}

type loggerWithLevel struct {
	logLevel *LogLevel
	logger   *logger
}

func (l loggerWithLevel) Printf(format string, objs ...interface{}) {
	l.logger.Log(*l.logLevel, format, objs...)
}

func (l logger) GetOutput(logLevel LogLevel) Output {
	return loggerWithLevel{
		logLevel: &logLevel,
		logger:   &l,
	}
}

func (l logger) TraceLogger() Output {
	return l.GetOutput(TRACE)
}

func (l logger) Log(logLevel LogLevel, format string, objs ...interface{}) {
	logFormat := logLevel.prefix + " " + format

	if logLevel == PANIC {
		l.err.Panicf(logFormat, objs...)
		return
	}
	if logLevel == FATAL {
		l.fatalf(logFormat, objs...)
		return
	}

	var out Output
	if logLevel.weight < l.logLevel.weight {
		out = dumbLoggerInstance
	} else if logLevel.weight >= WARN.weight {
		out = l.err
	} else {
		out = l.out
	}

	out.Printf(logFormat, objs...)
}

func (l logger) Debug(format string, objs ...interface{}) {
	l.Log(DEBUG, format, objs...)
}

func (l logger) DebugLogger() Output {
	return l.GetOutput(DEBUG)
}

func (l logger) Trace(format string, objs ...interface{}) {
	l.Log(TRACE, format, objs...)
}

func (l logger) Info(format string, objs ...interface{}) {
	l.Log(INFO, format, objs...)
}

func (l logger) Warn(format string, objs ...interface{}) {
	l.Log(WARN, format, objs...)
}

func (l logger) Error(format string, objs ...interface{}) error {
	err := fmt.Errorf(format, objs...)
	l.Log(ERROR, "%s", err)
	return err
}

func (l logger) Panic(format string, objs ...interface{}) {
	l.Log(PANIC, format, objs...)
}

func (l logger) Fatal(format string, objs ...interface{}) {
	l.Log(FATAL, format, objs...)
}
