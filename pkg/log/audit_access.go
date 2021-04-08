package log

import (
	"context"
	"fmt"
	"time"
)

type AccessLogEvent string

const (
	Read AccessLogEvent = "read"
)

type accessLog struct {
	baseAuditLog

	// owner is the user owning the resource accessed
	OwnerID        string         `json:"owner-id"`
	EventType      AccessLogEvent `json:"event-type"`
	ResourceType   string         `json:"resource-type"`
	AdditionalData interface{}    `json:"additional-data,omitempty"`
}

type singleAccessLog struct {
	accessLog
	ResourceID string `json:"resource-id"`
}

type bulkAccessLog struct {
	accessLog
	ResourceIDs interface{} `json:"resource-ids"` // use interface{} here to allow any kind of array
}

// AuditRead should be used to log a read access to some personal sensitive piece of information.
// It will attempt to get `TraceID`, `ClientID`, `RequestURL`, `RequestDomain`, `CallerIPAddress` and `SubjectID`
// from the context.
// `ownerID` represents the owner of the resource.
// `resourceType` should be used to specify what kind of information was accessed
// `resourceID` must include the information needed to uniquely identify the resource (ID of the resource or the data set)
// `extras` allows to add optional information or override default values:
//    - `SubjectID` allows to override the ID of the user performing the action (by default it is expected in the context)
//    - `AdditionalData` allows to provide any extra information relevant for the audit log
func (l *Logger) AuditRead(
	ctx context.Context,
	ownerID fmt.Stringer,
	resourceType fmt.Stringer,
	resourceID fmt.Stringer,
	extras ...ExtraAuditInfoProvider,
) error {
	log := singleAccessLog{
		accessLog: accessLog{
			baseAuditLog: baseAuditLog{
				Timestamp:       time.Now(),
				LogType:         Audit,
				AuditLogType:    AccessLog,
				TraceID:         getFromContext(ctx, TraceIDContextKey),
				ServiceName:     l.serviceName,
				ServiceVersion:  l.serviceVersion,
				Hostname:        l.hostname,
				PodName:         l.podName,
				Environment:     l.environment,
				ClientID:        getFromContext(ctx, ClientIDContextKey),
				RequestURL:      getFromContext(ctx, RequestURLContextKey),
				RequestDomain:   getFromContext(ctx, RequestDomainContextKey),
				CallerIPAddress: getFromContext(ctx, CallerIPContextKey),
				SubjectID:       getFromContext(ctx, UserIDContextKey),
				TenantID:        getFromContextWithDefault(ctx, TenantIDContextKey, l.tenantID),
			},
			OwnerID:      ownerID.String(),
			EventType:    Read,
			ResourceType: resourceType.String(),
		},
		ResourceID: resourceID.String(),
	}

	for _, f := range extras {
		f(&log.accessLog)
	}

	return l.Log(log)
}

// AuditBulkRead should be used to log a read access to some personal sensitive piece of information.
// It will attempt to get `TraceID`, `ClientID`, `RequestURL`, `RequestDomain`, `CallerIPAddress` and `SubjectID`
// from the context.
// `ownerID` represents the owner of the resource.
// `resourceType` should be used to specify what kind of information was accessed
// `resourceIDs` must include the information needed to uniquely identify all the resources (IDs of the resource or the data set)
// `extras` allows to add optional information or override default values:
//    - `SubjectID` allows to override the ID of the user performing the action (by default it is expected in the context)
//    - `ClientID` allows to override the oauth client ID (by default it is expected in the context)
//    - `AdditionalData` allows to provide any extra information relevant for the audit log
func (l *Logger) AuditBulkRead(
	ctx context.Context,
	ownerID fmt.Stringer,
	resourceType fmt.Stringer,
	resourceIDs interface{},
	extras ...ExtraAuditInfoProvider,
) error {
	log := bulkAccessLog{
		accessLog: accessLog{
			baseAuditLog: baseAuditLog{
				Timestamp:       time.Now(),
				LogType:         Audit,
				AuditLogType:    AccessLog,
				TraceID:         getFromContext(ctx, TraceIDContextKey),
				ServiceName:     l.serviceName,
				ServiceVersion:  l.serviceVersion,
				Hostname:        l.hostname,
				PodName:         l.podName,
				Environment:     l.environment,
				ClientID:        getFromContext(ctx, ClientIDContextKey),
				RequestURL:      getFromContext(ctx, RequestURLContextKey),
				RequestDomain:   getFromContext(ctx, RequestDomainContextKey),
				CallerIPAddress: getFromContext(ctx, CallerIPContextKey),
				SubjectID:       getFromContext(ctx, UserIDContextKey),
				TenantID:        getFromContextWithDefault(ctx, TenantIDContextKey, l.tenantID),
			},
			OwnerID:      ownerID.String(),
			EventType:    Read,
			ResourceType: resourceType.String(),
		},
		ResourceIDs: resourceIDs,
	}

	for _, f := range extras {
		f(&log.accessLog)
	}

	return l.Log(log)
}
