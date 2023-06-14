package standard

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
)

// NewGRPCGatewayServer creates a gRPC-Gateway server with given grpc handler registration functions and the given metrics handler.
func NewGRPCGatewayServer(
	server *grpc.Server,
	grpcPort, gatewayPort string,
	corsOptions cors.Options,
	handlerRegisterFunctions []func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error,
	metricsHandler runtime.HandlerFunc,
) (*http.Server, error) {
	// Adds gRPC internal logs. This is quite verbose, so adjust as desired!
	log := grpclog.NewLoggerV2(io.Discard, io.Discard, os.Stderr)
	grpclog.SetLoggerV2(log)

	// Create a client connection to the gRPC Server we just started.
	// This is where the gRPC-Gateway proxies the requests.
	conn, err := grpc.DialContext(
		context.Background(),
		"dns:///"+grpcPort,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %w", err)
	}

	gwmux := runtime.NewServeMux()
	// Register all handlers
	for _, f := range handlerRegisterFunctions {
		if err := f(context.Background(), gwmux, conn); err != nil {
			return nil, fmt.Errorf("failed to register handlers: %w", err)
		}
	}

	// Add Prometheus metrics endpoint
	err = gwmux.HandlePath("GET", "/metrics", metricsHandler)
	if err != nil {
		return nil, fmt.Errorf("failed to register metrics handler: %w", err)
	}

	// Wrap the GRPC server so that web clients can talk grpc with this server
	// see https://grpc.io/blog/state-of-grpc-web
	wrappedGrpc := grpcweb.WrapServer(server)

	listenAddress := net.JoinHostPort("", gatewayPort)
	logging.LogInfof("http gateway listeninig on %s", listenAddress)

	gwHandler := http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if wrappedGrpc.IsGrpcWebRequest(req) {
			wrappedGrpc.ServeHTTP(resp, req)
			return
		}
		gwmux.ServeHTTP(resp, req)
	})

	gwServer := &http.Server{
		Addr:              listenAddress,
		Handler:           cors.New(corsOptions).Handler(gwHandler),
		ReadHeaderTimeout: 3 * time.Second,
	}
	return gwServer, nil
}
