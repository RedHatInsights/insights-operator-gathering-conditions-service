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

###################
# Conditions
###################

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest AS conditions

RUN microdnf install --nodocs -y jq git

COPY get_conditions.sh .

RUN ./get_conditions.sh 

###################
# Builder
###################
FROM registry.access.redhat.com/ubi8/go-toolset:1.22.9-1.1736434946 AS builder

USER 0

ENV GOFLAGS="-buildvcs=false"

COPY . .

RUN make build && \
    chmod a+x ./insights-operator-gathering-conditions-service

###################
# Service
###################
FROM registry.access.redhat.com/ubi8/ubi-micro:latest

# copy the service
COPY --from=builder /opt/app-root/src/config.toml /config.toml
COPY --from=builder /opt/app-root/src/insights-operator-gathering-conditions-service .
COPY --from=builder /opt/app-root/src/openapi.json .

# copy the certificates
COPY --from=builder /etc/ssl /etc/ssl
COPY --from=builder /etc/pki /etc/pki

# copy the conditions
COPY --from=conditions  /conditions /conditions
COPY --from=conditions  /remote-configurations /remote-configurations
COPY --from=conditions  /mapping /mapping

USER 1001

CMD ["/insights-operator-gathering-conditions-service"]
