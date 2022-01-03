# Insights Operator Gathering Conditions Service

Gathering Conditions Services to [Insights Operator](https://github.com/openshift/insights-operator).

<!-- vim-markdown-toc GFM -->

- [Insights Operator Gathering Conditions Service](#insights-operator-gathering-conditions-service)
- [Build](#build)
  - [Configure](#configure)
  - [Conditions](#conditions)
  - [Run](#run)
- [Container](#container)
- [License](#license)

<!-- vim-markdown-toc -->

# Build

To build the service, install Go 1.14 or above and run:

```shell script
make build
```

## Configure

Configuration is done by `toml` config, taking the `config/config.toml` in the working directory if no configuration is provided. This can be overriden by `INSIGHTS_OPERATOR_CONDITIONAL_SERVICE_CONFIG_FILE` environment variable.

## Conditions

First you need to clone the conditions repository and build it

```shell script
git clone https://github.com/redhatinsights/insights-operator-gathering-conditions
cd insights-operator-gathering-conditions
./build.sh
```

It will build the gathering conditions image.

## Run

To execute the service, run:

```shell script
bin/insights-conditions-service
```

There are some flags for different purposes:

- `bin/insights-conditions-service -show-configuration`: used to print the configuration in `stdout`.

# Container

To build the container use the command:

```shell script
make container-build
```

Once build you can run it using:

```shell script
make container-run
```

Then you can test it:

```shell script
curl -s http://localhost:8081/gathering_rules | jq
```

# License

This project is licensed by the Apache License 2.0. For more information check the LICENSE file.
