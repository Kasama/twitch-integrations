package logger

import (
	"fmt"
	"os"

	"github.com/labstack/gommon/log"
	"github.com/mattn/go-isatty"
)

const RESET = "\u001b[0m"
const CYAN = "\u001b[36m"
const GREEN = "\u001b[32m"

type Logger interface {
	Info(message ...any)
	Infof(format string, args ...any)
	Debug(message ...any)
	Debugf(format string, args ...any)
	Warn(message ...any)
	Warnf(format string, args ...any)
	Error(message ...any)
	Errorf(format string, args ...any)

	WithField(key string, value any) Logger
	WithFields(map[string]any) Logger
}

type field struct {
	key   string
	value any
}

type logger struct {
	log         *log.Logger
	fields      []field
	fieldFormat string
}

// Debug implements Logger.
func (l *logger) Debug(message ...any) {
	for _, field := range l.fields {
		message = append(message, fmt.Sprintf(l.fieldFormat, field.key, field.value))
	}

	l.log.Debug(message...)
}

// Debugf implements Logger.
func (l *logger) Debugf(format string, args ...any) {
	for _, field := range l.fields {
		args = append(args, fmt.Sprintf(l.fieldFormat, field.key, field.value))
	}

	l.log.Debugf(format, args...)
}

// Error implements Logger.
func (l *logger) Error(message ...any) {
	for _, field := range l.fields {
		message = append(message, fmt.Sprintf(l.fieldFormat, field.key, field.value))
	}

	l.log.Error(message...)
}

// Errorf implements Logger.
func (l *logger) Errorf(format string, args ...any) {
	for _, field := range l.fields {
		args = append(args, fmt.Sprintf(l.fieldFormat, field.key, field.value))
	}

	l.log.Errorf(format, args...)
}

// Info implements Logger.
func (l *logger) Info(message ...any) {
	for _, field := range l.fields {
		message = append(message, fmt.Sprintf(l.fieldFormat, field.key, field.value))
	}

	l.log.Info(message...)
}

// Infof implements Logger.
func (l *logger) Infof(format string, args ...any) {
	for _, field := range l.fields {
		args = append(args, fmt.Sprintf(l.fieldFormat, field.key, field.value))
	}

	l.log.Infof(format, args...)
}

// Warn implements Logger.
func (l *logger) Warn(message ...any) {
	for _, field := range l.fields {
		message = append(message, fmt.Sprintf(l.fieldFormat, field.key, field.value))
	}

	l.log.Warn(message...)
}

// Warnf implements Logger.
func (l *logger) Warnf(format string, args ...any) {
	for _, field := range l.fields {
		args = append(args, fmt.Sprintf(l.fieldFormat, field.key, field.value))
	}

	l.log.Warnf(format, args...)
}

// WithField implements Logger.
func (l *logger) WithField(key string, value any) Logger {
	newLogger := &logger{log: l.log, fields: l.fields, fieldFormat: l.fieldFormat}
	newLogger.fields = append(newLogger.fields, field{key: key, value: value})
	return newLogger
}

// WithFields implements Logger.
func (l *logger) WithFields(fields map[string]any) Logger {
	newLogger := &logger{log: l.log, fields: l.fields, fieldFormat: l.fieldFormat}
	for k, v := range fields {
		newLogger.fields = append(newLogger.fields, field{key: k, value: v})
	}
	return newLogger
}

var _ Logger = &logger{}

func New(prefix string, lvl log.Lvl) Logger {
	logging := log.New(prefix)
	logging.SetLevel(lvl)
	format := `,"%s":"%v"`
	if isatty.IsTerminal(os.Stdout.Fd()) {
		format = "\t%s=%v"
		logging.SetHeader(`${time_rfc3339} ${prefix} ${level} -`)
		logging.EnableColor()
	}

	return &logger{log: logging, fields: []field{}, fieldFormat: format}
}

