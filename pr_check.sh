#!/bin/bash

set -exv

# --------------------------------------------
# Options that must be configured by app owner
# --------------------------------------------
APP_NAME="ccx-data-pipeline"  # name of app-sre "application" folder this component lives in
COMPONENT_NAME="io-gathering-service"  # name of app-sre "resourceTemplate" in deploy.yaml for this component
IMAGE="quay.io/cloudservices/io-gathering-conditions-service"
COMPONENTS="io-gathering-service"  # space-separated list of components to laod
COMPONENTS_W_RESOURCES="io-gathering-service"  # component to keep

export IQE_PLUGINS="ccx"
export IQE_MARKER_EXPRESSION="io_gathering"
export IQE_FILTER_EXPRESSION=""
export IQE_CJI_TIMEOUT="60m"

# Install bonfire repo/initialize
CICD_URL=https://raw.githubusercontent.com/RedHatInsights/bonfire/master/cicd
curl -s $CICD_URL/bootstrap.sh > .cicd_bootstrap.sh && source .cicd_bootstrap.sh

# Workaround to avoid issue with long name of namespace 'requestor'
# Jenkins job name is overriden. Must be exported after 'bootstrap.sh' is sourced.
# Issue: https://github.com/RedHatInsights/bonfire/issues/199
# FIXME: Remove this line when it is fixed.
export BONFIRE_NS_REQUESTER="io-gather-conds-serv-pr-check-${BUILD_NUMBER}"


# Build the image and push to quay
source $CICD_ROOT/build.sh

# Run the unit tests with an ephemeral db
# source $APP_ROOT/unit_test.sh

# Deploy rbac to an ephemeral namespace for testing
source $CICD_ROOT/deploy_ephemeral_env.sh

# Run smoke tests with ClowdJobInvocation
source $CICD_ROOT/cji_smoke_test.sh
