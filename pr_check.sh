#!/bin/bash

# --------------------------------------------
# Options that must be configured by app owner
# --------------------------------------------
APP_NAME="ccx-data-pipeline"  # name of app-sre "application" folder this component lives in
REF_ENV="insights-production"
COMPONENT_NAME="io-gathering-service"  # name of app-sre "resourceTemplate" in deploy.yaml for this component
COMPONENTS="io-gathering-service"  # space-separated list of components to laod
COMPONENTS_W_RESOURCES="io-gathering-service"  # component to keep
IMAGE="quay.io/cloudservices/io-gathering-conditions-service"

export IQE_PLUGINS="ccx"
export IQE_MARKER_EXPRESSION="io_gathering"
export IQE_FILTER_EXPRESSION=""
export IQE_CJI_TIMEOUT="60m"
export IQE_ENV_VARS="DYNACONF_USER_PROVIDER__rbac_enabled=false"

# Workaround to avoid issue with long name of namespace 'requestor'
# Jenkins job name is overriden.
# Issue: https://github.com/RedHatInsights/bonfire/issues/199
export JOB_NAME="io-gather-conds-serv-pr-check"

# Install bonfire repo/initialize
CICD_URL=https://raw.githubusercontent.com/RedHatInsights/bonfire/master/cicd
curl -s $CICD_URL/bootstrap.sh > .cicd_bootstrap.sh && source .cicd_bootstrap.sh

echo "creating PR image"
# Build the image and push to quay
source $CICD_ROOT/build.sh

source $CICD_ROOT/deploy_ephemeral_env.sh

source $CICD_ROOT/cji_smoke_test.sh
source $CICD_ROOT/post_test_results.sh  # publish results in Ibutsu
