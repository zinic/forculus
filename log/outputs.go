package log

import (
	golog "log"
	"os"
)

func NewStdoutLogger(threshold Level, prefix string) Logger {
	return &logger{
		threshold: threshold,
		output:    golog.New(os.Stdout, prefix, 0),
	}
}
