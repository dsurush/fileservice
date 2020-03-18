package file

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
	"mime"
	"os"
	"path/filepath"
)

type Service struct {
	Filepath string
}

type MyFile interface {
	Save()
}

func NewService(dir string) *Service {
	return &Service{Filepath: dir}
}

func (s *Service) Save(src io.Reader, contentType string) (name string, err error) {
	extensions, err := mime.ExtensionsByType(contentType)
	if err != nil {
		return "", err
	}

	if len(extensions) == 0 {
		return "", errors.New("invalid extension")
	}

	uuidV4 := uuid.New().String()
	name = fmt.Sprintf("%s%s", uuidV4, extensions[len(extensions)-1])
	path := filepath.Join(string(s.Filepath), name)

	dst, _ := os.Create(path)
	defer dst.Close()
	_, err = io.Copy(dst, src)
	if err != nil {
		return "", err
	}
	return name, nil
}