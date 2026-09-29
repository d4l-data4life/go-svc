package donortoken

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/d4l-data4life/go-svc/pkg/log"
	"github.com/d4l-data4life/go-svc/pkg/logging"
)

// Generic messages returned to callers; the detailed reason is only logged.
const (
	MsgMissingToken = "missing donor access token"
	MsgInvalidToken = "invalid donor access token"

	authorizationMetadataKey = "authorization"
	rejectedAuditMsg         = "invalid donor access token presented"
)

var errMissingToken = errors.New("no bearer token in the authorization header")

// bearerToken extracts the token of an `Authorization: Bearer <token>` header value.
func bearerToken(header string) (string, error) {
	scheme, token, found := strings.Cut(strings.TrimSpace(header), " ")
	token = strings.TrimSpace(token)
	if !found || !strings.EqualFold(scheme, "bearer") || token == "" || strings.Contains(token, " ") {
		return "", errMissingToken
	}
	return token, nil
}

// verifyAndLog verifies raw and logs rejections: expiry (expected in normal operation) at
// debug level, every other failure as a security audit event.
func verifyAndLog(ctx context.Context, v *Verifier, raw string) (Claims, error) {
	claims, err := v.Verify(ctx, raw)
	if err != nil {
		if IsExpired(err) {
			logging.LogDebugfCtx(ctx, "expired donor access token presented")
		} else {
			logging.LogAuditSecurityFailure(ctx, rejectedAuditMsg, log.AdditionalData(err.Error()))
		}
		return Claims{}, err
	}
	return claims, nil
}

// HTTPMiddleware authenticates requests with the donor access token in the Authorization
// header (`Bearer <jwt>`). On success the verified Claims are put into the request context
// (see FromContext); a missing or invalid token is answered with 401 and a generic body.
func HTTPMiddleware(v *Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			raw, err := bearerToken(r.Header.Get("Authorization"))
			if err != nil {
				logging.LogDebugfCtx(ctx, "request without donor access token")
				unauthorized(w, MsgMissingToken)
				return
			}
			claims, err := verifyAndLog(ctx, v, raw)
			if err != nil {
				unauthorized(w, MsgInvalidToken)
				return
			}
			next.ServeHTTP(w, r.WithContext(WithClaims(ctx, claims)))
		})
	}
}

func unauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	http.Error(w, msg, http.StatusUnauthorized)
}

// UnaryServerInterceptor authenticates the given gRPC methods (full names such as
// "/pkg.Service/Method"; all methods if none are given) with the donor access token in the
// `authorization` metadata. grpc-gateway maps the HTTP Authorization header to that key and
// grpc-web sends it as metadata as well, so this covers all three transports. On success the
// verified Claims are put into the context (see FromContext); a missing or invalid token
// fails with codes.Unauthenticated and a generic message. Other methods pass through.
func UnaryServerInterceptor(v *Verifier, fullMethods ...string) grpc.UnaryServerInterceptor {
	methods := make(map[string]struct{}, len(fullMethods))
	for _, m := range fullMethods {
		methods[m] = struct{}{}
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if len(methods) > 0 {
			if _, ok := methods[info.FullMethod]; !ok {
				return handler(ctx, req)
			}
		}
		var header string
		if values := metadata.ValueFromIncomingContext(ctx, authorizationMetadataKey); len(values) > 0 {
			header = values[0]
		}
		raw, err := bearerToken(header)
		if err != nil {
			logging.LogDebugfCtx(ctx, "call without donor access token")
			return nil, status.Error(codes.Unauthenticated, MsgMissingToken)
		}
		claims, err := verifyAndLog(ctx, v, raw)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, MsgInvalidToken)
		}
		return handler(WithClaims(ctx, claims), req)
	}
}
