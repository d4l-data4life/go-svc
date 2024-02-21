package log

import (
	"context"
)

type ChangeLogEvent string

const (
	Create ChangeLogEvent = "create"
	Update ChangeLogEvent = "update"
	Delete ChangeLogEvent = "delete"
)

type changeLog struct {
	baseAuditLog

	// owner is the user owning the resource accessed
	OwnerID string `json:"owner-id,omitempty"`

	EventType      ChangeLogEvent `json:"event-type"`
	ResourceType   string         `json:"resource-type"`
	OldValue       interface{}    `json:"value-old,omitempty"`
	NewValue       interface{}    `json:"value-new,omitempty"`
	AdditionalData interface{}    `json:"additional-data,omitempty"`
}

type singleChangeLog struct {
	changeLog
	ResourceID string `json:"resource-id"`
}

type bulkChangeLog struct {
	changeLog
	ResourceIDs interface{} `json:"resource-ids"` // use interface{} here to allow any kind of array
}

// AuditUpdate should be used to log a creation of some personal sensitive piece of information.
// It will attempt to get `TraceID`, `ClientID`, `RequestURL`, `RequestDomain`, `CallerIPAddress` and `SubjectID`
// from the context.
// `ownerID` represents the owner of the resource.
// `resourceType` should be used to specify what kind of information was accessed
// `resourceID` must include the information needed to uniquely identify the resource (ID of the resource or the data set)
// `value` must contain the created resource (secrets must be excluded)
// `extras` allows to add optional information or override default values:
//   - `SubjectID` allows to override the ID of the user performing the action (by default it is expected in the context)
//   - `ClientID` allows to override the oauth client ID (by default it is expected in the context)
//   - `AdditionalData` allows to provide any extra information relevant for the audit log
func (l *Logger) AuditCreate(
	ctx context.Context,
	ownerID string,
	resourceType string,
	resourceID string,
	value interface{},
	extras ...ExtraAuditInfoProvider,
) error {
	log := singleChangeLog{
		changeLog: changeLog{
			baseAuditLog: l.createBaseAuditLog(ctx, ChangeLog),
			OwnerID:      ownerID,
			EventType:    Create,
			ResourceType: resourceType,
			NewValue:     value,
		},
		ResourceID: resourceID,
	}

	for _, f := range extras {
		f(&log.changeLog)
	}

	return l.Log(log)
}

// AuditUpdate should be used to log an update of some personal sensitive piece of information.
// AuditDelete should be used to log a deletion of some personal sensitive piece of information.
// It will attempt to get `TraceID`, `ClientID`, `RequestURL`, `RequestDomain`, `CallerIPAddress` and `SubjectID`
// from the context.
// `ownerID` represents the owner of the resource.
// `resourceType` should be used to specify what kind of information was accessed
// `resourceID` must include the information needed to uniquely identify the resource (ID of the resource or the data set)
// `value` must contain the value after the update (secrets must be excluded)
// `extras` allows to add optional information or override default values:
//   - `SubjectID` allows to override the ID of the user performing the action (by default it is expected in the context)
//   - `ClientID` allows to override the oauth client ID (by default it is expected in the context)
//   - `OldValue` allows to provide the value before the update if available (secrets must be excluded)
//   - `AdditionalData` allows to provide any extra information relevant for the audit log
func (l *Logger) AuditUpdate(
	ctx context.Context,
	ownerID string,
	resourceType string,
	resourceID string,
	value interface{},
	extras ...ExtraAuditInfoProvider,
) error {
	log := singleChangeLog{
		changeLog: changeLog{
			baseAuditLog: l.createBaseAuditLog(ctx, ChangeLog),
			OwnerID:      ownerID,
			EventType:    Update,
			ResourceType: resourceType,
			NewValue:     value,
		},
		ResourceID: resourceID,
	}

	for _, f := range extras {
		f(&log.changeLog)
	}

	return l.Log(log)
}

// AuditDelete should be used to log a deletion of some personal sensitive piece of information.
// It will attempt to get `TraceID`, `ClientID`, `RequestURL`, `RequestDomain`, `CallerIPAddress` and `SubjectID`
// from the context.
// `ownerID` represents the owner of the resource.
// `resourceType` should be used to specify what kind of information was accessed
// `resourceID` must include the information needed to uniquely identify the resource (ID of the resource or the data set)
// `extras` allows to add optional information or override default values:
//   - `SubjectID` allows to override the ID of the user performing the action (by default it is expected in the context)
//   - `ClientID` allows to override the oauth client ID (by default it is expected in the context)
//   - `OldValue` allows to provide the value before the delete if available (secrets must be excluded)
//   - `AdditionalData` allows to provide any extra information relevant for the audit log
func (l *Logger) AuditDelete(
	ctx context.Context,
	ownerID string,
	resourceType string,
	resourceID string,
	extras ...ExtraAuditInfoProvider,
) error {
	log := singleChangeLog{
		changeLog: changeLog{
			baseAuditLog: l.createBaseAuditLog(ctx, ChangeLog),
			OwnerID:      ownerID,
			EventType:    Delete,
			ResourceType: resourceType,
		},
		ResourceID: resourceID,
	}

	for _, f := range extras {
		f(&log.changeLog)
	}

	return l.Log(log)
}

// AuditBulkDelete should be used to log a bulk deletion of data containing personal information.
// It will attempt to get `TraceID`, `ClientID`, `RequestURL`, `RequestDomain`, `CallerIPAddress` and `SubjectID`
// from the context.
// `ownerID` represents the owner of the resource.
// `resourceType` should be used to specify what kind of information was accessed
// `resourceIDs` must include the information needed to uniquely identify all the resources deleted
// `extras` allows to add optional information or override default values:
//   - `SubjectID` allows to override the ID of the user performing the action (by default it is expected in the context)
//   - `ClientID` allows to override the oauth client ID (by default it is expected in the context)
//   - `OldValue` allows to provide the value before the delete if available (secrets must be excluded)
//   - `AdditionalData` allows to provide any extra information relevant for the audit log
func (l *Logger) AuditBulkDelete(
	ctx context.Context,
	ownerID string,
	resourceType string,
	resourceIDs interface{},
	extras ...ExtraAuditInfoProvider,
) error {
	log := bulkChangeLog{
		changeLog: changeLog{
			baseAuditLog: l.createBaseAuditLog(ctx, ChangeLog),
			OwnerID:      ownerID,
			EventType:    Delete,
			ResourceType: resourceType,
		},
		ResourceIDs: resourceIDs,
	}

	for _, f := range extras {
		f(&log.changeLog)
	}

	return l.Log(log)
}
