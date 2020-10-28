package d4lcontext

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestParseRequesterID(t *testing.T) {
	id := "123e4567-e89b-12d3-a456-426655440000"
	reqID, _ := uuid.FromString(id)
	tests := []struct {
		name        string
		requesterID interface{}
		expectedID  uuid.UUID
		expectedErr string
	}{
		{"RequesterID as string", id, reqID, ""},
		{"RequesterID as uuid", reqID, reqID, ""},
		{"No requestid", nil, uuid.Nil, "missing account id"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requesterID, err := ParseRequesterID(w, r)
				if err != nil {
					assert.Equal(t, tt.expectedErr, err.Error())
				}
				assert.Equal(t, tt.expectedID, requesterID)
			})

			req, _ := http.NewRequest(http.MethodGet, "/route", nil)
			ctx := context.WithValue(req.Context(), UserIDContextKey, tt.requesterID)
			req = req.WithContext(ctx)

			res := httptest.NewRecorder()
			handler.ServeHTTP(res, req)

		})
	}
}
