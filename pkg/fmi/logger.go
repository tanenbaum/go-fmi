package fmi

import (
	"errors"
	"fmt"
)

const (
	loggerCategoryNone   loggerCategory = 0
	loggerCategoryEvents loggerCategory = 1 << iota
	loggerCategoryWarning
	loggerCategoryDiscard
	loggerCategoryError
	loggerCategoryFatal
	loggerCategoryPending
	loggerCategoryAll = ^loggerCategory(0)
)

// Logger abstracts the fmi2CallbackLogger callback function
// Log messages are not sent if logging is disabled in fmi2Instantiate
type Logger interface {
	// Error logs an error to the FMU logger
	Error(err error)
	// Fatal logs a fatal error to the FMU logger
	Fatal(err error)
	// Warning logs warning to FMU logger
	Warning(msg string)
	// Discard logs discard message to FMU logger
	Discard(msg string)
	// Event logs event message to FMU logger
	Event(msg string)
	// Info logs info messages to FMU logger
	Info(msg string)

	setMask(mask loggerCategory)
}

type loggerCategory uint

var loggerCategories = map[loggerCategory]string{
	loggerCategoryEvents:  "logEvents",
	loggerCategoryWarning: "logStatusWarning",
	loggerCategoryDiscard: "logStatusDiscard",
	loggerCategoryError:   "logStatusError",
	loggerCategoryFatal:   "logStatusFatal",
	loggerCategoryPending: "logStatusPending",
	loggerCategoryAll:     "logAll",
}

func (l loggerCategory) String() string {
	if s, ok := loggerCategories[l]; ok {
		return s
	}
	return "unknown"
}

type logger struct {
	mask loggerCategory

	fmiCallbackLogger LoggerCallback
}

type LoggerCallback func(status Status, category, message string)

func newLogger(categories []string, callback LoggerCallback) (Logger, error) {
	mask := loggerCategory(0)
	for _, c := range categories {
		m, err := loggerCategoryFromString(c)
		if err != nil {
			return nil, err
		}
		mask |= m
	}
	return &logger{
		mask:              mask,
		fmiCallbackLogger: callback,
	}, nil
}

func (l logger) Error(err error) {
	l.logMessage(StatusError, loggerCategoryError, err.Error())
}

func (l logger) Fatal(err error) {
	l.logMessage(StatusFatal, loggerCategoryFatal, err.Error())
}

func (l logger) Warning(msg string) {
	l.logMessage(StatusWarning, loggerCategoryWarning, msg)
}

func (l logger) Discard(msg string) {
	l.logMessage(StatusDiscard, loggerCategoryDiscard, msg)
}

func (l logger) Event(msg string) {
	l.logMessage(StatusOK, loggerCategoryEvents, msg)
}

func (l logger) Info(msg string) {
	l.logMessage(StatusOK, loggerCategoryAll, msg)
}

func (l *logger) setMask(mask loggerCategory) {
	l.mask = mask
}

func (l logger) logMessage(status Status, category loggerCategory, message string) {
	if l.mask&category == 0 {
		return
	}

	l.fmiCallbackLogger(status, category.String(), message)
}

func loggerCategoryFromString(category string) (loggerCategory, error) {
	if category == "" {
		return 0, errors.New("Log category cannot be empty")
	}

	for c, s := range loggerCategories {
		if s == category {
			return c, nil
		}
	}
	return 0, fmt.Errorf("Log category %s is unknown", category)
}
