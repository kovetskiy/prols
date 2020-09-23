package main

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/monochromegane/go-gitignore"
)

type Scanner struct {
	scheduler        *Scheduler
	ignoreMap        map[string]struct{}
	ignoreMatcher    gitignore.IgnoreMatcher
	shouldDetectType bool
	mutex            sync.Mutex
	items            []*File
}

func makeMap(slice []string) map[string]struct{} {
	hash := map[string]struct{}{}
	for _, item := range slice {
		hash[item] = struct{}{}
	}

	return hash
}

func (scanner *Scanner) append(file *File) {
	scanner.mutex.Lock()
	scanner.items = append(scanner.items, file)
	scanner.mutex.Unlock()
}

func (scanner *Scanner) Scan(dir string) {
	infos, err := readdir(dir)
	if err != nil {
		log.Errorf(err, dir)
		return
	}

	var path string
	for _, info := range infos {
		if dir == "." {
			path = info.Name()
		} else {
			path = dir + "/" + info.Name()
		}

		if scanner.ignoreMatcher != nil {
			if scanner.ignoreMatcher.Match(path, info.IsDir()) {
				if info.IsDir() {
					continue
				}

				continue
			}
		}

		if info.IsDir() {
			if _, ok := scanner.ignoreMap[info.Name()]; ok {
				continue
			}

			scanner.scheduler.Schedule(path)
			continue
		}

		if !info.Mode().IsRegular() {
			continue
		}

		file := &File{
			Path:    path,
			ModTime: info.ModTime(),
		}

		if scanner.shouldDetectType {
			contentType, err := detectType(".", path)
			if err != nil {
				log.Errorf(err, path)
			}

			if contentType == "application/octet-stream" {
				file.Binary = true
			}
		}

		scanner.append(file)
	}
}

func readdir(path string) ([]os.FileInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	infos, err := file.Readdir(-1)
	if err != nil {
		return nil, err
	}

	result := []os.FileInfo{}
	for _, info := range infos {
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			_, err := filepath.EvalSymlinks(
				filepath.Join(path, info.Name()),
			)
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return nil, err
			}
		}

		result = append(result, info)
	}

	return result, nil
}
