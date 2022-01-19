package standard

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gesundheitscloud/go-svc/internal/channels"
	"github.com/gesundheitscloud/go-svc/pkg/cache"
	"github.com/gesundheitscloud/go-svc/pkg/db"
	"github.com/gesundheitscloud/go-svc/pkg/db2"
	"github.com/gesundheitscloud/go-svc/pkg/logging"
	"github.com/gesundheitscloud/go-svc/pkg/probe"
)

// setupSingals connects SIGTERM and SIGINT with cancelation of the running context
func setupSingals(runCtx context.Context, stopService context.CancelFunc) {
	termSignal := make(chan os.Signal, 1)
	defer close(termSignal)
	signal.Notify(termSignal, syscall.SIGTERM, syscall.SIGINT)
	for {
		select {
		case <-runCtx.Done():
			probe.Liveness().SetDead()
			return
		case <-termSignal:
			stopService()
			return
		}
	}
}

type MainFunction func(context.Context, string) <-chan struct{}

type MainOption func(context.Context)

func WithPostgres(opts *db.ConnectionOptions) MainOption {
	return func(runCtx context.Context) {
		dboptions = opts
	}
}

func WithPostgresDB2(opts *db2.ConnectionOptions) MainOption {
	return func(runCtx context.Context) {
		db2options = opts
	}
}

func WithRedis(opts *cache.RedisConnectionOptions) MainOption {
	return func(runCtx context.Context) {
		redisoptions = opts
	}
}

var dboptions *db.ConnectionOptions
var db2options *db2.ConnectionOptions
var redisoptions *cache.RedisConnectionOptions

// Main is a wrapper function over main - it handles the typical tasks like starting DB connection, handling OS singnals, etc.
func Main(serviceMain MainFunction, svcName string, options ...MainOption) {
	probe.Liveness().SetLive()
	// runCtx is a running context. Canceling this contexs means that the service should stop running asap
	runCtx, stopService := context.WithCancel(context.Background())
	defer stopService()
	go setupSingals(runCtx, stopService)

	for _, option := range options {
		option(runCtx)
	}

	redisUp := cache.Initialize(runCtx, redisoptions)
	dbUp := db.Initialize(runCtx, dboptions)
	db2Up := db2.Initialize(runCtx, db2options)

	if waitForDB(runCtx, dbUp, redisUp, db2Up) {
		logging.LogInfof("dbs connected")
	} else {
		stopService()
	}

	mainStopped := serviceMain(runCtx, svcName)
	logging.LogInfof("service is up and running!")

	logging.LogInfof("%s: waiting for: (1) error, (2) HTTP server to stop, (3) run context canceled by the user", svcName)

	for range channels.OrDone(runCtx.Done(), mainStopped) {
		logging.LogInfof("%s exits - HTTP server stopped", svcName)
		return
	}
	logging.LogInfof("%s exits - run context canceled", svcName)
}

// waitForDB returns true when DB and/or Redis is up and connected, false when DB connection failed and the service should be shutdown
func waitForDB(ctx context.Context, dbUp <-chan struct{}, redisUp <-chan struct{}, db2Up <-chan struct{}) bool {
	logging.LogInfof("Waiting up to 2 minutes for DB connection...")
	mergedChannel := channels.FanIn(ctx.Done(), dbUp, redisUp, db2Up)
	for range channels.OrDoneTimeout(ctx.Done(), time.After(120*time.Second), mergedChannel) {
		// a message on dbUp = database is ready
		return true
	}
	// context canceled or timeout = still not connected to the database
	logging.LogInfof("waitForDB: timeout or waiting canceled by the user")
	probe.Liveness().SetDead()
	return false
}
