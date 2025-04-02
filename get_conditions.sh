#!/bin/bash
set -e

STABLE_VERSION=1.1.3
CANARY_VERSION=1.1.3

# Clone the conditions repo and build it to gather the conditions
if [ ! -d 'insights-operator-gathering-conditions' ]; then git clone https://github.com/RedHatInsights/insights-operator-gathering-conditions; fi
mkdir -p conditions
mkdir -p remote-configurations
cd insights-operator-gathering-conditions

git checkout ${STABLE_VERSION}
./build.sh
cp -r build/v1 ../conditions/stable
cp -r build/v2 ../remote-configurations/stable
cp build/cluster-mapping.json ../remote-configurations/stable/cluster_version_mapping.json
rm -r build

git checkout ${CANARY_VERSION}
./build.sh
cp -r build/v1 ../conditions/canary
cp -r build/v2 ../remote-configurations/canary
cp build/cluster-mapping.json ../remote-configurations/canary/cluster_version_mapping.json
rm -r build
