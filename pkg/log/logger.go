package log

import (
	"context"
	"io"
	"os"
	"sync"
)

// nolint: revive
type LogType string

const (
	Audit       LogType = "audit"
	Operational LogType = "operational"
)

// Logger enables GHC-compliant logging. The zero-value is not usable;
// initialise with NewLogger.
type Logger struct {
	serviceName    string
	serviceVersion string
	hostname       string
	podName        string
	environment    string
	tenantID       string

	sync.Mutex
	out Encoder
}

func parseContext(ctx context.Context) (traceID, userID, clientID string) {
	if ctx != nil {
		traceID, _ = ctx.Value(TraceIDContextKey).(string)
		userID, _ = ctx.Value(UserIDContextKey).(string)
		clientID, _ = ctx.Value(ClientIDContextKey).(string)
	}
	return traceID, userID, clientID
}

func getFromContext(ctx context.Context, key contextKey) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(key).(string)
	return value
}

// getFromContextWithDefault tries to get a value from the context.
// If the value is not found, it uses the default value.
// nolint: unparam
func getFromContextWithDefault(ctx context.Context, key contextKey, defaultValue string) string {
	value := getFromContext(ctx, key)
	if value != "" {
		return value
	}
	return defaultValue
}

// WithEncoder configures the logger with the provided encoder
// Encoders will deprecate WithWriter at some point
// Using it simultaneously with WithWriter will cause overrides
func WithEncoder(e Encoder) func(*Logger) {
	return func(l *Logger) {
		l.out = e
	}
}

// WithWriter is an option for NewLogger which lets the caller specify where to
// dump the logs.
// WithWriter defaults to JSON encoding
// Using it simultaneously with WithEncoder will cause overrides
func WithWriter(w io.Writer) func(*Logger) {
	return func(l *Logger) {
		l.out = NewJSONEncoder(w)
	}
}

func WithPodName(n string) func(*Logger) {
	return func(l *Logger) {
		l.podName = n
	}
}

func WithEnv(e string) func(*Logger) {
	return func(l *Logger) {
		l.environment = e
	}
}

// WithTenantID is an option for NewLogger which lets the caller specify the
// tenant ID for which the logger instance is logging.
// If a per-request tenant ID is needed, that needs to be specified in
// the http wrapper and passed with the context.
// In case of conflict, the tenant ID from the context will be logged.
func WithTenantID(tenantID string) func(*Logger) {
	return func(l *Logger) {
		l.tenantID = tenantID
	}
}

// NewLogger initialises a Logger with the values that are not susceptible to
// change in the service lifetime.
// All logs will be dumped to os.Stdout, unless a WithWriter option is passed.
func NewLogger(serviceName, serviceVersion, hostname string, options ...func(*Logger)) *Logger {
	l := &Logger{
		serviceName:    serviceName,
		serviceVersion: serviceVersion,
		hostname:       hostname,
		out:            NewJSONEncoder(os.Stdout),
	}

	for _, apply := range options {
		apply(l)
	}

	return l
}

// Log marshals the given value as JSON and writes it to the logger's
// io.Writer.
// Log is safe for concurrent use.
func (l *Logger) Log(v interface{}) error {
	l.Lock()
	defer l.Unlock()

	return l.out.Encode(v)
}
