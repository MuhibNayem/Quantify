package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total", Help: "Total HTTP requests",
	}, []string{"method", "path", "status"})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_request_duration_seconds", Help: "Request latency",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	CacheHitsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_hits_total", Help: "Cache hits",
	}, []string{"kind"})

	CacheMissesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_misses_total", Help: "Cache misses",
	}, []string{"kind"})

	JobsProcessedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "jobs_processed_total", Help: "Jobs processed",
	}, []string{"type", "status"})

	PaymentsProcessedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "payments_processed_total", Help: "Payments processed",
	}, []string{"method", "status"})

	AlertsFiredTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "alerts_fired_total", Help: "Alerts fired",
	}, []string{"type"})
)
