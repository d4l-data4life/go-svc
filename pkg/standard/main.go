package standard

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/cache"
	"github.com/gesundheitscloud/go-svc/pkg/channels"
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

func WithPostgres(opts *db2.ConnectionOptions) MainOption {
	return func(runCtx context.Context) {
		dboptions = opts
	}
}

func WithPostgresDB2(opts *db2.ConnectionOptions) MainOption {
	return func(runCtx context.Context) {
		dboptions = opts
	}
}

func WithRedis(opts *cache.RedisConnectionOptions) MainOption {
	return func(runCtx context.Context) {
		redisoptions = opts
	}
}

var dboptions *db2.ConnectionOptions
var redisoptions *cache.RedisConnectionOptions

// Main is a wrapper function over main - it handles the typical tasks like starting DB connection, handling OS signals, etc.
func Main(serviceMain MainFunction, svcName string, options ...MainOption) {
	probe.Liveness().SetLive()
	// runCtx is a running context. Canceling this context means that the service should stop running asap
	runCtx, runCtxCancelFunc := context.WithCancel(context.Background())
	defer runCtxCancelFunc()
	go setupSingals(runCtx, runCtxCancelFunc)

	for _, option := range options {
		option(runCtx)
	}

	redisUp := cache.Initialize(runCtx, redisoptions)
	dbUp := db2.Initialize(runCtx, dboptions)

	// The above channels use the following convention:
	// 1. closing means that the initialization procedures has finished (either successfully or with error)
	// 2. on success: A value has been sent through the channel before closing
	// 3. on error: No value has been sent before closing

	if waitForDB(runCtx, dbUp, redisUp) {
		logging.LogInfof("dbs connected")
	} else {
		runCtxCancelFunc()
	}

	mainStopped := serviceMain(runCtx, svcName)
	logging.LogInfof("service is up and running!")

	logging.LogInfof("%s: waiting for: (1) error, (2) HTTP server to stop, (3) run context canceled by the user", svcName)

	// Block until serviceMain closes the channel
	for range mainStopped {
		logging.LogInfof("%s exits - HTTP server stopped", svcName)
		return
	}
	logging.LogInfof("%s exits - run context canceled", svcName)
}

// waitForDB returns true when DB and/or Redis is up and connected, false when DB connection failed and the service should be shutdown
func waitForDB(ctx context.Context, dbUp <-chan struct{}, redisUp <-chan struct{}) bool {
	logging.LogInfof("Waiting up to 2 minutes for DB connection(s)...")
	mergedChannel := channels.Barrier(ctx.Done(), dbUp, redisUp)
	for range channels.OrDoneTimeout(ctx.Done(), time.After(120*time.Second), mergedChannel) {
		// a message on dbUp = database is ready
		return true
	}
	// context canceled or timeout = still not connected to the database
	logging.LogInfof("waitForDB: timeout or waiting canceled by the user")
	probe.Liveness().SetDead()
	return false
}
