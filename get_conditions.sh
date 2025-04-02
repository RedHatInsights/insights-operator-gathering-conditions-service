#!/bin/bash
set -e

STABLE_VERSION=1.1.3
CANARY_VERSION=1.1.3


build() {

  pip3.11 install -U -r requirements.txt
  python3.11 build.py

  cp -r build/v1 "../conditions/$1"
  cp -r build/v2 "../remote-configurations/$1"

  rm -r build
}

build_legacy() {
  ./build.sh
  cp -r build/v1 "../conditions/$1"
  cp -r build/v2 "../remote-configurations/$1"
  cp build/cluster-mapping.json "../remote-configurations/$1/cluster_version_mapping.json"
  rm -r build
}

# Clone the conditions repo and build it to gather the conditions
if [ ! -d 'insights-operator-gathering-conditions' ]; then git clone https://github.com/RedHatInsights/insights-operator-gathering-conditions; fi
mkdir -p conditions
mkdir -p remote-configurations
cd insights-operator-gathering-conditions
git fetch

python3.11 -m venv .env
source .env/bin/activate

git checkout ${STABLE_VERSION}
if [ -f build.py ]; then
  build "stable"
else
  build_legacy "stable"
fi

git checkout ${CANARY_VERSION}
if [ -f build.py ]; then
  build "canary"
else
  build_legacy "canary"
fi

rm -r .env
