package storage

import (
	"io"
)

type Storager interface {
	List(path string) ([]*File, error)
	Delete(path string) error
	Upload(file io.Reader, destFilename, contentType string) (*File, error)
	Close() error
}

type File struct {
	Path       string `json:"path"`
	PublicURL  string `json:"public_url"`
	StorageURL string `json:"storage_url"`
}

type StoragerMock struct {
	Storager
}

func (s *StoragerMock) List(path string) ([]*File, error) {
	return nil, nil
}
func (s *StoragerMock) Delete(path string) error {
	return nil
}
func (s *StoragerMock) Upload(file io.Reader, destFilename, contentType string) (*File, error) {
	return nil, nil
}
func (a *StoragerMock) Close() error {
	return nil
}
