package bievents

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/gesundheitscloud/go-svc/pkg/jwt"
	"github.com/gofrs/uuid"
)

// Emitter enables business intelligence complaint events.
type Emitter struct {
	serviceName    string
	serviceVersion string
	hostname       string
	tenantID       string
	sync.Mutex
	out *json.Encoder
}

// BaseEvent is the concrete event which is common to all the events.
// Parameters ServiceName, ServiceVersion, HostName, EventType, Timestamp are populated by library.
// Paremeters in Event are passed on from the consumer of the library.
type BaseEvent struct {
	ServiceName    string    `json:"service-name"`
	ServiceVersion string    `json:"service-version"`
	HostName       string    `json:"hostname"`
	EventType      string    `json:"event-type"`
	EventID        uuid.UUID `json:"event-id"`
	Timestamp      time.Time `json:"timestamp"`
	// Embed Event struct. This will enable Event json to be printed on the same level as BaseEvent during json encoding.
	Event
}

// State represents the state of a BI event:
// Success: the event happened successful
// Failure: the event was attempted but failed
// Attempt: the event was attempted but not further information about its success is available at that time.
type State string

const (
	Success State = "success"
	Failure State = "failure"
	Attempt State = "attempt"
)

// Event is the info passed on specific to event.
type Event struct {
	ActivityType       ActivityType `json:"activity-type"`
	UserID             string       `json:"user-id"`
	Data               interface{}  `json:"data"`
	TenantID           string       `json:"tenant-id"`
	ConsentDocumentKey string       `json:"consent-document-key"`
	State              State        `json:"state,omitempty"`
	EventSource        string       `json:"event-source"`
	SessionID          string       `json:"session-id"`
}

// WithWriter is an option for NewEventEmitter which lets the caller specify where to
// dump the logs.
// WithWriter defaults to JSON encoding
// Using it simultaneously with WithEncoder will cause overrides
func WithWriter(w io.Writer) func(*Emitter) {
	return func(l *Emitter) {
		l.out = json.NewEncoder(w)
	}
}

// WithTenantID is an option for NewEventEmitter which lets the caller specify which tenant
// ID the event emitter represents. This is useful when the tenant ID is constant for
// all the events logged by on instance.
// The tenant ID explicitly set for an event will override this value.
func WithTenantID(tenantID string) func(*Emitter) {
	return func(l *Emitter) {
		l.tenantID = tenantID
	}
}

// NewEventEmitter creates a new event emitter
// All events will be dumped to os.Stdout, unless a WithWriter option is passed.
func NewEventEmitter(serviceName, serviceVersion, hostname string, options ...func(*Emitter)) *Emitter {
	e := &Emitter{
		serviceName:    serviceName,
		serviceVersion: serviceVersion,
		hostname:       hostname,

		out: json.NewEncoder(os.Stdout),
	}

	for _, apply := range options {
		apply(e)
	}
	return e
}

// Log marshals the given value as JSON and writes it to the logger's
// io.Writer.
// Log is safe for concurrent use.
// Log logs the event.
func (e *Emitter) Log(event Event) error {
	if event.EventSource == "" {
		event.EventSource = GetEventSource(2)
	}

	e.Lock()
	defer e.Unlock()
	return e.out.Encode(e.emit(event))
}

// LogCtx Wraps Log(event Event) and extracts userID and SessionID from context
func (e *Emitter) LogCtx(ctx context.Context, event Event) error {
	if event.SessionID == "" {
		appID, err := jwt.GetAppID(ctx)
		if err != nil {
			return err
		}
		event.SessionID = Hash(appID.Bytes())
	}

	if event.UserID == "" {
		userID, err := jwt.GetUserID(ctx)
		if err != nil {
			return err
		}
		event.UserID = userID.String()
	}

	if event.EventSource == "" {
		event.EventSource = GetEventSource(2)
	}

	return e.Log(event)
}

func (e *Emitter) emit(event Event) BaseEvent {
	if event.TenantID == "" {
		event.TenantID = e.tenantID
	}

	return BaseEvent{
		ServiceName:    e.serviceName,
		ServiceVersion: e.serviceVersion,
		HostName:       e.hostname,
		EventType:      "bi-event",
		EventID:        uuid.Must(uuid.NewV4()),
		Timestamp:      time.Now().Truncate(time.Second),
		Event:          event,
	}
}

// GetEventSource returns the file and line where this func was called from
// example: /pkg/handlers/userConsentHandler.go:292
// the skip parameter can be used to ascend up the call stack
func GetEventSource(skip int) string {
	_, fn, line, _ := runtime.Caller(skip)
	return fmt.Sprintf("%s:%d", fn, line)
}

// Hash uses SHA 256 to hash data and transform it into a string.
// It can be used to hash pieces of information that shouldn't be leaked to the
// BI events.
func Hash(data []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(data))
}
