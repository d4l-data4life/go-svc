package middlewares

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-on/wrap"
)

// HandleContextCancel to handle context cancel with error code 499 instead of error code 500
func HandleContextCancel(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		buf := wrap.NewBuffer(w)

		h.ServeHTTP(buf, req)

		if errors.Is(req.Context().Err(), context.Canceled) {
			buf.WriteHeader(499)
		}

		bufFlushFixed(buf)
	})
}

func bufFlushHeaders(bf *wrap.Buffer) {
	header := bf.ResponseWriter.Header()
	for k, v := range bf.Header() {
		for _, val := range v {
			header.Add(k, val)
		}
	}
}

func bufFlushFixed(bf *wrap.Buffer) {
	if bf.HasChanged() {
		bufFlushHeaders(bf)
		bf.FlushCode()
		_, _ = bf.ResponseWriter.Write(bf.Buffer.Bytes())
	}
}
