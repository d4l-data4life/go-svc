package test

import (
	"net/http"

	"github.com/gorilla/mux"
)

type HandlerGorillaBuilder func(*mux.Router)

func BuildGorillaHandler(builder ...HandlerGorillaBuilder) http.Handler {
	h := mux.NewRouter()

	for _, build := range builder {
		build(h)
	}

	return h
}

func WithGorillaMiddleware(mw func(http.Handler) http.Handler) HandlerGorillaBuilder {
	return func(mx *mux.Router) {
		mx.Use(mux.MiddlewareFunc(mw))
	}
}

func WithGorillaHandler(path string, h http.Handler) HandlerGorillaBuilder {
	return func(mx *mux.Router) {
		mx.Handle(path, h)
	}
}
