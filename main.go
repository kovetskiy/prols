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

	files, err := walk(config)
	if err != nil {
		log.Fatalf(err, "unable to walk directory")
	}

	files = applyPreSort(files, config.PreSort)
	files = applyRules(files, config.Rules)
	files = applySortScore(files)

	if debug {
		for _, file := range files {
			log.Debugf(nil, "%s %d", file.Path, file.Score)
		}
	}

	if config.Reverse {
		for i := len(files)/2 - 1; i >= 0; i-- {
			opp := len(files) - 1 - i
			files[i], files[opp] = files[opp], files[i]
		}
	}

	for _, file := range files {
		if config.HideNegative && file.Score < 0 {
			continue
		}

		fmt.Println(file.Path)
	}
}

func walk(config *Config) ([]*File, error) {
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

	err := filepath.Walk(".", walk)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func applyPreSort(files []*File, presorts []PreSort) []*File {
	sort.SliceStable(files, func(i, j int) bool {
		for _, presort := range presorts {
			switch {
			case presort.depth:
				if presort.Reverse {
					if files[i].Depth() < files[j].Depth() {
						return true
					}
				} else {
					if files[i].Depth() > files[j].Depth() {
						return true
					}
				}

			case presort.path:
				if presort.Reverse {
					if files[i].Path > files[j].Path {
						return true
					}
				} else {
					if files[i].Path < files[j].Path {
						return true
					}
				}
			default:
				panic("unexpected presort field: " + presort.Field)
			}
		}

		return false
	})

	return files
}

func applySortScore(files []*File) []*File {
	sort.SliceStable(files, func(i, j int) bool {
		return files[i].Score < files[j].Score
	})
	return files
}

func applyRules(files []*File, rules []Rule) []*File {
	for _, file := range files {
		for _, rule := range rules {
			if rule.Pass(file) {
				if debug {
					log.Debugf(nil, "%s passed %s", file.Path, rule)
				}

				file.Score += rule.Score
			}
		}
	}

	return files
}
