package service

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	remoteConfigurationsMetric = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "io_gathering_remote_configuration",
			Help: "The number of times a remote configuration was returned",
		},
		[]string{"file", "version"})
)
