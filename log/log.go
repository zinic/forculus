package log

import (
	golog "log"
)

type Level int

func (s Level) Accepts(level Level) bool {
	return level <= s
}

const (
	LevelFatal Level = iota
	LevelError
	LevelWarn
	LevelInfo
	LevelDebug
)

type ThresholdList []Level

func (s ThresholdList) Accepts(level Level) bool {
	for _, threshold := range s {
		if threshold.Accepts(level) {
			return true
		}
	}

	return false
}

type Logger interface {
	Threshold() Level

	Fatalf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Debugf(format string, args ...interface{})

	Fatal(args ...interface{})
	Error(args ...interface{})
	Warn(args ...interface{})
	Info(args ...interface{})
	Debug(args ...interface{})
}

type logger struct {
	threshold Level
	output    *golog.Logger
}

func NewLogger(threshold Level, output *golog.Logger) Logger {
	return &logger{
		threshold: threshold,
		output:    output,
	}
}

func (s *logger) Threshold() Level {
	return s.threshold
}

func (s *logger) logf(level Level, format string, args ...interface{}) {
	if s.threshold.Accepts(level) {
		s.output.Printf(format, args...)
	}
}

func (s *logger) Fatalf(format string, args ...interface{}) {
	s.logf(LevelFatal, format, args...)
}

func (s *logger) Errorf(format string, args ...interface{}) {
	s.logf(LevelError, format, args...)
}

func (s *logger) Warnf(format string, args ...interface{}) {
	s.logf(LevelWarn, format, args...)
}

func (s *logger) Infof(format string, args ...interface{}) {
	s.logf(LevelInfo, format, args...)
}

func (s *logger) Debugf(format string, args ...interface{}) {
	s.logf(LevelDebug, format, args...)
}

func (s *logger) log(level Level, args ...interface{}) {
	if s.threshold.Accepts(level) {
		s.output.Print(args...)
	}
}

func (s *logger) Fatal(args ...interface{}) {
	s.log(LevelFatal, args...)
}

func (s *logger) Error(args ...interface{}) {
	s.log(LevelError, args...)
}

func (s *logger) Warn(args ...interface{}) {
	s.log(LevelWarn, args...)
}

func (s *logger) Info(args ...interface{}) {
	s.log(LevelInfo, args...)
}

func (s *logger) Debug(args ...interface{}) {
	s.log(LevelDebug, args...)
}
