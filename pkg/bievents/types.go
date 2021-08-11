package bievents

type ActivityType = string

// The following values are the supported activity types for BI events.
const (
	Register          ActivityType = "register"
	EmailVerify       ActivityType = "email-verify"
	Login             ActivityType = "login"
	EIDLogin          ActivityType = "eid-login"
	Logout            ActivityType = "logout"
	PhoneVerify       ActivityType = "phone-verify"
	DeviceRegister    ActivityType = "device-register"
	DeviceDelete      ActivityType = "device-delete"
	EIDEntrance       ActivityType = "eid-saml-entrance"
	LoginComplete     ActivityType = "login-complete"
	SharingStart      ActivityType = "sharing-start"
	SharingComplete   ActivityType = "sharing-complete"
	DocumentDelete    ActivityType = "document-delete"
	DocumentDeleteAll ActivityType = "document-delete-all"
	DocumentUpload    ActivityType = "document-upload"
	RecordRead        ActivityType = "record-read"
	RecordBulkRead    ActivityType = "record-bulk-read"
	RecordCreate      ActivityType = "record-create"
)

// OnboardingData is used to define details about onboarding data.
// This type can be used to define structs
type OnboardingData struct {
	CUC         string    `json:"cuc"`
	AccountType EmailType `json:"account-type"`
	SourceURL   string    `json:"source-url"`
	Source      string    `json:"source"`
}

type LoginData struct {
	ClientID string `json:"client-id"`
}

type EIDLoginData struct {
	Challenge string `json:"eid-challenge"`
	ClientID  string `json:"client-id"`
}

type LoginCompleteData struct {
	SessionID string // a session identifier: it allows to connect a login and a logout event
}

type LogoutData struct {
	SessionID string // a session identifier: it allows to connect a login and a logout event
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
	OwnerID    string `json:"owner-id"`
}

type DeviceDeleteData struct {
	DeviceID string
}

type PhoneVerifyData struct {
	DeviceID string
}

type SharingStartData struct {
	SharingSessionID string
}

type SharingCompleteData struct {
	SharingSessionID string
}
