package tut

import (
	"context"
	"fmt"
	"sync"

	"github.com/gesundheitscloud/go-svc/pkg/log"
)

type AuditEvent struct {
	Type     log.AuditLogType
	SubType  interface{}  // SubType is only relevant for access and change logs
	Resource fmt.Stringer // Resource is only relevant for access and change logs
	Event    fmt.Stringer // Event is only relevant for security events
	Success  bool         // Success is only relevant for security events
}

// AuditLogger is a helper structure for unit testing the audit events that were logged.
// It implements the Audit log methods and records internally which events were logged.
type AuditLogger struct {
	auditedEvents map[AuditEvent]struct{}
	eventsMu      sync.Mutex
}

// NewAuditLogger creates a new AuditLogger
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{
		auditedEvents: make(map[AuditEvent]struct{}),
	}
}

// AuditWasLogged allows to check if a particular event was logged
func (l *AuditLogger) AuditWasLogged(event AuditEvent) bool {
	l.eventsMu.Lock()
	defer l.eventsMu.Unlock()

	_, ok := l.auditedEvents[event]
	return ok
}

// AuditWasNotLogged allows to check if a particular event was NOT logged
func (l *AuditLogger) AuditWasNotLogged(event AuditEvent) bool {
	l.eventsMu.Lock()
	defer l.eventsMu.Unlock()

	_, ok := l.auditedEvents[event]
	return !ok
}

// GetAllAuditLogs allows to get all the audit logs that were logged with this instance of AuditLogger.
func (l *AuditLogger) GetAllAuditLogs() []AuditEvent {
	l.eventsMu.Lock()
	defer l.eventsMu.Unlock()

	result := make([]AuditEvent, 0, len(l.auditedEvents))
	for k := range l.auditedEvents {
		result = append(result, k)
	}

	return result
}

// AuditSecuritySuccess is part of the audit logger interface and will be called by the unit under test
func (l *AuditLogger) AuditSecuritySuccess(
	ctx context.Context,
	event fmt.Stringer,
	extras ...log.ExtraAuditInfoProvider,
) error {
	l.eventsMu.Lock()
	defer l.eventsMu.Unlock()

	l.auditedEvents[AuditEvent{
		Type:    log.SecurityLog,
		Event:   event,
		Success: true,
	}] = struct{}{}
	return nil
}

// AuditSecurityFailure is part of the audit logger interface and will be called by the unit under test
func (l *AuditLogger) AuditSecurityFailure(
	ctx context.Context,
	event fmt.Stringer,
	extras ...log.ExtraAuditInfoProvider,
) error {
	l.eventsMu.Lock()
	defer l.eventsMu.Unlock()

	l.auditedEvents[AuditEvent{
		Type:    log.SecurityLog,
		Event:   event,
		Success: false,
	}] = struct{}{}
	return nil
}

func (l *AuditLogger) AuditRead(
	ctx context.Context,
	ownerID, resourceType, _ fmt.Stringer,
	extras ...log.ExtraAuditInfoProvider,
) error {
	l.eventsMu.Lock()
	defer l.eventsMu.Unlock()

	l.auditedEvents[AuditEvent{
		Type:     log.AccessLog,
		SubType:  log.Read,
		Resource: resourceType,
	}] = struct{}{}
	return nil
}

func (l *AuditLogger) AuditCreate(
	ctx context.Context,
	ownerID, resourceType, _ fmt.Stringer,
	_ interface{},
	extras ...log.ExtraAuditInfoProvider,
) error {
	l.eventsMu.Lock()
	defer l.eventsMu.Unlock()

	l.auditedEvents[AuditEvent{
		Type:     log.ChangeLog,
		SubType:  log.Create,
		Resource: resourceType,
	}] = struct{}{}
	return nil
}

func (l *AuditLogger) AuditUpdate(
	ctx context.Context,
	ownerID, resourceType, _ fmt.Stringer,
	_ interface{},
	extras ...log.ExtraAuditInfoProvider,
) error {
	l.eventsMu.Lock()
	defer l.eventsMu.Unlock()

	l.auditedEvents[AuditEvent{
		Type:     log.ChangeLog,
		SubType:  log.Update,
		Resource: resourceType,
	}] = struct{}{}
	return nil
}

func (l *AuditLogger) AuditDelete(
	ctx context.Context,
	ownerID, resourceType, _ fmt.Stringer,
	extras ...log.ExtraAuditInfoProvider,
) error {
	l.eventsMu.Lock()
	defer l.eventsMu.Unlock()

	l.auditedEvents[AuditEvent{
		Type:     log.ChangeLog,
		SubType:  log.Delete,
		Resource: resourceType,
	}] = struct{}{}
	return nil
}

type AuditCheckFunc func(*AuditLogger) error

func HasAuditEntry(wantEvent AuditEvent) AuditCheckFunc {
	return func(al *AuditLogger) error {
		if ok := al.AuditWasLogged(wantEvent); !ok {
			return fmt.Errorf("expected event not found in the audit logs. Want: %v. Got: %v",
				wantEvent,
				al.GetAllAuditLogs(),
			)
		}

		return nil
	}
}

func HasNoAuditEntry(wantEvent AuditEvent) AuditCheckFunc {
	return func(al *AuditLogger) error {
		if ok := al.AuditWasNotLogged(wantEvent); !ok {
			return fmt.Errorf("expected missing event in the audit logs, but got %v", wantEvent)
		}

		return nil
	}
}

func CheckAllAudits(checks ...AuditCheckFunc) AuditCheckFunc {
	return func(al *AuditLogger) error {
		for _, check := range checks {
			err := check(al)
			if err != nil {
				return err
			}
		}
		return nil
	}
}
