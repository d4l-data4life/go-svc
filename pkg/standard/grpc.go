package standard

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"google.golang.org/grpc"

	"github.com/d4l-data4life/go-svc/pkg/logging"
)

// ListenAndServeGRPC starts both a GRPC server and a RESTful gateway server
func ListenAndServeGRPC(runCtx context.Context, grpcServer *grpc.Server, listenAddress string, gwServer *http.Server) <-chan struct{} {
	serverStopped := make(chan struct{})

	logging.LogInfof("grpc server listeninig on %s", listenAddress)
	lis, err := net.Listen("tcp", listenAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// goroutine that runs the grpc server
	go func() {
		defer close(serverStopped)
		if err := grpcServer.Serve(lis); err != grpc.ErrServerStopped {
			logging.LogErrorf(err, "GRPC server error")
		}
	}()

	// goroutine that runs the gateway server
	go func() {
		defer close(serverStopped)
		if err := gwServer.ListenAndServe(); err != http.ErrServerClosed {
			logging.LogErrorf(err, "HTTP Gateway server error")
		}
	}()

	// goroutine that waits for the run context to be canceled
	go func(runCtx context.Context) {
		<-runCtx.Done()
		gracefulStopGRPC(grpcServer, gwServer)
	}(runCtx)

	return serverStopped
}

func gracefulStopGRPC(grpcServer *grpc.Server, gwServer *http.Server) {
	// give server maximally 2 seconds to handle all current requests
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// NOTE: GracefulStop will return instantly
	// when Stop it called, preventing this
	// goroutine from leaking.
	go func() {
		grpcServer.GracefulStop()
	}()

	if err := gwServer.Shutdown(shutdownCtx); err != nil {
		logging.LogErrorf(err, "shutting down HTTP server failed")
	}

	<-shutdownCtx.Done()
	grpcServer.Stop()
}
