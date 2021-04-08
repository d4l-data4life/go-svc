# go-log

A package for logging.

## Usage

The package entrypoint is the Logger convenience  type.

```Go
func getUserIDFromRequest(req *http.Request) string {
	userID, _ := authenticator.GetUserID(req.Header.Get("Authorization"))
	return userID
}

func main() {
	logger := log.NewLogger(
		os.Getenv("APP_NAME"),
		os.Getenv("APP_VERSION"),
		os.Getenv("HOSTNAME"),
	)

	logger.ServiceStart()
	defer logger.ServiceStop()

	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		err := DoBusinessLogic()
		if err != nil {
			// The request Context contains the `trace-id` and the
			// `user-id` required for error logging.
			logger.ErrGeneric(req.Context(), err)
			return
		}
		logger.InfoGeneric(req.Context(), "Add some random info here.")

		rw.WriteHeader(http.StatusTeapot)
		rw.Write([]byte("I'm a teapot!"))
	})

	// For maximum privacy (dada donation etc.) anonymize the caller IP with IPAnonymizer
	anonymizer := log.IPAnonymizer{IPType: log.IPTypeAll, With: "0.0.0.0"},

	// WrapHTTP will log the request and the response.
	// Additionally, it will add metadata information into the
	// *http.Request context.
	loggingHandler := logger.WrapHTTP(
		handler,
		getUserIDFromRequest,
		anonymizer,
	)

	http.ListenAndServe(":8080", loggingHandler)
}
```

### Audit Methods Usage

The audit methods rely on the following information from the context:
- `user-id` (i.e. subject ID - the user performing an action). Can be overridden with `SubjectID()` builder method
- `client-id`
- `trace-id`


```Go
package main

import (
	"net/http"
	"os"

	"github.com/gesundheitscloud/go-log/v2/log"
)

type auditEvent string

// Implement the fmt.Stringer interface
func (e auditEvent) String() string { return string(e) }

// getUserIDFromJWT extracts the user ID from the request
func getUserIDFromJWT(*http.Request) string { /* code omitted */ }

// getClientIDFromJWT extracts the OAuth Client ID from the request
func getClientIDFromJWT(*http.Request) string { /* code omitted */ }

// App-specific security event that needs to be audited
const DeleteUserEvent auditEvent = "DeleteUserEvent"

func main() {
	logger := log.NewLogger(
		os.Getenv("APP_NAME"),
		os.Getenv("APP_VERSION"),
		os.Getenv("HOSTNAME"),
	)

	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		err := DeleteUser()
		if err != nil {
			// The context contains the `trace-id`, `user-id` and `client-id`
			logger.AuditSecurityFailure(
				req.Context(),
				DeleteUserEvent,
				log.AdditionalData(err),        // Use `AdditionalData()` to add more information
				log.Message("super important"), // One can also use `Message()` to log some relevant message
			)
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		logger.AuditSecuritySuccess(
			req.Context(),
			DeleteUserEvent,
		)
		rw.WriteHeader(http.StatusOK)
	})

	loggingHandler := logger.WrapHTTP(
		handler,
		log.WithUserParser(getUserIDFromJWT),
		log.WithClientIDParser(getClientIDFromJWT),
	)

	http.ListenAndServe(":8080", loggingHandler)
}
```

## Note on error catching

All the Log functions return an error value.
There are two potential unsuccessful operations that might lead to a non-nil error:
* JSON marshaling;
* writing to the io.Writer the Logger was initialised with.

The caller service might or might not be interested in catching the error.

## Test

Test with `make test`.
