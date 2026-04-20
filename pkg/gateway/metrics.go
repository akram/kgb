package gateway

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	evaluations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kgb_evaluations_total",
			Help: "Total gateway evaluate calls by outcome label.",
		},
		[]string{"outcome"},
	)
)
