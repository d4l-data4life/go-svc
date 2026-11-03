package standard

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
)

func TestNewGRPCGatewayServerServesNoMetrics(t *testing.T) {
	gw, err := NewGRPCGatewayServer(grpc.NewServer(), "localhost:0", "0", cors.Options{},
		[]func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rec := httptest.NewRecorder()
	gw.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "http://example/metrics", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for /metrics, got %d", rec.Code)
	}
}
