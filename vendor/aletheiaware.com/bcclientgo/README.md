bcclientgo
==========

This is a Go implementation of a BC client using the BC data structures.

# About

`bcclientgo` provides utilities and a command line interface to explore and interact with BC.

# Build

    go build -tags release

# Install

Install the binary (or download from https://github.com/AletheiaWareLLC/bcclientgo/releases/latest)


    go install -tags release aletheiaware.com/bcclientgo/cmd/bc

# Usage

```
$ bc
BC Usage:
    bc - display usage
    bc init - initializes environment, generates key pair, and registers alias

    bc node - display registered alias and public key
    bc alias [alias] - display public key for given alias

    bc keys - display all stored keys
    bc import-keys [alias] [access-code] - imports the alias and keypair from BC server
    bc export-keys [alias] - generates a new access code and exports the alias and keypair to BC server

    bc push [channel] - pushes the channel to peers
    bc pull [channel] - pulls the channel from peers
    bc head [channel] - display head of given channel
    bc chain [channel] - display chain of given channel
    bc block [channel] [block-hash] - display block with given hash
    bc record [channel] [record-hash] - display record with given hash

    bc read [channel] [block-hash] [record-hash]- reads entries the given channel and writes to stdout
    bc read-key [channel] [block-hash] [record-hash]- reads keys the given channel and writes to stdout
    bc read-payload [channel] [block-hash] [record-hash]- reads payloads the given channel and writes to stdout
    bc write [channel] [access...] - reads data from stdin and writes it to cache for the given channel and grants access to the given aliases
    bc mine [channel] [threshold] - mines the given channel to the given threshold

    bc peers - display list of peers
    bc add-peer [peer] - adds the given peer to the list of peers
    bc keystore - display location of keystore
    bc cache - display location of cache
    bc purge - deletes contents of cache

    bc random - generate a random number
```
