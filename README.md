# go-symlinkfarm

Given a set of directories, creates a symlink farm of them in a specified target directory.
This is most useful in instances where there are a lot of packages that you want to merge into
one directory, but without duplicating files.

## Installing

`go get github.com/robxu9/go-symlinkfarm`

## Usage

* `func Create(config *FarmConfig, targetDir string, sourceDirs ...string) error`:
    creates a symlink farm, made up of its target directory, and the source directories to symlink into.

## License

Copyright © 2018 Robert Xu. [MIT Licensed](https://robxu9.mit-license.org/2018)