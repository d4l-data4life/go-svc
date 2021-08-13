package bievents

type ActivityType = string

// The following values are the supported activity types for BI events.
const (
	Register          ActivityType = "register"
	EmailVerify       ActivityType = "email-verify"
	Login             ActivityType = "login"
	EIDLogin          ActivityType = "eid-login"
	Logout            ActivityType = "logout"
	TokenRefreshed    ActivityType = "token-refreshed"
	PhoneVerify       ActivityType = "phone-verify"
	DeviceRegister    ActivityType = "device-register"
	DeviceDelete      ActivityType = "device-delete"
	EIDEntrance       ActivityType = "eid-saml-entrance"
	LoginComplete     ActivityType = "login-complete"
	SharingStart      ActivityType = "sharing-start"
	SharingComplete   ActivityType = "sharing-complete"
	SharingRevoked    ActivityType = "sharing-revoked"
	DocumentDelete    ActivityType = "document-delete"
	DocumentDeleteAll ActivityType = "document-delete-all"
	DocumentUpload    ActivityType = "document-upload"
	RecordRead        ActivityType = "record-read"
	RecordBulkRead    ActivityType = "record-bulk-read"
	RecordCreate      ActivityType = "record-create"
)

type UserRegisterData struct {
	CucID       string    `json:"cuc-id"`
	AccountType EmailType `json:"account-type"`
	SourceURL   string    `json:"source-url"`
	Source      string    `json:"source"`
}

type LoginData struct {
	ClientID  string `json:"client-id"`
	SourceURL string `json:"source-url"`
}

type EIDLoginData struct {
	Challenge string `json:"eid-challenge"`
	ClientID  string `json:"client-id"`
	SourceURL string `json:"source-url"`
}

type LoginCompleteData struct {
	SessionID string `json:"session-id"` // a session identifier: it allows to connect a login and a logout event
	ClientID  string `json:"client-id"`
}

type LogoutData struct {
	SessionID string `json:"session-id"` // a session identifier: it allows to connect a login and a logout event
}

type TokenRefreshedData struct {
	SessionID string `json:"session-id"`
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

type DeviceDeleteData struct {
	DeviceID string `json:"device-id"`
}

type PhoneVerifyData struct {
	DeviceID string `json:"device-id"`
}

type SharingStartData struct {
	SharingSessionID string `json:"sharing-session-id"`
	ClientID         string `json:"client-id"`
}

type SharingCompleteData struct {
	SharingSessionID string `json:"sharing-session-id"`
}

type SharingRevokedData struct {
	SharingSessionID string `json:"sharing-session-id"`
}
