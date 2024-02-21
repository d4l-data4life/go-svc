package log

import (
	"context"
)

type AccessLogEvent string

const (
	Read AccessLogEvent = "read"
)

type accessLog struct {
	baseAuditLog

	// owner is the user owning the resource accessed
	OwnerID        string         `json:"owner-id,omitempty"`
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
//   - `SubjectID` allows to override the ID of the user performing the action (by default it is expected in the context)
//   - `AdditionalData` allows to provide any extra information relevant for the audit log
func (l *Logger) AuditRead(
	ctx context.Context,
	ownerID string,
	resourceType string,
	resourceID string,
	extras ...ExtraAuditInfoProvider,
) error {
	log := singleAccessLog{
		accessLog: accessLog{
			baseAuditLog: l.createBaseAuditLog(ctx, AccessLog),
			OwnerID:      ownerID,
			EventType:    Read,
			ResourceType: resourceType,
		},
		ResourceID: resourceID,
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
//   - `SubjectID` allows to override the ID of the user performing the action (by default it is expected in the context)
//   - `ClientID` allows to override the oauth client ID (by default it is expected in the context)
//   - `AdditionalData` allows to provide any extra information relevant for the audit log
func (l *Logger) AuditBulkRead(
	ctx context.Context,
	ownerID string,
	resourceType string,
	resourceIDs []string,
	extras ...ExtraAuditInfoProvider,
) error {
	log := bulkAccessLog{
		accessLog: accessLog{
			baseAuditLog: l.createBaseAuditLog(ctx, AccessLog),
			OwnerID:      ownerID,
			EventType:    Read,
			ResourceType: resourceType,
		},
		ResourceIDs: resourceIDs,
	}

	for _, f := range extras {
		f(&log.accessLog)
	}

	return l.Log(log)
}
