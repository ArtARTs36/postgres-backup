package storage

import (
	"context"
	"io"
)

type FileInfo struct {
	Key  string
	Size int64
}

type Storage interface {
	Put(key string, data io.Reader) error
	List(dir string) ([]FileInfo, error)
	Delete(ctx context.Context, key string) error
	Close() error
}
