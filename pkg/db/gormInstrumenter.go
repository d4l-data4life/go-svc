package db

import (
	"context"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gorm.io/gorm"

	"github.com/gesundheitscloud/go-svc/pkg/prom"
)

type contextKey string

const (
	QueryIDContextKey contextKey = "query-id"
)

type Instrumenter struct {
}

func NewInstrumenter() *Instrumenter {
	return &Instrumenter{}
}

func createSetWithLock(m map[string]time.Time, mutex *sync.RWMutex) func(key string, ts time.Time) {
	return func(key string, ts time.Time) {
		mutex.Lock()
		defer mutex.Unlock()
		m[key] = ts
	}
}

func createGetWithLock(m map[string]time.Time, mutex *sync.RWMutex) func(key string) (time.Time, bool) {
	return func(key string) (time.Time, bool) {
		mutex.Lock()
		defer mutex.Unlock()
		ts, ok := m[key]
		return ts, ok
	}
}

func createDeleteWithLock(m map[string]time.Time, mutex *sync.RWMutex) func(key string) {
	return func(key string) {
		mutex.Lock()
		defer mutex.Unlock()
		delete(m, key)
	}
}

func createBeforeRequestCallback(m map[string]time.Time, mMutex *sync.RWMutex) func(tx *gorm.DB) {
	setWithLock := createSetWithLock(m, mMutex)
	return func(tx *gorm.DB) {
		// nolint: gosec
		queryID := strconv.Itoa(rand.Int())
		tx.Statement.Context = context.WithValue(tx.Statement.Context, QueryIDContextKey, queryID)
		setWithLock(queryID, time.Now())
	}
}

func createAfterRequestCallback(m map[string]time.Time, mMutex *sync.RWMutex, metric *prometheus.HistogramVec) func(tx *gorm.DB) {
	getWithLock := createGetWithLock(m, mMutex)
	deleteWithLock := createDeleteWithLock(m, mMutex)

	return func(tx *gorm.DB) {
		queryID := tx.Statement.Context.Value(QueryIDContextKey).(string)
		if ts, ok := getWithLock(queryID); ok {
			now := time.Now()
			duration := now.Sub(ts)
			durationSeconds := float64(duration) / float64(time.Second)

			sqlString := tx.Statement.SQL.String()

			var sqlStringLabel string
			if len(sqlString) > 29 {
				sqlStringLabel = sqlString[0:29]
			} else {
				sqlStringLabel = sqlString
			}

			metric.WithLabelValues(sqlStringLabel).Observe(durationSeconds)
			deleteWithLock(queryID)
		}
	}
}

func registerDbRequestDurationMetric() *prometheus.HistogramVec {
	return promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: prom.GetNamespace(),
		Name:      "db_request_duration_seconds",
		Help:      "histogram",
		Buckets:   []float64{.001, .01, .1, .25, .5, 1, 2.5, 5, 10},
	}, []string{"sqlstring"})
}

func (i *Instrumenter) Name() string {
	return "gorm:instrumenter"
}

// Initialize adds gorm Plugin for collecting database request metrics
func (i *Instrumenter) Initialize(_ *gorm.DB) (err error) {
	m := make(map[string]time.Time)
	// mutex for goroutine safe map access
	mutex := &sync.RWMutex{}

	conn := Get()

	dbRequestDurationMetric := registerDbRequestDurationMetric()

	err = conn.Callback().Create().Before("gorm:create").Register("gorminstrumenter:before_create", createBeforeRequestCallback(m, mutex))
	if err != nil {
		return err
	}
	err = conn.Callback().
		Create().
		After("gorm:create").
		Register("gorminstrumenter:after_create", createAfterRequestCallback(m, mutex, dbRequestDurationMetric))
	if err != nil {
		return err
	}
	err = conn.Callback().Query().Before("gorm:query").Register("gorminstrumenter:before_query", createBeforeRequestCallback(m, mutex))
	if err != nil {
		return err
	}
	err = conn.Callback().
		Query().
		After("gorm:query").
		Register("gorminstrumenter:after_query", createAfterRequestCallback(m, mutex, dbRequestDurationMetric))
	if err != nil {
		return err
	}
	err = conn.Callback().Delete().Before("gorm:delete").Register("gorminstrumenter:before_delete", createBeforeRequestCallback(m, mutex))
	if err != nil {
		return err
	}
	err = conn.Callback().
		Delete().
		After("gorm:delete").
		Register("gorminstrumenter:after_delete", createAfterRequestCallback(m, mutex, dbRequestDurationMetric))
	if err != nil {
		return err
	}
	err = conn.Callback().Update().Before("gorm:update").Register("gorminstrumenter:before_update", createBeforeRequestCallback(m, mutex))
	if err != nil {
		return err
	}
	err = conn.Callback().
		Update().
		After("gorm:update").
		Register("gorminstrumenter:after_update", createAfterRequestCallback(m, mutex, dbRequestDurationMetric))
	if err != nil {
		return err
	}
	return err
}
