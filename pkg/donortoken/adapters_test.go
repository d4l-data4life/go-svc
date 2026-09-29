package donortoken_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/d4l-data4life/go-svc/pkg/donortoken"
)

const (
	proxyMethod = "/proto.static.Utilities/Proxy"
	otherMethod = "/proto.static.Programs/GetProgram"
)

type adapterFixture struct {
	verifier *donortoken.Verifier
	valid    string
	expired  string
	invalid  string
}

func newAdapterFixture(t *testing.T) adapterFixture {
	t.Helper()
	clock := newFakeClock()
	key := newSigningKey(t, testKid)
	stub := newDispatcherStub(t, key)
	now := clock.Now()
	return adapterFixture{
		verifier: stub.verifier(t, clock),
		valid:    validClaims(now).sign(t, jwa.ES256, key.private),
		expired:  validClaims(now).with(jwt.ExpirationKey, now.Add(-time.Hour)).sign(t, jwa.ES256, key.private),
		invalid:  validClaims(now).with(jwt.IssuerKey, "someone").sign(t, jwa.ES256, key.private),
	}
}

func TestHTTPMiddleware(t *testing.T) {
	f := newAdapterFixture(t)
	var seen *donortoken.Claims
	handler := donortoken.HTTPMiddleware(f.verifier)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if claims, ok := donortoken.FromContext(r.Context()); ok {
			seen = &claims
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	tests := []struct {
		name   string
		header string
		status int
		body   string
	}{
		{name: "valid", header: "Bearer " + f.valid, status: http.StatusNoContent},
		{name: "lower-case scheme", header: "bearer " + f.valid, status: http.StatusNoContent},
		{name: "missing header", status: http.StatusUnauthorized, body: donortoken.MsgMissingToken},
		{name: "other scheme", header: "Basic dXNlcjpwYXNz", status: http.StatusUnauthorized, body: donortoken.MsgMissingToken},
		{name: "no token", header: "Bearer ", status: http.StatusUnauthorized, body: donortoken.MsgMissingToken},
		{name: "token without scheme", header: f.valid, status: http.StatusUnauthorized, body: donortoken.MsgMissingToken},
		{name: "extra parts", header: "Bearer " + f.valid + " x", status: http.StatusUnauthorized, body: donortoken.MsgMissingToken},
		{name: "expired", header: "Bearer " + f.expired, status: http.StatusUnauthorized, body: donortoken.MsgInvalidToken},
		{name: "invalid", header: "Bearer " + f.invalid, status: http.StatusUnauthorized, body: donortoken.MsgInvalidToken},
		{name: "garbage", header: "Bearer garbage", status: http.StatusUnauthorized, body: donortoken.MsgInvalidToken},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seen = nil
			req := httptest.NewRequest(http.MethodPost, "/api/v1/studies/ecov/data", http.NoBody)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.status, rec.Code)
			if tt.status != http.StatusUnauthorized {
				require.NotNil(t, seen)
				assert.Equal(t, testAlpID, seen.AlpID)
				assert.Equal(t, testStudyID, seen.StudyID)
				return
			}
			assert.Nil(t, seen)
			assert.Equal(t, tt.body, strings.TrimSpace(rec.Body.String()))
			assert.Equal(t, "Bearer", rec.Header().Get("WWW-Authenticate"))
		})
	}
}

func TestUnaryServerInterceptor(t *testing.T) {
	f := newAdapterFixture(t)

	tests := []struct {
		name     string
		methods  []string
		method   string
		auth     []string
		wantCode codes.Code
		wantMsg  string
		claims   bool
	}{
		{name: "listed method, valid", methods: []string{proxyMethod}, method: proxyMethod,
			auth: []string{"Bearer " + f.valid}, wantCode: codes.OK, claims: true},
		{name: "listed method, no metadata", methods: []string{proxyMethod}, method: proxyMethod,
			wantCode: codes.Unauthenticated, wantMsg: donortoken.MsgMissingToken},
		{name: "listed method, empty metadata", methods: []string{proxyMethod}, method: proxyMethod, auth: []string{},
			wantCode: codes.Unauthenticated, wantMsg: donortoken.MsgMissingToken},
		{name: "listed method, not bearer", methods: []string{proxyMethod}, method: proxyMethod, auth: []string{"Basic abc"},
			wantCode: codes.Unauthenticated, wantMsg: donortoken.MsgMissingToken},
		{name: "listed method, expired", methods: []string{proxyMethod}, method: proxyMethod, auth: []string{"Bearer " + f.expired},
			wantCode: codes.Unauthenticated, wantMsg: donortoken.MsgInvalidToken},
		{name: "listed method, invalid", methods: []string{proxyMethod}, method: proxyMethod, auth: []string{"Bearer " + f.invalid},
			wantCode: codes.Unauthenticated, wantMsg: donortoken.MsgInvalidToken},
		{name: "other method passes through without token", methods: []string{proxyMethod}, method: otherMethod,
			wantCode: codes.OK},
		{name: "other method keeps no claims even with token", methods: []string{proxyMethod}, method: otherMethod,
			auth: []string{"Bearer " + f.valid}, wantCode: codes.OK},
		{name: "no methods guards everything", method: otherMethod,
			wantCode: codes.Unauthenticated, wantMsg: donortoken.MsgMissingToken},
		{name: "no methods, valid", method: otherMethod, auth: []string{"Bearer " + f.valid}, wantCode: codes.OK, claims: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := donortoken.UnaryServerInterceptor(f.verifier, tt.methods...)
			ctx := context.Background()
			if tt.auth != nil {
				md := metadata.MD{}
				for _, a := range tt.auth {
					md.Append("authorization", a)
				}
				ctx = metadata.NewIncomingContext(ctx, md)
			}
			var handlerClaims *donortoken.Claims
			handler := func(ctx context.Context, _ any) (any, error) {
				if claims, ok := donortoken.FromContext(ctx); ok {
					handlerClaims = &claims
				}
				return "ok", nil
			}

			resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: tt.method}, handler)

			if tt.wantCode != codes.OK {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
				assert.Equal(t, tt.wantMsg, st.Message())
				assert.Nil(t, resp)
				assert.Nil(t, handlerClaims)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "ok", resp)
			if tt.claims {
				require.NotNil(t, handlerClaims)
				assert.Equal(
					t,
					donortoken.Claims{AlpID: testAlpID, StudyID: testStudyID, ExpiresAt: handlerClaims.ExpiresAt},
					*handlerClaims,
				)
			} else {
				assert.Nil(t, handlerClaims)
			}
		})
	}
}
