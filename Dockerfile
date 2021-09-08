###################
# Conditions
###################
FROM quay.io/redhatinsights/insights-operator-gathering-conditions:latest AS conditions

###################
# Builder
###################
FROM registry.redhat.io/rhel8/go-toolset:1.14 AS builder
WORKDIR $GOPATH/src/github.com/redhatinsights/insights-conditions-service

USER 0

COPY . .

RUN make build && \
    chmod a+x bin/insights-conditions-service


###################
# Service
###################
FROM registry.redhat.io/ubi8-minimal:latest

# copy the service
COPY --from=builder $GOPATH/src/github.com/redhatinsights/insights-conditions-service/config/config.toml /config/config.toml
COPY --from=builder $GOPATH/src/github.com/redhatinsights/insights-conditions-service/bin/insights-conditions-service .

# copy the conditions
COPY --from=conditions /conditions /conditions
# COPY /rules /conditions

USER 1001

CMD ["/insights-conditions-service"]