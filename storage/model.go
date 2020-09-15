package storage

import (
	"io"

	"github.com/zinic/forculus/config"
)

type Provider interface {
	Configure(cfg config.StorageProvider) error
	Validate(cfg config.StorageProvider) error
	Write(key string, reader io.Reader) error
	Read(key string) (io.ReadCloser, error)
	Stat(key string) (Details, error)
}

type Details struct {
	Size int64
}
