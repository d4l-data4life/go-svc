package bievents

import (
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"
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
	e.Lock()
	defer e.Unlock()
	return e.out.Encode(e.emit(event))
}

func (e *Emitter) emit(event Event) BaseEvent {
	var tenantID string
	if event.TenantID != "" {
		tenantID = event.TenantID
	} else {
		tenantID = e.tenantID
	}

	return BaseEvent{
		ServiceName:    e.serviceName,
		ServiceVersion: e.serviceVersion,
		HostName:       e.hostname,
		EventType:      "bi-event",
		Timestamp:      time.Now(),
		Event: Event{
			ActivityType:       event.ActivityType,
			UserID:             event.UserID,
			TenantID:           tenantID,
			ConsentDocumentKey: event.ConsentDocumentKey,
			Data:               event.Data,
		},
	}
}
