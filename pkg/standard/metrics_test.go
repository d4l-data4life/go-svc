package standard

import (
	"context"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

func freePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to find a free port: %v", err)
	}
	defer l.Close()
	return strconv.Itoa(l.Addr().(*net.TCPAddr).Port)
}

func get(t *testing.T, url string) (int, string) {
	t.Helper()
	var lastErr error
	for range 50 {
		resp, err := http.Get(url) //nolint:gosec,noctx
		if err == nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			return resp.StatusCode, string(body)
		}
		lastErr = err
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("server didn't answer: %v", lastErr)
	return 0, ""
}

func TestListenAndServeMetrics(t *testing.T) {
	port := freePort(t)
	ctx, cancel := context.WithCancel(context.Background())
	stopped := ListenAndServeMetrics(ctx, port)

	code, body := get(t, "http://localhost:"+port+"/metrics")
	if code != http.StatusOK || !strings.Contains(body, "go_goroutines") {
		t.Fatalf("expected Prometheus metrics, got %d", code)
	}

	if code, _ := get(t, "http://localhost:"+port+"/other"); code != http.StatusNotFound {
		t.Fatalf("expected 404 for other paths, got %d", code)
	}

	cancel()
	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		t.Fatal("server didn't stop after the context was canceled")
	}
}
