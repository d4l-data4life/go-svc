package d4lerror

type ErrorCode string

// List of ErrorCode
const (
	CredentialsInvalid             ErrorCode = "CREDENTIALS_INVALID"
	SessionNotFound                ErrorCode = "SESSION_NOT_FOUND"
	SessionInvalid                 ErrorCode = "SESSION_INVALID"
	InternalError                  ErrorCode = "INTERNAL_ERROR"
	ParameterMissing               ErrorCode = "PARAMETER_MISSING"
	ParameterInvalid               ErrorCode = "PARAMETER_INVALID"
	QueryInvalid                   ErrorCode = "QUERY_INVALID"
	BodyInvalid                    ErrorCode = "BODY_INVALID"
	TooManyRequests                ErrorCode = "TOO_MANY_REQUESTS"
)

// ErrorMessages defines verbose error messages
var ErrorMessages = map[ErrorCode]string{
	CredentialsInvalid:          "The provided credentials are invalid.",
	SessionNotFound:             "No session found in the request.",
	SessionInvalid:              "The session included in the request is invalid.",
	InternalError:               "An unexpected internal error occurred.",
	ParameterMissing:            "A parameter is missing from the request.",
	ParameterInvalid:            "A parameter from the request is malformed or wrong.",
	BodyInvalid:                 "The request body is malformed or has an unexpected format.",
	TooManyRequests:             "The client is sending requests too quickly and needs to wait.",
}

type ErrorItem struct {
	Code    ErrorCode   `json:"code"`
	TraceID string      `json:"trace_id,omitempty"`
	Message string      `json:"message,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

type ErrorResponse struct {
	Errors []ErrorItem `json:"errors,omitempty"`
}
