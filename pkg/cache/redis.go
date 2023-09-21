package cache

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
	"github.com/gesundheitscloud/go-svc/pkg/probe"
)

var (
	rdb *redis.Client
)

const (
	numConnectAttempts uint = 7 // with expTimeBackoff 2^7 = 2 minutes + eps
)

// define general error messages
var (
	ErrDBConnection   = errors.New("redis connection error")
	ErrRunCtxCanceled = errors.New("run context canceled by the user")
)

type RedisConnectionOptions struct {
	Host        string
	Port        int
	Password    string
	UseTLS      bool
	SSLRootCert string
}

func Initialize(runCtx context.Context, opts *RedisConnectionOptions) <-chan struct{} {
	dbUp := make(chan struct{})

	go func() {
		defer close(dbUp)
		if opts == nil {
			dbUp <- struct{}{} // nothing to be done, so show success
			return
		}
		connectFn := func() (*redis.Client, error) { return connect(opts) }
		client, err := retryExponential(runCtx, numConnectAttempts, 1*time.Second, connectFn)
		if err != nil {
			logging.LogErrorf(err, "Could not connect to redis")
			return
		}

		go func() {
			<-runCtx.Done()
			logging.LogInfof("run context canceled, closing redis connection")
			defer Close(client)
			defer logging.LogInfof("redis connection closed")
		}()

		rdb = client
		logging.LogInfof("connection to redis succeeded")
		dbUp <- struct{}{}
	}()

	return dbUp
}

// Get returns a handle to the redis object
func Get() *redis.Client {
	if rdb == nil {
		logging.LogErrorf(ErrDBConnection, "Get() - redis handle is nil")
		probe.Liveness().SetDead()
	}
	return rdb
}

// Close closes the redis connecton
func Close(client *redis.Client) {
	if client != nil {
		err := client.Close()
		if err != nil {
			logging.LogErrorf(err, "error closing redis")
		}
	}
}

func connect(opts *RedisConnectionOptions) (*redis.Client, error) {
	var tlsConfig *tls.Config
	if opts.UseTLS {
		cer, err := os.ReadFile(opts.SSLRootCert)
		if err != nil {
			logging.LogErrorf(ErrDBConnection, "could not load tls cert")
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(cer)
		tlsConfig = &tls.Config{RootCAs: caCertPool}
	}

	addr := fmt.Sprintf("%s:%d", opts.Host, opts.Port)
	logging.LogDebugf("Attempting to connect to redis: addr=%s tls=%s", addr, strconv.FormatBool(opts.UseTLS))

	client := redis.NewClient(&redis.Options{
		Addr:      addr,
		Password:  opts.Password,
		TLSConfig: tlsConfig,
		DB:        0, // use default DB
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		logging.LogErrorf(err, "connection to redis failed")
		return nil, err
	}
	return client, nil
}

// retryExponential runc function fn() as long as fn() returns no error, but maximally 'attempts' times
func retryExponential(runCtx context.Context, attempts uint, waitPeriod time.Duration, fn func() (*redis.Client, error)) (*redis.Client, error) {
	timeout := time.After(waitPeriod)
	logging.LogDebugf("retryExponential: timeout is %s ", waitPeriod)
	client, err := fn()
	if err != nil {
		if attempts--; attempts > 0 {
			select {
			case <-runCtx.Done():
				return nil, ErrRunCtxCanceled
			case <-timeout:
				logging.LogDebugf("timeout event - attempts = %d ", attempts)
				return retryExponential(runCtx, attempts, 2*waitPeriod, fn)
			}
		}
		return nil, err
	}
	return client, nil
}
