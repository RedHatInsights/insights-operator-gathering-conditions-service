#!/bin/bash -x
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

# retrieve the latest tag set in repository
version=$(git describe --always --tags --abbrev=0)

buildtime=$(date)
branch=$(git rev-parse --abbrev-ref HEAD)
commit=$(git rev-parse HEAD)

package_prefix=github.com/RedHatInsights/insights-operator-gathering-conditions-service/internal

go build -ldflags="-X '$package_prefix/cli.BuildTime=$buildtime' -X '$package_prefix/cli.BuildVersion=$version' -X '$package_prefix/cli.BuildBranch=$branch' -X '$package_prefix/cli.BuildCommit=$commit'"
exit $?
