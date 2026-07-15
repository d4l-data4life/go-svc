package bievents

type ActivityType = string

// The following values are the generic activity types shared across services.
// Services define additional, service-specific activity types (with matching
// data structs) in their own packages.
const (
	Register      ActivityType = "register"
	EmailVerify   ActivityType = "email-verify"
	LoginEmail    ActivityType = "login-email"
	LoginComplete ActivityType = "login-complete"
	Logout        ActivityType = "logout"
	TokenRefresh  ActivityType = "token-refresh"
	AccountDelete ActivityType = "account-delete"
)

type LoginFailureType string

const (
	WrongEmail          LoginFailureType = "unknown-email"
	WrongPassword       LoginFailureType = "wrong-pwd"
	EmailNotValidated   LoginFailureType = "email-no-validated"
	TooManyFailedLogins LoginFailureType = "too-many-failed-logins"
	Other               LoginFailureType = "other"
)

type LoginEmailData struct {
	ClientID   string           `json:"client-id"`
	SourceURL  string           `json:"source-url"`
	ErrorCause LoginFailureType `json:"failure-cause,omitempty"`
}

type LoginCompleteData struct {
	SessionID     string `json:"session-id"` // a session identifier: it allows to connect a login and a logout event
	ClientID      string `json:"client-id"`
	ClientVersion string `json:"client-version"`
}

type LogoutData struct {
	SessionID string `json:"session-id"` // a session identifier: it allows to connect a login and a logout event
}

type TokenRefreshedData struct {
	SessionID string `json:"session-id"`
}

type AccountDeleteData struct {
	Source string `json:"source"`
}
