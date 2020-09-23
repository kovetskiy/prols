package main

import (
	"runtime"

	"github.com/go-yaml/yaml"
	"github.com/kovetskiy/ko"
	"github.com/reconquest/karma-go"
)

type Config struct {
	MaxThreads   int      `yaml:"max_threads" default:"0"`
	IgnoreDirs   []string `yaml:"ignore_dirs"`
	UseGitignore bool     `yaml:"use_gitignore"`
	HideNegative bool     `yaml:"hide_negative"`
	Rules        []Rule
	Reverse      bool `yaml:"reverse"`

	PreSort   []PreSort `yaml:"presort"`
	ScoreDirs bool      `yaml:"score_dirs"`
}

type PreSort struct {
	Field   string
	modTime bool
	depth   bool
	path    bool
	Reverse bool
}

func LoadConfig(path string) (*Config, error) {
	var config Config
	err := ko.Load(path, &config, yaml.Unmarshal)
	if err != nil {
		return nil, err
	}

	if config.MaxThreads == 0 {
		config.MaxThreads = runtime.NumCPU()
	}

	for i, rule := range config.Rules {
		err := rule.init()
		if err != nil {
			return nil, karma.Format(
				err,
				"invalid config rule #%v", i+1,
			)
		}
	}

	for i, presort := range config.PreSort {
		switch presort.Field {
		case "depth":
			presort.depth = true
		case "path":
			presort.path = true
		case "mod_time":
			presort.modTime = true

		default:
			return nil, karma.Format(
				err,
				"invalid config presort #v", i+1,
			)
		}

		config.PreSort[i] = presort
	}

	return &config, nil
}
