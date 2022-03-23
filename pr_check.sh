#!/bin/bash

# --------------------------------------------
# Options that must be configured by app owner
# --------------------------------------------
export APP_NAME="io-gathering-service"  # name of app-sre "application" folder this component lives in
export COMPONENT_NAME="io-gathering-service"  # name of app-sre "resourceTemplate" in deploy.yaml for this component
export IMAGE="quay.io/cloudservices/io-gathering-conditions-service"

export IQE_PLUGINS="ccx"
export IQE_MARKER_EXPRESSION="io_gathering"
export IQE_FILTER_EXPRESSION=""
export IQE_CJI_TIMEOUT="10m"

# Install bonfire repo/initialize
CICD_URL=https://raw.githubusercontent.com/RedHatInsights/bonfire/master/cicd
curl -s $CICD_URL/bootstrap.sh > .cicd_bootstrap.sh && source .cicd_bootstrap.sh

# Build the image and push to quay
source $CICD_ROOT/build.sh

# Run the unit tests with an ephemeral db
# source $APP_ROOT/unit_test.sh

# Deploy rbac to an ephemeral namespace for testing
source $CICD_ROOT/deploy_ephemeral_env.sh

# Run smoke tests with ClowdJobInvocation
source $CICD_ROOT/cji_smoke_test.sh
