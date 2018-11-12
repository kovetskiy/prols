package main

import (
	"github.com/go-yaml/yaml"
	"github.com/kovetskiy/ko"
	"github.com/reconquest/karma-go"
)

type Config struct {
	IgnoreDirs []string `yaml:"ignore_dirs" required:"true"`
	HideNegative bool `yaml:"hide_negative"`
	Rules      []Rule
}

func LoadConfig(path string) (*Config, error) {
	var config Config
	err := ko.Load(path, &config, yaml.Unmarshal)
	if err != nil {
		return nil, err
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

	return &config, nil
}
