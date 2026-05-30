// Package metrics exposes Prometheus HTTP metrics and a Gin middleware that
// records request count and latency per route (NFR §5.5).
package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests by method, route and status.",
	}, []string{"method", "route", "status"})

	httpDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency by method and route.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route"})
)

func init() {
	prometheus.MustRegister(httpRequests, httpDuration)
}

// Middleware records request count and latency. Uses the route pattern
// (c.FullPath) as the label to bound cardinality.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		httpRequests.WithLabelValues(c.Request.Method, route, strconv.Itoa(c.Writer.Status())).Inc()
		httpDuration.WithLabelValues(c.Request.Method, route).Observe(time.Since(start).Seconds())
	}
}

// Handler serves the Prometheus exposition endpoint.
func Handler() gin.HandlerFunc { return gin.WrapH(promhttp.Handler()) }
