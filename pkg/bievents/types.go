package bievents

import "time"

// BaseEvent is the concrete event which is common to all the events.
// Parameters ServiceName, ServiceVersion, HostName, EventType, Timestamp are populated by library.
// Paremeters in Event are passed on from the consumer of the library.
type BaseEvent struct {
	ServiceName    string    `json:"service-name"`
	ServiceVersion string    `json:"service-version"`
	HostName       string    `json:"hostname"`
	EventType      string    `json:"event-type"`
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
	ActivityType       string      `json:"activity-type"`
	UserID             string      `json:"user-id"`
	Data               interface{} `json:"data"`
	TenantID           string      `json:"tenant-id"`
	ConsentDocumentKey string      `json:"consent-document-key"`
	State              State       `json:"state,omitempty"`
}

// OnboardingData is used to define details about onboarding data.
// This type can be used to define structs
type OnboardingData struct {
	CUC         string    `json:"cuc"`
	AccountType EmailType `json:"account-type"`
	SourceURL   string    `json:"source-url"`
	Source      string    `json:"source"`
}

type EIDLoginData struct {
	Challenge string `json:"eid-challenge"`
	ClientID  string `json:"client-id"`
}

type EIDSamlEntranceData struct {
	Usecase   string `json:"eid-usecase"`
	Country   string `json:"eid-country"`
	Challenge string `json:"eid-challenge"`
}

type DeviceRegisterData struct {
	DeviceType string `json:"device-type"`
	Challenge  string `json:"eid-challenge,omitempty"` // only for eID devices
	DeviceID   string `json:"device-id"`
}
