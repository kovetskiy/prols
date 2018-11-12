# prols

This program is suitable for people who has such thoughts while working with
list of files in their favorite editor:

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

Full configuration file will look like:
```yaml
ignore_dirs:
    - ".git"

hide_negative: true

rules:
    - suffix: .go
      score: 10
    - suffix: .md
      score: 5
    - binary: true
      score: -10

# vim: ft=yaml
```

Let's save this file to `~/.config/prols/prols.conf` and run it in this
project:

```
$ prols
prols.conf
README.md
config.go
file.go
main.go
rule.go
```

If you want to reverse sort, you can run program like `prols | tac`.

As you can see, files are sorted as it's expected.

# License
MIT
