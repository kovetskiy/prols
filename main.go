package main

import (
	"bytes"
	"io"
	"os"
	"runtime/pprof"
	"sort"
	"strings"

	"github.com/docopt/docopt-go"
	"github.com/kovetskiy/lorg"
	"github.com/monochromegane/go-gitignore"
	"github.com/reconquest/cog"
	"github.com/reconquest/karma-go"
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
  --cpuprofile <path> Enable cpu profiling.
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

	if path, ok := args["--cpuprofile"].(string); ok {
		file, err := os.Create(path)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(file)
		defer pprof.StopCPUProfile()
	}

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
		log.Debug("")
		log.Debug("items with all scores")
		for i := 0; i < len(files); i++ {
			log.Debugf(nil, "%s %d", files[i].Path, files[i].Score)
		}
	}

	if config.HideNegative {
		files = removeNegative(files)
	}

	if config.ScoreDirs {
		files = applyScoreDirs(files)
		// after changing scores we need to re-sort files again
		files = applySortScore(files)
	}

	if config.Reverse {
		for i := len(files)/2 - 1; i >= 0; i-- {
			opp := len(files) - 1 - i
			files[i], files[opp] = files[opp], files[i]
		}
	}

	if debug {
		log.Debug("")
		log.Debug("resulting scores (possibly reversed)")
		for i := 0; i < len(files); i++ {
			log.Debugf(nil, "%s %d", files[i].Path, files[i].Score)
		}
	}

	writeResult(files)
}

func writeResult(files []*File) {
	buffer := bytes.NewBufferString("")
	for _, file := range files {
		buffer.WriteString(file.Path)
		buffer.WriteString("\n")
	}

	// io.Copy works much better than os.Stdout.Write(buffer)
	// results on 180k of files:
	// io.Copy - 0.15s
	// os.Stdout.Write - 0.20s
	io.Copy(os.Stdout, buffer)
}

func removeNegative(files []*File) []*File {
	// files are already sorted so we need to find the latest element with
	// negative score and re-slice files
	last := -1
	for i := 0; i < len(files); i++ {
		if files[i].Score >= 0 {
			break
		}

		last = i
	}

	if last == -1 {
		return files
	}

	return files[last+1:]
}

func walk(config *Config) ([]*File, error) {
	var ignoreMatcher gitignore.IgnoreMatcher

	if config.UseGitignore {
		var err error
		ignoreMatcher, err = gitignore.NewGitIgnore(".gitignore")
		if err != nil && !os.IsNotExist(err) {
			return nil, karma.Format(
				err,
				"unable to read .gitignore",
			)
		}
	}

	shouldDetectType := false
	for _, rule := range config.Rules {
		if rule.Binary != nil {
			shouldDetectType = true
			break
		}
	}

	scanner := &Scanner{
		ignoreMap:        makeMap(config.IgnoreDirs),
		ignoreMatcher:    ignoreMatcher,
		shouldDetectType: shouldDetectType,
	}

	scheduler := &Scheduler{
		maxThreads: config.MaxThreads,
	}

	scheduler.scanner = scanner
	scanner.scheduler = scheduler

	scheduler.Schedule(".")

	scheduler.Wait()

	return scanner.items, nil
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

func applyScoreDirs(files []*File) []*File {
	dirs := map[string]int{}
	getBase := func(path string) string {
		index := strings.Index(path, "/")
		if index == -1 {
			return "."
		}

		return path[:index] + "/"
	}

	write := func(path string, score int) {
		name := getBase(path)

		_, ok := dirs[name]
		if !ok {
			dirs[name] = score
		} else {
			dirs[name] += score
		}
	}

	for _, file := range files {
		write(file.Path, file.Score)
	}

	for _, file := range files {
		file.Score += dirs[getBase(file.Path)]
	}

	return files
}
