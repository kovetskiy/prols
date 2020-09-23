package main

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/reconquest/karma-go"
)

type File struct {
	Path    string
	Binary  bool
	Score   int
	ModTime time.Time
	depth   int
}

func (file *File) Depth() int {
	if file.depth == 0 {
		file.depth = strings.Count(file.Path, "/") + 1
	}

	return file.depth
}

func detectType(
	base string,
	path string,
) (string, error) {
	fullpath := filepath.Join(base, path)

	file, err := os.OpenFile(fullpath, os.O_RDONLY, 0644)
	if err != nil {
		return "", karma.Format(
			err,
			"unable to open %s", fullpath,
		)
	}

	defer file.Close()

	buffer := make([]byte, 512)

	size, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", karma.Format(
			err,
			"unable to read file %s", fullpath,
		)
	}

	return http.DetectContentType(buffer[:size]), nil
}
