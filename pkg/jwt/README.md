# go-jwt

Middleware for verifying JWTs.

## Description

The middleware package has a middleware for [gorilla mux](https://github.com/gorilla/mux).
It should be used for authentication and authorization of Gorilla mux handlers.

If you use a different multiplexer, feel free to create a PR for this library here.

## Claims

Claims are attributes of a JWT that satisfy the [OAuth2 RFC](https://tools.ietf.org/html/rfc6749).

### Scope

A Scope is a list of Tokens. A Token is a string that defines what a client can do on behalf of the owner.
The JWT creation authority needs to verify that the request from the client is valid and that only the allowed Scope is requested. The source of this is currently the infamous clients.json file, which is mostly maintained in 1password.
The verification of a valid Scope for performing a certain action needs to be verified on the endpoint on a Token by Token basis .

A list of Tokens that are currently in use:

```txt
    TokenPermissionsRead   = "perm:r"
    TokenPermissionsWrite  = "perm:w"
    TokenRecordsRead       = "rec:r"
    TokenRecordsWrite      = "rec:w"
    TokenRecordsAppend     = "rec:a"
    TokenAttachmentsRead   = "attachment:r"
    TokenAttachmentsWrite  = "attachment:w"
    TokenAttachmentsAppend = "attachment:a"
    TokenUserRead          = "user:r"
    TokenUserWrite         = "user:w"
    TokenUserQuery         = "user:q"
    TokenUserKeysRead      = "ku:r"
    TokenUserKeysWrite     = "ku:w"
    TokenUserKeysAppend    = "ku:a"
    TokenAppKeysRead       = "ka:r"
    TokenAppKeysWrite      = "ka:w"
    TokenAppKeysAppend     = "ka:a"
    TokenDeviceRead        = "dev:r"
    TokenDeviceWrite       = "dev:w"
    TokenDeviceAppend      = "dev:a"
    TokenUserMailVerify    = "mail:v"
    TokenTerraDB           = "terradb"
    TokenTags              = "tag:*"
```

#### Tag

A Tag is a dynamic Token. Its prefix starts with `tag:` and its postfix is a base64 encrypted Tag. The Tag is defined by the client to help it identify, search and share records and blobs. In order to enable a client to use the Tag Token, it needs to have a `tag:*` Token in its Scope.

It is very important that the content is validated as it is a potential loophole for reflected what nots.

#### Extended Token

An Extended Token is an unmanaged Token.
It is not verified by the package as a `KnownToken`.
It is meant to be used by implementors within d4l Data4Life gGmbH to enable them to create new tags without deploying new versions of `jwt` pkg to phdp-vega.

It has `ext:` as prefix and can have any alphabetical characters or a colon as postfix (in regex: `ext:[a-zA-Z:]+`). The limitation in special signs is to avoid the usage of reflected XSS or other security vulnerabilities. Example for a valid Extended Token would be `ext:labOrder:r`.

It is of utmost importance, that the values for an Extended Token are picked carefully and not misused to create dozens of new Tokens.

## How to use the middleware

Wrapper solution:

```Go
package handler

import (
	"net/http"

	jwt "github.com/gesundheitscloud/go-jwt"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
)

func Handler(cfg *Config) http.Handler {
	router := mux.NewRouter()
	auth := jwt.New(
		cfg.PublicKey, cfg.Logger,
	)

	writeAccessMiddleware := auth.Verify(
		jwt.WithGorillaOwner("owner"),
		jwt.WithAnyScopes(
			jwt.TokenRecordsWrite,
			jwt.TokenRecordsAppend,
		),
	)
	router.Methods(http.MethodPost).
		Path("/users/{owner:[A-Fa-f0-9-]+}/commonkeys").
		Handler(
			writeAccessMiddleware(cfg.DocumentPostHandler),
		)

	readAccessMiddleware := auth.Verify(
		jwt.WithGorillaOwner("owner"),
		jwt.WithAnyScopes(jwt.TokenRecordsRead),
	)
	router.Methods(http.MethodGet).
		Path("/users/{owner:[A-Fa-f0-9-]+}/commonkeys").
		Handler(
			readAccessMiddleware(cfg.DocumentPostHandler),
		)

	// And so on...

	return router
}
```
