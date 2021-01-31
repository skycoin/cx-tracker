# cx-tracker
CX chain spec and node tracker.

## Install

### Dependencies

CX Chain requires [Golang](https://golang.org/) to compile (version `1.14+`). Detailed installation instructions can be found here: https://github.com/SkycoinProject/skycoin/blob/develop/INSTALLATION.md

### Build

To build `cx-tracker`, the typical Golang binary build process applies. The following command builds `cx-tracker` into the target directory specified by the `GOBIN` env.

```bash
$ git clone git@github.com:skycoin/cx-tracker.git && cd cx-tracker
$ go install ./cmd/...
```

The `go install` command is also available as a `Makefile` target.

```bash
$ make install
```

## Run

The `cx-tracker` binary has two flag options.

```bash
$ cx-tracker -h

# Usage of cx-tracker:
#  -addr ADDRESS
#        HTTP ADDRESS to serve on (default ":9091")
#  -db FILEPATH
#        database FILEPATH (default "./cx_tracker.db")
```
