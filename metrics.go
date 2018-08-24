package main

import (
	"net/url"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

type Metrics struct {
	*StatusInfo

	CurrentAddressTakenTime time.Duration
	RealAddressTakenTime    time.Duration
	UpdateTakenTime         time.Duration
}

func NewMetrics(info *StatusInfo) *Metrics {
	return &Metrics{StatusInfo: info}
}

func PushToPrometheus(server *url.URL, metrics *Metrics) error {
	gateway := push.New(server.String(), "updns")
	gateway.Grouping("domain", metrics.domain)

	if metrics.CurrentAddressTakenTime.Seconds() > 0 {
		currentAddress := prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "updns",
			Name:      "current_address_get_seconds",
			Help:      "time taken to obtain the current address that set into DNS server.",
		})
		currentAddress.Set(metrics.CurrentAddressTakenTime.Seconds())
		gateway.Collector(currentAddress)
	}

	if metrics.RealAddressTakenTime.Seconds() > 0 {
		realAddress := prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "updns",
			Name:      "real_address_get_seconds",
			Help:      "time taken to obtain the real IP address.",
		})
		realAddress.Set(metrics.RealAddressTakenTime.Seconds())
		gateway.Collector(realAddress)
	}

	if metrics.UpdateTakenTime.Seconds() > 0 {
		update := prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "updns",
			Name:      "record_update_seconds",
			Help:      "time taken to updating DNS record.",
		})
		update.Add(metrics.UpdateTakenTime.Seconds())
		gateway.Collector(update)
	}

	lastUpdated := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "updns",
		Name:      "last_updated_timestamp_seconds",
		Help:      "DNS record last updated time.",
	})
	lastUpdated.Add(float64(metrics.LastUpdated.Unix()))
	gateway.Collector(lastUpdated)

	updatedCount := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "updns",
		Name:      "executed_total",
		ConstLabels: prometheus.Labels{
			"type": "updated",
		},
		Help: "the count that updns was executed.",
	})
	updatedCount.Add(float64(metrics.UpdatedCount))
	gateway.Collector(updatedCount)

	notUpdatedCount := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "updns",
		Name:      "executed_total",
		ConstLabels: prometheus.Labels{
			"type": "not-updated",
		},
		Help: "the count that updns was executed.",
	})
	notUpdatedCount.Add(float64(metrics.ExecutedCount - metrics.UpdatedCount))
	gateway.Collector(notUpdatedCount)

	minorErrorCount := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "updns",
		Name:      "error_total",
		ConstLabels: prometheus.Labels{
			"type": "minor",
		},
		Help: "the count that coused errors.",
	})
	minorErrorCount.Add(float64(metrics.MinorErrorCount))
	gateway.Collector(minorErrorCount)

	fatalErrorCount := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "updns",
		Name:      "error_total",
		ConstLabels: prometheus.Labels{
			"type": "fatal",
		},
		Help: "the count that coused errors.",
	})
	fatalErrorCount.Add(float64(metrics.FatalErrorCount))
	gateway.Collector(fatalErrorCount)

	return gateway.Add()
}
