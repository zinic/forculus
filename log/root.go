package log

import (
	"os"
	"sync"
)

type manager struct {
	outputs []Logger
	lock    *sync.RWMutex
}

var (
	root *manager
)

func init() {
	root = &manager{
		lock: &sync.RWMutex{},
	}
}

func AddOutput(logger Logger) {
	root.lock.Lock()
	defer root.lock.Unlock()

	root.outputs = append(root.outputs, logger)
}

func Configure() {
	root.lock.Lock()
	defer root.lock.Unlock()

	root.outputs = nil
}

func ConfigureDefaults() {
	AddOutput(NewStdoutLogger(LevelInfo, ""))
}

func Thresholds() ThresholdList {
	root.lock.Lock()
	defer root.lock.Unlock()

	thresholds := make([]Level, len(root.outputs))
	for idx, output := range root.outputs {
		thresholds[idx] = output.Threshold()
	}

	return thresholds
}

func Fatalf(format string, args ...interface{}) {
	root.lock.RLock()
	defer root.lock.RUnlock()

	for _, output := range root.outputs {
		output.Fatalf(format, args...)
	}

	os.Exit(1)
}

func Errorf(format string, args ...interface{}) {
	root.lock.RLock()
	defer root.lock.RUnlock()

	for _, output := range root.outputs {
		output.Errorf(format, args...)
	}
}

func Warnf(format string, args ...interface{}) {
	root.lock.RLock()
	defer root.lock.RUnlock()

	for _, output := range root.outputs {
		output.Warnf(format, args...)
	}
}

func Infof(format string, args ...interface{}) {
	root.lock.RLock()
	defer root.lock.RUnlock()

	for _, output := range root.outputs {
		output.Infof(format, args...)
	}
}

func Debugf(format string, args ...interface{}) {
	root.lock.RLock()
	defer root.lock.RUnlock()

	for _, output := range root.outputs {
		output.Debugf(format, args...)
	}
}

func Fatal(args ...interface{}) {
	root.lock.RLock()
	defer root.lock.RUnlock()

	for _, output := range root.outputs {
		output.Fatal(args...)
	}

	os.Exit(1)
}

func Error(args ...interface{}) {
	root.lock.RLock()
	defer root.lock.RUnlock()

	for _, output := range root.outputs {
		output.Error(args...)
	}
}

func Warn(args ...interface{}) {
	root.lock.RLock()
	defer root.lock.RUnlock()

	for _, output := range root.outputs {
		output.Warn(args...)
	}
}

func Info(args ...interface{}) {
	root.lock.RLock()
	defer root.lock.RUnlock()

	for _, output := range root.outputs {
		output.Info(args...)
	}
}

func Debug(args ...interface{}) {
	root.lock.RLock()
	defer root.lock.RUnlock()

	for _, output := range root.outputs {
		output.Debug(args...)
	}
}
