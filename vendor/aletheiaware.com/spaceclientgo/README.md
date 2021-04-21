spaceclientgo
=============

This is a Go implementation of a S P A C E client - secure, private, storage.

# About

`spaceclientgo` provides utilities and a command line interface to explore and interact with S P A C E.

# Build

    go build -tags release

# Install

Install the binary (or download from https://github.com/AletheiaWareLLC/spaceclientgo/releases/latest)

    go install -tags release aletheiaware.com/spaceclientgo/cmd/space

# Usage

```
$ space
Space Usage:
    space - display usage
    space init - initializes environment, generates key pair, and registers alias

    space add [name] [type] - read stdin and mine a new record into blockchain
    space add [name] [type] [file] - read file and mine a new record into blockchain

    space list - prints all files created by this key
    space list [type] - display metadata of all files with given MIME type
    space show [hash] - display metadata of file with given hash
    space get [hash] - write file with given hash to stdout
    space get [hash] [file] - write file with given hash to file
    space get-all [directory] - write all files to given directory

    space tag [hash] [tag...] - tags file with given hash with given tags
    space search [tag...] - search files for given tags

    space registration [merchant] - display registration information between this alias and the given merchant
    space subscription [merchant] - display subscription information between this alias and the given merchant

    space registrars - display registration and subscription information of this alias' registrars
```
