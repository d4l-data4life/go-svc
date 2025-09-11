package standard

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi"

	"github.com/d4l-data4life/go-svc/pkg/logging"
)

// ListenAndServe starts the server
func ListenAndServe(runCtx context.Context, mux *chi.Mux, port string) <-chan struct{} {
	serverStopped := make(chan struct{})

	listenAddress := net.JoinHostPort("", port)
	logging.LogInfof("listeninig on %s", listenAddress)
	server := &http.Server{Addr: listenAddress, Handler: mux, ReadHeaderTimeout: 0}

	// goroutine that runs the server
	go func() {
		defer close(serverStopped)
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				logging.LogInfof("HTTP server shut down")
				return
			}
			logging.LogErrorf(err, "HTTP server error")
		}
	}()

	// goroutine that waits for the run context to be canceled
	go func(runCtx context.Context) {
		<-runCtx.Done()
		gracefulStop(server)
	}(runCtx)

	return serverStopped
}

func gracefulStop(server *http.Server) {
	// give server maximally 2 seconds to handle all current requests
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logging.LogErrorf(err, "shutting down HTTP server failed")
	}
}
