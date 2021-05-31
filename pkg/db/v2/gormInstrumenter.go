package db

// import (
// 	"sync"
// 	"time"

// 	"github.com/prometheus/client_golang/prometheus"
// 	"github.com/prometheus/client_golang/prometheus/promauto"
// 	"gorm.io/gorm"
// )

// func createSetWithLock(m map[string]time.Time, mutex *sync.RWMutex) func(key string, ts time.Time) {
// 	return func(key string, ts time.Time) {
// 		mutex.Lock()
// 		defer mutex.Unlock()
// 		m[key] = ts
// 	}
// }

// func createGetWithLock(m map[string]time.Time, mutex *sync.RWMutex) func(key string) (time.Time, bool) {
// 	return func(key string) (time.Time, bool) {
// 		mutex.Lock()
// 		defer mutex.Unlock()
// 		ts, ok := m[key]
// 		return ts, ok
// 	}
// }

// func createDeleteWithLock(m map[string]time.Time, mutex *sync.RWMutex) func(key string) {
// 	return func(key string) {
// 		mutex.Lock()
// 		defer mutex.Unlock()
// 		delete(m, key)
// 	}
// }

// func createBeforeRequestCallback(m map[string]time.Time, mMutex *sync.RWMutex) func(scope *gorm.Scope) {
// 	setWithLock := createSetWithLock(m, mMutex)
// 	return func(scope *gorm.Scope) {
// 		setWithLock(scope.InstanceID(), time.Now())
// 	}
// }

// func createAfterRequestCallback(m map[string]time.Time, mMutex *sync.RWMutex, metric *prometheus.HistogramVec) func(scope *gorm.Scope) {
// 	getWithLock := createGetWithLock(m, mMutex)
// 	deleteWithLock := createDeleteWithLock(m, mMutex)

// 	return func(scope *gorm.Scope) {
// 		if ts, ok := getWithLock(scope.InstanceID()); ok {
// 			now := time.Now()
// 			duration := now.Sub(ts)
// 			durationSeconds := float64(duration) / float64(time.Second)

// 			var sqlStringLabel string
// 			if len(scope.SQL) > 29 {
// 				sqlStringLabel = scope.SQL[0:29]
// 			} else {
// 				sqlStringLabel = scope.SQL
// 			}

// 			metric.WithLabelValues(sqlStringLabel).Observe(durationSeconds)
// 			deleteWithLock(scope.InstanceID())
// 		}
// 	}
// }

// func registerDbRequestDurationMetric() *prometheus.HistogramVec {
// 	return promauto.NewHistogramVec(prometheus.HistogramOpts{
// 		Namespace: "d4l",
// 		Name:      "db_request_duration_seconds",
// 		Help:      "histogram",
// 		Buckets:   []float64{.001, .01, .1, .25, .5, 1, 2.5, 5, 10},
// 	}, []string{"sqlstring"})

// }

// // RegisterInstrumenterPlugin adds gorm Plugin for collecting database request metrics
// func registerInstrumenterPlugin() {
// 	m := make(map[string]time.Time)
// 	//mutex for goroutine safe map access
// 	mutex := &sync.RWMutex{}

// 	// disable internal logs, as we have our own ones
// 	conn := Get().LogMode(false)

// 	dbRequestDurationMetric := registerDbRequestDurationMetric()

// 	conn.Callback().Create().Before("gorm:create").Register("gorminstrumenter:before_create", createBeforeRequestCallback(m, mutex))
// 	conn.Callback().Create().After("gorm:create").Register("gorminstrumenter:after_create", createAfterRequestCallback(m, mutex, dbRequestDurationMetric))

// 	conn.Callback().Query().Before("gorm:query").Register("gorminstrumenter:before_querz", createBeforeRequestCallback(m, mutex))
// 	conn.Callback().Query().After("gorm:query").Register("gorminstrumenter:after_query", createAfterRequestCallback(m, mutex, dbRequestDurationMetric))

// 	conn.Callback().Delete().Before("gorm:delete").Register("gorminstrumenter:before_delete", createBeforeRequestCallback(m, mutex))
// 	conn.Callback().Delete().After("gorm:delete").Register("gorminstrumenter:after_delete", createAfterRequestCallback(m, mutex, dbRequestDurationMetric))

// 	conn.Callback().Update().Before("gorm:update").Register("gorminstrumenter:before_update", createBeforeRequestCallback(m, mutex))
// 	conn.Callback().Update().After("gorm:update").Register("gorminstrumenter:after_update", createAfterRequestCallback(m, mutex, dbRequestDurationMetric))
// }
