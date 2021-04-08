package test

import (
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gofrs/uuid"
)

type RequestBuilder func(*http.Request) error

func BuildRequest(url string, fns ...RequestBuilder) *http.Request {
	r := httptest.NewRequest("", url, nil)

	for _, fn := range fns {
		_ = fn(r)
	}

	return r
}

func WithHeader(header map[string]string) RequestBuilder {
	return func(r *http.Request) error {
		for k, v := range header {
			r.Header.Add(k, v)
		}

		return nil
	}
}

func WithAuthHeader(key *rsa.PrivateKey, owner uuid.UUID, scopes ...string) RequestBuilder {
	scope := strings.Join(scopes, " ")

	return func(r *http.Request) error {
		bt, err := BearerToken(
			key, owner, scope,
		)
		if err != nil {
			return err
		}

		r.Header.Add("Authorization", bt)

		return nil
	}
}

func WithOwnerURL(owner string) string {
	var builder strings.Builder

	builder.WriteString("http://test.data4life.care/users/")
	builder.WriteString(owner)
	builder.WriteString("/records/456")

	return builder.String()
}

func WithURLNoOwner() string {
	return "http://test.data4life.care/records"
}

func OkHandler(w http.ResponseWriter, r *http.Request) {}
