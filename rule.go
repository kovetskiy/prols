package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-yaml/yaml"
	"github.com/reconquest/karma-go"
)

type Rule struct {
	Suffix          string `yaml:"suffix,omitempty"`
	Prefix          string `yaml:"prefix,omitempty"`
	Depth           string `yaml:"depth,omitempty"`
	depthValue      int
	depthComparison byte
	Binary          *bool `yaml:"binary,omitempty"`
	Score           int   `yaml:"score" required:"true"`
}

func ParseRule(raw string) (Rule, error) {
	var rule Rule

	lines := []string{}
	for _, line := range strings.Split(raw, ",") {
		chunks := strings.SplitN(line, ":", 2)
		if len(chunks) < 2 {
			return rule, fmt.Errorf(
				"invalid part of rule: expected key:value, but got: %s",
				line,
			)
		}

		lines = append(
			lines,
			strings.TrimSpace(chunks[0])+": "+strings.TrimSpace(chunks[1]),
		)
	}

	err := yaml.Unmarshal([]byte(strings.Join(lines, "\n")), &rule)
	if err != nil {
		return rule, err
	}

	return rule, rule.init()
}

func (rule Rule) String() string {
	data, err := yaml.Marshal(rule)
	if err != nil {
		panic(err)
	}

	contents := string(data)
	contents = strings.TrimSpace(contents)

	return "[" + strings.Replace(contents, "\n", "; ", -1) + "]"
}

func (rule *Rule) init() error {
	var err error

	if rule.Depth != "" {
		var value int

		sign := rule.Depth[0]
		if sign == '<' || sign == '>' {
			if len(rule.Depth) == 1 {
				return errors.New("invalid depth value")
			}

			value, err = strconv.Atoi(rule.Depth[1:])
			if err != nil {
				return karma.Format(
					err,
					"invalid depth value",
				)
			}
		} else {
			value, err = strconv.Atoi(rule.Depth)
			if err != nil {
				return karma.Format(
					err,
					"invalid depth value",
				)
			}
		}

		rule.depthValue = value
		switch sign {
		case '>', '<':
			rule.depthComparison = sign
		}
	}

	return nil
}

func (rule *Rule) Pass(file *File) bool {
	if rule.Prefix != "" {
		if !strings.HasPrefix(file.Path, rule.Prefix) {
			return false
		}
	}

	if rule.Suffix != "" {
		if !strings.HasSuffix(file.Path, rule.Suffix) {
			return false
		}
	}

	if rule.depthValue != 0 {
		depth := file.Depth()

		switch rule.depthComparison {
		case '<':
			if depth >= rule.depthValue {
				return false
			}
		case '>':
			if depth <= rule.depthValue {
				return false
			}
		default:
			if depth != rule.depthValue {
				return false
			}
		}
	}

	if rule.Binary != nil {
		if *rule.Binary != file.Binary {
			return false
		}
	}

	return true
}
