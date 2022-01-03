# go-capture

[![Go Report Card](https://goreportcard.com/badge/github.com/tisnik/go-capture)](https://goreportcard.com/report/github.com/tisnik/go-capture)
[![Build Status](https://travis-ci.org/tisnik/go-capture.svg?branch=master)](https://travis-ci.org/tisnik/go-capture)
[![codecov](https://codecov.io/gh/tisnik/go-capture/branch/master/graph/badge.svg)](https://codecov.io/gh/tisnik/go-capture)

## Description

Utility functions to capture standard output, error output, or log output

## Testing

Unit tests can be started by the following command:

```
./test.sh
```

It is also possible to specify CLI options for Go test. For example, if you need to disable test results caching, use the following command:

```
./test -count=1
```

## CI

[Travis CI](https://travis-ci.com/) is configured for this repository. Several tests and checks are started for all pull requests:

* Unit tests that use the standard tool `go test`
* `go fmt` tool to check code formatting. That tool is run with `-s` flag to perform [following transformations](https://golang.org/cmd/gofmt/#hdr-The_simplify_command)
* `go vet` to report likely mistakes in source code, for example suspicious constructs, such as Printf calls whose arguments do not align with the format string.
* `golint` as a linter for all Go sources stored in this repository
* `gocyclo` to report all functions and methods with too high cyclomatic complexity. The cyclomatic complexity of a function is calculated according to the following rules: 1 is the base complexity of a function +1 for each 'if', 'for', 'case', '&&' or '||' Go Report Card warns on functions with cyclomatic complexity > 9

History of checks done by CI is available at [RedHatInsights / insights-operator-controller](https://travis-ci.org/RedHatInsights/insights-operator-controller).

## Contribution

Please look into document [CONTRIBUTING.md](CONTRIBUTING.md) that contains all information about how to contribute to this project.
