package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/docopt/docopt-go"
	"github.com/kovetskiy/lorg"
	"github.com/reconquest/cog"
)

var (
	version = "[manual build]"
	usage   = "prols " + version + os.ExpandEnv(`

Flexible project-wide search tool based on rules and scores.

Usage:
  prols [options]
  prols -h | --help
  prols --version

Options:
  -c --global <path>  Use specified global prols file.
                       [default: $HOME/.config/prols/prols.conf]
  --debug             Print debug messages.
  -h --help           Show this screen.
  --version           Show version.
`,
	)
)

var (
	log   *cog.Logger
	debug bool
)

func initLogger(args map[string]interface{}) {
	stderr := lorg.NewLog()
	stderr.SetIndentLines(true)
	stderr.SetFormat(
		lorg.NewFormat("${time} ${level:[%s]:right:short} ${prefix}%s"),
	)

	debug = args["--debug"].(bool)

	if debug {
		stderr.SetLevel(lorg.LevelDebug)
	}

	log = cog.NewLogger(stderr)
}

func main() {
	args, err := docopt.Parse(usage, nil, true, version, false)
	if err != nil {
		panic(err)
	}

	initLogger(args)

	globalPath := args["--global"].(string)

	config, err := LoadConfig(globalPath)
	if err != nil {
		log.Fatalf(
			err,
			"unable to load configuration file: %s", globalPath,
		)
	}

	_ = config

	ignoreDirs := map[string]struct{}{}
	for _, path := range config.IgnoreDirs {
		ignoreDirs[path] = struct{}{}
	}

	shouldDetectType := false
	for _, rule := range config.Rules {
		if rule.Binary != nil {
			shouldDetectType = true
			break
		}
	}

	files := []*File{}
	walk := func(path string, info os.FileInfo, err error) error {
		if path == "." {
			return nil
		}

		if info.IsDir() {
			if _, ok := ignoreDirs[path]; ok {
				return filepath.SkipDir
			}

			return nil
		}

		file := &File{
			Path: path,
		}

		if shouldDetectType {
			contentType, err := detectType(".", path)
			if err != nil {
				return err
			}

			if contentType == "application/octet-stream" {
				file.Binary = true
			}
		}

		files = append(files, file)

		return nil
	}

	err = filepath.Walk(".", walk)
	if err != nil {
		log.Fatalf(err, "unable to walk")
	}

	for _, file := range files {
		for _, rule := range config.Rules {
			if rule.Pass(file) {
				if debug {
					log.Debugf(nil, "%s passed %s", file.Path, rule)
				}

				file.Score += rule.Score
			}
		}
	}

	sort.SliceStable(files, func(i, j int) bool {
		return files[i].Score < files[j].Score
	})

	if debug {
		for _, file := range files {
			log.Debugf(nil, "%s %d", file.Path, file.Score)
		}
	}

	for _, file := range files {
		if config.HideNegative && file.Score < 0 {
			continue
		}

		fmt.Println(file.Path)
	}
}
