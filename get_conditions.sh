#!/bin/bash

STABLE_VERSION=1.1.0
CANARY_VERSION=1.1.0

# Clone the conditions repo and build it to gather the conditions
if [ ! -d 'insights-operator-gathering-conditions' ]; then git clone https://github.com/RedHatInsights/insights-operator-gathering-conditions; fi
mkdir -p conditions
mkdir -p remote-configurations
cd insights-operator-gathering-conditions || exit 1

git checkout ${STABLE_VERSION} && \
./build.sh && \
cp -r build/v1 ../conditions/stable && \
cp -r build/v2 ../remote-configurations/stable && \
rm -r build ; \

git checkout ${CANARY_VERSION} && \
./build.sh && \
cp -r build/v1 ../conditions/canary && \
cp -r build/v2 ../remote-configurations/canary && \
rm -r build ; \
