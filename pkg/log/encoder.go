package log

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// Encoder transforms a log entry into the wanted format
// Interface will change when introducing proper log entry type
type Encoder interface {
	Encode(v interface{}) error
}

// NewJSONEncoder creates a new JSON encoder
func NewJSONEncoder(w io.Writer) Encoder {
	return json.NewEncoder(w)
}

// PrettyEncoder transforms log entries to a readable format
type PrettyEncoder struct {
	out io.Writer
}

const timeFmt = time.StampMilli

// Encode transforms the given value and writes it to the configured io.Writer
func (e *PrettyEncoder) Encode(entry interface{}) error {
	var s string

	switch v := entry.(type) {
	case singleChangeLog:
		s = fmt.Sprintf("%s %s %s %s | traceID: %s; subjectID: %s; clientID: %s; old: %s; new: %s; resourceID: %s resourceType: %s",
			v.Timestamp.Format(timeFmt), v.LogType, v.AuditLogType, v.RequestURL,
			v.TraceID, v.SubjectID, v.ClientID, v.OldValue, v.NewValue, v.ResourceID, v.ResourceType)
	case bulkChangeLog:
		s = fmt.Sprintf("%s %s %s %s | traceID: %s; subjectID: %s; clientID: %s; old: %s; new: %s; resourceIDs: %s, resourceTypes: %s",
			v.Timestamp.Format(timeFmt), v.LogType, v.AuditLogType, v.RequestURL,
			v.TraceID, v.SubjectID, v.ClientID, v.OldValue, v.NewValue, v.ResourceIDs, v.ResourceType)
	case singleAccessLog:
		s = fmt.Sprintf("%s %s %s %s | traceID: %s; subjectID: %s; clientID: %s; resourceID: %s; resourceType: %s",
			v.Timestamp.Format(timeFmt), v.LogType, v.AuditLogType, v.RequestURL, v.TraceID, v.SubjectID, v.ClientID, v.ResourceID, v.ResourceType)
	case securityLog:
		s = fmt.Sprintf("%s %s %s %s | traceID: %s; subjectID: %s; clientID: %s; securityEvent: %s successful: %t",
			v.Timestamp.Format(timeFmt), v.LogType, v.AuditLogType, v.RequestURL, v.TraceID, v.SubjectID, v.ClientID, v.SecurityEvent, v.Successful)
	case logEntry:
		s = fmt.Sprintf("%s %s %s | traceID: %s; userID: %s | %s",
			v.Timestamp.Format(timeFmt), v.LogLevel, v.EventType, v.TraceID, v.UserID, v.Message)
		if v.Error != "" {
			s += fmt.Sprintf(" err: %s", v.Error)
		}
	case inRequestLog:
		s = fmt.Sprintf("%s %s %s | traceID: %s; userID: %s; method: %s; url: %s | %s",
			v.Timestamp.Format(timeFmt), v.LogLevel, v.EventType, v.TraceID, v.UserID, v.ReqMethod, v.ReqURL, v.ReqBody)
	case inResponseLog:
		s = fmt.Sprintf("%s %s %s | traceID: %s; userID: %s; url: %s; code: %d; roundtrip-duration: %d | %s",
			v.Timestamp.Format(timeFmt), v.LogLevel, v.EventType, v.TraceID, v.UserID, v.ReqURL, v.ResponseCode, v.Duration, v.ResponseBody)
	case outRequestLog:
		s = fmt.Sprintf("%s %s %s | traceID: %s; userID: %s; method: %s; url: %s | %s",
			v.Timestamp.Format(timeFmt), v.LogLevel, v.EventType, v.TraceID, v.UserID, v.ReqMethod, v.ReqURL, v.ReqBody)
	case outResponseLog:
		s = fmt.Sprintf("%s %s %s | traceID: %s; userID: %s; url: %s; code: %d; roundtrip-duration: %d | %s",
			v.Timestamp.Format(timeFmt), v.LogLevel, v.EventType, v.TraceID, v.UserID, v.ReqURL, v.ResponseCode, v.Duration, v.ResponseBody)
	case sqlLogEntry:
		s = fmt.Sprintf("%s %s %s | traceId: %s; userID: %s | action: %s; Duration: %d",
			v.Timestamp.Format(timeFmt), v.LogLevel, v.EventType, v.TraceID, v.UserID, v.Action, v.Duration)

		if v.Error != "" {
			s += fmt.Sprintf(" err: %s", v.Error)
		}

		if v.Sql != "" {
			s += fmt.Sprintf(" sql: %s", v.Sql)
		}

		if v.Args != "" {
			s += fmt.Sprintf(" args: %s", v.Args)
		}

	default:
		return fmt.Errorf("unknown log type: %T", v)
	}

	s += "\n"

	_, err := e.out.Write([]byte(s))
	if err != nil {
		return fmt.Errorf("writing log failed: %w", err)
	}

	return nil
}

// NewPrettyEncoder creates a new pretty encoder
// To be used to local development and tests
func NewPrettyEncoder(w io.Writer) Encoder {
	return &PrettyEncoder{out: w}
}
