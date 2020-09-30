package db

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func createBeforeRequestCallback(m map[string]time.Time) func(scope *gorm.Scope) {
	return func(scope *gorm.Scope) {
		m[scope.InstanceID()] = time.Now()
	}
}

func createAfterRequestCallback(m map[string]time.Time, metric *prometheus.HistogramVec) func(scope *gorm.Scope) {
	return func(scope *gorm.Scope) {
		if ts, ok := m[scope.InstanceID()]; ok {
			now := time.Now()
			duration := now.Sub(ts)
			durationSeconds := float64(duration) / float64(time.Second)
			var sqlStringLabel string
			if len(scope.SQL) > 29 {
				sqlStringLabel = scope.SQL[0:29]
			} else {
				sqlStringLabel = scope.SQL
			}
			metric.WithLabelValues(sqlStringLabel).Observe(durationSeconds)
			delete(m, scope.InstanceID())
		}
	}
}

func registerDbRequestDurationMetric() *prometheus.HistogramVec {
	return promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "d4l",
		Name:      "db_request_duration_seconds",
		Help:      "histogram",
		Buckets:   []float64{.001, .01, .1, .25, .5, 1, 2.5, 5, 10},
	}, []string{"sqlstring"})

}

// RegisterInstrumenterPlugin adds gorm Plugin for collecting database request metrics
func registerInstrumenterPlugin() {
	m := make(map[string]time.Time)
	conn := Get()

	dbRequestDurationMetric := registerDbRequestDurationMetric()

	conn.Callback().Create().Before("gorm:create").Register("gorminstrumenter:before_create", createBeforeRequestCallback(m))
	conn.Callback().Create().After("gorm:create").Register("gorminstrumenter:after_create", createAfterRequestCallback(m, dbRequestDurationMetric))

	conn.Callback().Query().Before("gorm:query").Register("gorminstrumenter:before_querz", createBeforeRequestCallback(m))
	conn.Callback().Query().After("gorm:query").Register("gorminstrumenter:after_query", createAfterRequestCallback(m, dbRequestDurationMetric))

	conn.Callback().Delete().Before("gorm:delete").Register("gorminstrumenter:before_delete", createBeforeRequestCallback(m))
	conn.Callback().Delete().After("gorm:delete").Register("gorminstrumenter:after_delete", createAfterRequestCallback(m, dbRequestDurationMetric))

	conn.Callback().Update().Before("gorm:update").Register("gorminstrumenter:before_update", createBeforeRequestCallback(m))
	conn.Callback().Update().After("gorm:update").Register("gorminstrumenter:after_update", createAfterRequestCallback(m, dbRequestDurationMetric))
}
