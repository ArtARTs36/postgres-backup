package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

type FileSystemStorage struct {
	root string
}

func NewFileSystemStorage(root string) (*FileSystemStorage, error) {
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, err
	}
	return &FileSystemStorage{root: root}, nil
}

func (fs *FileSystemStorage) Put(key string, data io.Reader) error {
	path := filepath.Join(fs.root, key)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, data)
	return err
}

func (fs *FileSystemStorage) List(prefix string) ([]FileInfo, error) {
	dir := filepath.Join(fs.root, prefix)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []FileInfo{}, nil
		}
		return nil, err
	}

	var objects []FileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, _ := entry.Info()
		objects = append(objects, FileInfo{
			Key:  filepath.Join(prefix, entry.Name()),
			Size: info.Size(),
		})
	}
	return objects, nil
}

func (fs *FileSystemStorage) Delete(_ context.Context, key string) error {
	return os.Remove(filepath.Join(fs.root, key))
}

func (fs *FileSystemStorage) Close() error {
	return nil
}
