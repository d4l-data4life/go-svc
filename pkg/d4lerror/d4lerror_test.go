package d4lerror_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	err "github.com/gesundheitscloud/go-svc/pkg/d4lerror"
	"github.com/gesundheitscloud/go-svc/pkg/log"
	"github.com/gesundheitscloud/go-svc/pkg/tut"
)

func TestErrorV2(t *testing.T) {
	someTraceID := "some-trace-id"
	type exampleDetails struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	someDetails := exampleDetails{Name: "example", Value: 1}
	for _, tc := range [...]struct {
		name           string
		request        *http.Request
		errorCode      err.ErrorCode
		statusCode     int
		details        interface{}
		responseChecks tut.ResponseCheckFunc
	}{
		{
			name: "should work without details",
			request: tut.Request(
				tut.ReqWithContextValue(log.TraceIDContextKey, someTraceID),
			),
			errorCode:   err.SessionInvalid,
			statusCode:  http.StatusUnauthorized,
			responseChecks: tut.CheckResponse(
				tut.RespHasStatusCode(http.StatusUnauthorized),
				tut.BodyContainsErrorV2(
					tut.HasErrorCode(err.SessionInvalid),
					tut.HasErrorMessage(err.ErrorMessages[err.SessionInvalid]),
					tut.HasTraceID(someTraceID),
				),
			),
		},
		{
			name: "should work with details",
			request: tut.Request(
				tut.ReqWithContextValue(log.TraceIDContextKey, someTraceID),
			),
			errorCode:   err.SessionInvalid,
			statusCode:  http.StatusUnauthorized,
			details:     someDetails,
			responseChecks: tut.CheckResponse(
				tut.RespHasStatusCode(http.StatusUnauthorized),
				tut.BodyContainsErrorV2(
					tut.HasErrorCode(err.SessionInvalid),
					tut.HasErrorMessage(err.ErrorMessages[err.SessionInvalid]),
					tut.HasTraceID(someTraceID),
					tut.HasDetails(map[string]interface{}{
						"name":  "example",
						"value": float64(1),
					}),
				),
			),
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			callErr := err.ErrorV2(rec, tc.request, tc.errorCode, err.ErrorMessages[tc.errorCode], tc.statusCode, err.WithDetails(tc.details))
			if callErr != nil {
				t.Error(callErr)
			}
			res := rec.Result()
			defer res.Body.Close()
			if err := tc.responseChecks(res); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestInternalErrorResponse(t *testing.T) {
	for _, tc := range [...]struct {
		name    string
		traceID string
		want    err.ErrorResponse
	}{
		{
			name:    "happy path",
			traceID: "some-trace-id",
			want: err.ErrorResponse{
				Errors: []err.ErrorItem{
					{
						Code:    err.InternalError,
						TraceID: "some-trace-id",
						Message: err.ErrorMessages[err.InternalError],
					},
				},
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			have := err.InternalErrorResponse(tc.traceID)
			if !reflect.DeepEqual(tc.want, have) {
				t.Error(fmt.Errorf("expected to get %v, got %v", tc.want, have))
			}
		})
	}
}
