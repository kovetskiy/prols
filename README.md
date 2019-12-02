# prols

<img src="wow.jpg" height="200px" />

This program is suitable for people who met with following thoughts while
working with list of files in their favorite editor:

- I'm not editing any binary files therefore I don't want to see them
- I'm a %language% developer therefore I want to see files with specific extension on top of the list of files
- I'm also interested in %another-language% files, they should be listed after my
    primary language files

Then there is a big probability that this program will help you too.

# How it works

- walk and find all files in current directory
- read configuration file and read rules
- calculate score of files
- sort list of files by calculated score
- print results

# How to write a rule

A rule contains of following fields (all fields are optional):
- `suffix` - check that filename contains this suffix (extension)
- `prefix` - check that filename contains this prefix (some project oriented things)
- `binary` - check that file is binary
- `score` - score to apply if all conditions are passed

If one of given points of rule are not passed, the rule's score will not be
added to file's score.

Example of list of rules:
```yaml
rules:
    - suffix: .go
      score: 10
    - suffix: .md
      score: 5
    - binary: true
      score: -10

hide_negative: true
```

`hide_negative: true` means that prols will hide all files that has negative
score, binary files will be hidden according to given configuration file.

You can also add `ignore_dirs` to hide some git directories completely, like
.git:

```yaml
ignore_dirs:
    - ".git"
```

Use `use_gitignore: true` if you want prols to use .gitignore for list of
files/directories to ignore (glob patterns work too).

Full configuration file will look like in this file: [prols.conf](prols.conf)
Let's save this file to `~/.config/prols/prols.conf` and run it in this
project:

```bash
$ prols
prols.conf
README.md
config.go
file.go
main.go
rule.go
```

If you want to reverse sort, use `reverse: true` in your configuration file.

# My config

```
ignore_dirs:
    - ".git"
    - "vendor"

use_gitignore: true
hide_negative: true
reverse: true

presort:
    - field: depth

rules:
    - score: 5
    - suffix: .ttf
      score: -10
    - suffix: .xml
      score: -1
    - suffix: .png
      score: -10
    - suffix: .go
      score: 10
    - suffix: .java
      score: 10
    - suffix: .c
      score: 10
    - binary: true
      score: -10
```

# Installation

The program is go-gettable:

```bash
go get github.com/kovetskiy/prols
```

# License
MIT
