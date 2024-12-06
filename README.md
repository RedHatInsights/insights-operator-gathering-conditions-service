# Insights Operator Gathering Conditions Service

[![GitHub Pages](https://img.shields.io/badge/%20-GitHub%20Pages-informational)](https://redhatinsights.github.io/insights-operator-gathering-conditions-service/)
[![Go Report Card](https://goreportcard.com/badge/github.com/RedHatInsights/insights-operator-gathering-conditions-service)](https://goreportcard.com/report/github.com/RedHatInsights/insights-operator-gathering-conditions-service)
[![Build Status](https://app.travis-ci.com/RedHatInsights/insights-operator-gathering-conditions-service.svg?branch=main)](https://app.travis-ci.com/RedHatInsights/insights-operator-gathering-conditions-service)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/RedHatInsights/insights-operator-gathering-conditions-service)
[![codecov](https://codecov.io/gh/RedHatInsights/insights-operator-gathering-conditions-service/branch/main/graph/badge.svg?token=PJERQFO02H)](https://codecov.io/gh/RedHatInsights/insights-operator-gathering-conditions-service)
[![License](https://img.shields.io/badge/license-Apache-blue)](https://github.com/RedHatInsights/insights-operator-gathering-conditions-service/blob/main/LICENSE)

Gathering Conditions Services to [Insights Operator](https://github.com/openshift/insights-operator).

<!-- vim-markdown-toc GFM -->

- [Insights Operator Gathering Conditions Service](#insights-operator-gathering-conditions-service)
- [Description](#description)
  - [REST API](#rest-api)
- [Usage](#usage)
  - [Build](#build)
    - [As a container](#as-a-container)
  - [Configure](#configure)
  - [Conditions](#conditions)
  - [Run](#run)
    - [Flags](#flags)
    - [Rapid recommendations](#rapid-recommendations)
  - [Monitoring](#monitoring)
  - [Makefile](#makefile)
  - [BDD tests](#bdd-tests)
  - [Definition of Done for new features and fixes](#definition-of-done-for-new-features-and-fixes)
- [License](#license)

<!-- vim-markdown-toc -->

# Description

Simple service that provides conditional gathering-related rules. Such rules
are read from JSON structures and made available via REST API.

## REST API

REST API is described by [OpenAPI specification](openapi.json).

# Usage

## Build

To build the service locally, install Go 1.20 or above and run:

```shell script
make build
```

### As a container

To build the container you need to set up two environment variables:

1. `CONTAINER_RUNTIME` usually `docker` or `podman`
2. `CONTAINER_IMAGE_NAME` the image name


then use the command:

```shell script
make container-build
```

Once build you can run it using:

```shell script
make container-run
```

Then you can test it:

```shell script
curl -s http://localhost:8081/api/gathering/gathering_rules | jq
```

The image contains stable and canary versions of configurations. Exact values of these versions are specified
in `get_conditions.sh` script and should be changed each time we want release a new conditions version.
Unfortunatelly it is not possible to change these values in the app-interface, because the change requires image rebuild.

## Canary rollout

Service provides option of canary rollout for new version of configurations. The image contains two different versions of configurations (`stable` and `canary`) and serves one of them depending on cluster ID retrieved from request header. The ratio of cluster IDs being served `stable` version opposed to `canary` is defined in the Unleash instance - `insights.unleash.devshift.net` for production environment and `insights-stage.unleash.devshift.net` for stage environment. We do not have Unleash instance that could be used in ephemeral.

## Configure

Configuration is done by `toml` config, taking the `config.toml` in the working directory if no other configuration is provided. This can be overriden by `INSIGHTS_OPERATOR_CONDITIONAL_SERVICE_CONFIG_FILE` environment variable.

You can also override specific configuration options by setting specific
environment variables in this form:
`INSIGHTS_OPERATOR_GATHERING_CONDITIONS_SERVICE__{table}__{key}`,
being the `table` and `key ` concepts defined in the [TOML specification](https://toml.io/en/v1.0.0#table).

For example, for a configuration like this

```
[storage]
remote_configuration = "./conditions"
cluster_mapping = "./mapping"
```

you would override these values as follows:

```
export INSIGHTS_OPERATOR_GATHERING_CONDITIONS_SERVICE__STORAGE__REMOTE_CONFIGURATION=tests/rapid-recommendations
export INSIGHTS_OPERATOR_GATHERING_CONDITIONS_SERVICE__STORAGE__CLUSTER_MAPPING=tests/mapping/
```

`remote_configuration`, `conditions` and `cluster_mapping` configuration options describe locations of respective content served by the service. However, each of these also contain separate `stable` and `canary` subdirectories containing different versions of the content. `cluster_mapping_file` option then describes the name of the file within `${cluster_mapping}/stable` and`${cluster_mapping}/canary`, and this file maps different OCP versions to the specific content under `${remote_configuration}/stable` (or `${remote_configuration}/canary`).

## Conditions

This service exposes the conditions from the
[insights-operator-gathering-conditions](https://github.com/RedHatInsights/insights-operator-gathering-conditions)
repository as a REST API. First you need to clone that repository and build it using

```shell script
git clone https://github.com/RedHatInsights/insights-operator-gathering-conditions
cd insights-operator-gathering-conditions
./build.sh
cp -r ./build ../conditions
```

or `make conditions`. This will create a set of folders under `conditions` with
the v1 remote configurations and the v2 rapid recommendations (also called
remote-configurations).

## Run

To execute the service, you have 3 alternatives:

- `./insights-operator-gathering-conditions-service`
- `make run`
- `make container-run` (as explained above)

Then you can test it using `curl`, for example:

```shell script
curl -s http://localhost:8000/api/gathering/v1/gathering_rules | jq
```

### Flags

There are some flags for different purposes:

- `./insights-conditions-service -show-configuration`: used to print the configuration in `stdout`.
- `./insights-conditions-service -show-authors`: used to print the authors of the repository.
- `./insights-conditions-service -show-version`: used to print the binary version including commit, branch and build time.

### Rapid recommendations

As part of [CCXDEV-12849](https://issues.redhat.com/browse/CCXDEV-12849) we
introduced a new feature to map remote configurations to different OCP versions.

In order to use it you need to set the `cluster_mapping`, `cluster_mapping_file` and
`remote_configuration` fields in [config.toml] or use environment variables.
The cluster map should look like this:
```
[
	["1.0.0", "first.json"],
	["2.0.0", "second.json"],
	["3.0.0", "third.json"]
]
```
meaning clusters with versions between 1.0.0 and 2.0.0 would receive first.json
, second.json for versions between 2.0.0 and 3.0.0 and third.json for versions
greater than 3.0.0.

Note that when major, minor, and patch are equal, a pre-release version has
lower precedence than a normal version. For example,
`4.17.6-0.ci-2024-08-19-220527` is lower than `4.17.6`.

Use `curl -s http://localhost:8000/api/gathering/v2/4.17.0/gathering_rules` in
order to check this new endpoint.

## Monitoring

The service exposes some metrics in the `/metrics` endpoint. Apart from the
default metrics, it also exposes the ones defined in
[metrics.go](internal/service/metrics.go).

All these metrics are then used in [Grafana](https://grafana.app-sre.devshift.net/d/gathering/ccx-gathering-service)

## Makefile

There are many options inside the [Makefile](Makefile) that may be useful for debugging/deploying the service:

```
Usage: make <OPTIONS> ... <TARGETS>

Available targets are:

clean                Run go clean
build                Build binary containing service executable
fmt                  Run go fmt -w for all sources
lint                 Run golint
vet                  Run go vet. Report likely mistakes in source code
cyclo                Run gocyclo
ineffassign          Run ineffassign checker
shellcheck           Run shellcheck
errcheck             Run errcheck
goconst              Run goconst checker
gosec                Run gosec checker
abcgo                Run ABC metrics checker
openapi-check        Validate the OpenAPI specification files
conditions           Clone the conditions repo and build it to gather the conditions
check-config         Check all the configuration files are parsable
style                Run all the formatting related commands (fmt, vet, lint, cyclo) + check shell scripts
run                  Build the project and executes the binary
test                 Run the unit tests
integration_tests    Run all integration tests
before_commit        Checks done before commit
help                 Show this help screen
function_list        List all functions in generated binary file
container-build      Build the container image
container-run        Run the container image
```

## BDD tests

Behaviour tests for this service are included in [Insights Behavioral
Spec](https://github.com/RedHatInsights/insights-behavioral-spec) repository.
In order to run these tests, the following steps need to be made:

1. clone the [Insights Behavioral Spec](https://github.com/RedHatInsights/insights-behavioral-spec) repository
1. go into the cloned subdirectory `insights-behavioral-spec`
1. run the `aggregator_tests.sh` from this subdirectory

List of all test scenarios prepared for this service is available at
<https://redhatinsights.github.io/insights-behavioral-spec/feature_list.html#insights-operator-gathering-conditions-service>

## Definition of Done for new features and fixes

Please look at [DoD.md](DoD.md) document for definition of done for new features and fixes.

# License

This project is licensed by the Apache License 2.0. For more information check the LICENSE file.
