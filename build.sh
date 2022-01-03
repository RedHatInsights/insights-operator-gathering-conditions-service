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

# this is improper - we need to start using tags in GitHub properly
version=0.1

buildtime=$(date)
branch=$(git rev-parse --abbrev-ref HEAD)
commit=$(git rev-parse HEAD)

package_prefix=github.com/redhatinsights/insights-operator-conditional-gathering/internal

go build  -o $1 -ldflags="-X '$package_prefix/cli.BuildTime=$buildtime' -X '$package_prefix/cli.BuildVersion=$version' -X '$package_prefix/cli.BuildBranch=$branch' -X '$package_prefix/cli.BuildCommit=$commit'" $2
exit $?