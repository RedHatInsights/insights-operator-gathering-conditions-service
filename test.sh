#!/usr/bin/env bash
# Copyright 2022 Red Hat, Inc
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

COLORS_RED='\033[0;31m'
COLORS_RESET='\033[0m'
VERBOSE_OUTPUT=true

echo bash version is:
bash --version

if [[ $* == *verbose* ]] || [[ -n "${VERBOSE}" ]]; then
    # print all possible logs
    LOG_LEVEL=""
    VERBOSE_OUTPUT=true
fi

function cleanup() {
    print_descendent_pids() {
        pids=$(pgrep -P "$1")
        echo "$pids"
        for pid in $pids; do
            print_descendent_pids "$pid"
        done
    }

    echo Exiting and killing all children...

    children=$(print_descendent_pids $$)

    # disable the message when you send stop signal to child processes
    set +m

    for pid in $(echo -en "$children"); do
        # nicely asking a process to commit suicide
        if ! kill -PIPE "$pid" &>/dev/null; then
            # we even gave them plenty of time to think
            sleep 1
        fi
    done

    # restore the message back since we want to know that process wasn't stopped correctory
    # set -m

    for pid in $(echo -en "$children"); do
        # murdering those who're alive
        kill -9 "$pid" &>/dev/null
    done

    sleep 1
}
trap cleanup EXIT

go clean -testcache

if go build -race; then
    echo "Service build ok"
else
    echo "Build failed"
    exit 1
fi

function start_service() {
    echo "Starting a service"
    INSIGHTS_OPERATOR_GATHERING_CONDITIONS_SERVICE_CONFIG_FILE=./tests/config \
        ./insights-operator-gathering-conditions-service ||
        echo -e "${COLORS_RED}service exited with error${COLORS_RESET}" &
    # shellcheck disable=2181
    if [ $? -ne 0 ]; then
        echo "Could not start the service"
        exit 1
    fi
}

function test_rest_api() {
    start_service
    sleep 1

    echo "Building REST API tests utility"
    if go build -o rest-api-tests tests/rest_api_tests.go; then
        echo "REST API tests build ok"
    else
        echo "Build failed"
        return 1
    fi

    curl http://localhost:8081/openapi.json > /dev/null || {
        echo -e "${COLORS_RED}server is not running(for some reason)${COLORS_RESET}"
        exit 1
    }

    OUTPUT=$(./rest-api-tests 2>&1)
    EXIT_CODE=$?

    if [ "$VERBOSE_OUTPUT" = true ]; then
        echo "$OUTPUT"
    else
        echo "$OUTPUT" | grep -v -E "^Pass "
    fi

    return $EXIT_CODE
}

echo -e "------------------------------------------------------------------------------------------------"

case $1 in
rest_api)
    test_rest_api
    EXIT_VALUE=$?
    ;;
*)
    # all tests
    # exit value will be 0 if every test returned 0
    EXIT_VALUE=0

    test_rest_api
    EXIT_VALUE=$((EXIT_VALUE + $?))

    ;;
esac

echo -e "------------------------------------------------------------------------------------------------"

exit $EXIT_VALUE
