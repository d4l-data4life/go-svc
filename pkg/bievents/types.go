package bievents

type ActivityType = string

// The following values are the supported activity types for BI events.
const (
	Register          ActivityType = "register"
	EmailVerify       ActivityType = "email-verify"
	Logout            ActivityType = "logout"
	TokenRefresh      ActivityType = "token-refresh"
	DeviceRegister    ActivityType = "device-register"
	DeviceDelete      ActivityType = "device-delete"
	EIDEntrance       ActivityType = "eid-saml-entrance"
	LoginComplete     ActivityType = "login-complete"
	SharingStart      ActivityType = "sharing-start"
	SharingComplete   ActivityType = "sharing-complete"
	SharingRevoke     ActivityType = "sharing-revoke"
	DocumentDelete    ActivityType = "document-delete"
	DocumentDeleteAll ActivityType = "document-delete-all"
	DocumentUpload    ActivityType = "document-upload"
	RecordRead        ActivityType = "record-read"
	RecordBulkRead    ActivityType = "record-bulk-read"
	RecordCreate      ActivityType = "record-create"
	PasswordReset     ActivityType = "password-reset"
	AccountDelete     ActivityType = "account-delete"
	LoginEmail        ActivityType = "login-email"
	LoginSMS          ActivityType = "login-sms"
	LoginEID          ActivityType = "login-eid"

	// Login is deprecated. Use LoginEmail, LoginSMS or LoginEID instead
	Login ActivityType = "login-start"

	// PhoneVerify is deprecated. Use LoginSMS instead
	PhoneVerify ActivityType = "phone-verify"
)

// AuthnType is deprecated. Will be removed when the Login activity type is removed.
type AuthnType string

const (
	Email AuthnType = "email"
	SMS   AuthnType = "SMS"
	EID   AuthnType = "eID"
)

type LoginFailureType string

const (
	WrongEmail          LoginFailureType = "unknown-email"
	WrongPassword       LoginFailureType = "wrong-pwd"
	EmailNotValidated   LoginFailureType = "email-no-validated"
	TooManyFailedLogins LoginFailureType = "too-many-failed-logins"
	Other               LoginFailureType = "other"
)

type PasswordResetAuthnType string

const (
	EmailToken       PasswordResetAuthnType = "email-token"
	RecoveryPassword PasswordResetAuthnType = "recovery-password"
)

type UserRegisterData struct {
	CucID       string    `json:"cuc-id"`
	AccountType EmailType `json:"account-type"`
	SourceURL   string    `json:"source-url"`
	ClientID    string    `json:"client-id"`
}

type LoginEmailData struct {
	ClientID   string           `json:"client-id"`
	SourceURL  string           `json:"source-url"`
	ErrorCause LoginFailureType `json:"failure-cause,omitempty"`
}

type LoginSMSData struct {
	ClientID string `json:"client-id"`
	DeviceID string `json:"device-id"`
}

type LoginEIDData struct {
	ClientID  string `json:"client-id"`
	SourceURL string `json:"source-url"`
	Challenge string `json:"eid-challenge"`
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

type PasswordResetData struct {
	AuthenticationType PasswordResetAuthnType `json:"authn-type"`
}

type AccountDeleteData struct {
	Source string `json:"source"`
}

// LoginData is deprecated. Use LoginEmailData, LoginSMSData or LoginEIDData instead
type LoginData struct {
	AuthenticationType AuthnType `json:"authn-type"`
	ClientID           string    `json:"client-id"`
	SourceURL          string    `json:"source-url"`
	Challenge          string    `json:"eid-challenge,omitempty"` // only for authn type eID
}

// PhoneVerifyData is deprecated. Use LoginSMSData instead
type PhoneVerifyData struct {
	DeviceID string `json:"device-id"`
}
