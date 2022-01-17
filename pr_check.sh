#!/bin/bash

# --------------------------------------------
# Options that must be configured by app owner
# --------------------------------------------
APP_NAME="io-gathering-service"  # name of app-sre "application" folder this component lives in
COMPONENT_NAME="io-gathering-service"  # name of app-sre "resourceTemplate" in deploy.yaml for this component
IMAGE="quay.io/cloudservices/io-gathering-conditions-service"

IQE_PLUGINS="ccx"
IQE_MARKER_EXPRESSION="smoke"
IQE_FILTER_EXPRESSION=""

# Temporary stub
mkdir artifacts
echo '<?xml version="1.0" encoding="utf-8"?><testsuites><testsuite name="pytest" errors="0" failures="0" skipped="0" tests="1" time="0.014" timestamp="2021-05-13T07:54:11.934144" hostname="thinkpad-t480s"><testcase classname="test" name="test_stub" time="0.000" /></testsuite></testsuites>' > artifacts/junit-stub.xml
