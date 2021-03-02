package middlewares

import (
	"net/http"

	"github.com/go-chi/chi"

	"github.com/gesundheitscloud/go-svc/pkg/instrumented"
)

// NewXSRFHandler initializes a new handler
// Deprecated: Please keep this XSRF handler for backwards-compatibility until all clients have been adapted.
func NewXSRFHandler(xsrfSecret string, xsrfHeader string, handlerFactory *instrumented.HandlerFactory) *XSRFHandler {
	return &XSRFHandler{
		Handler:    handlerFactory.NewHandler("XSRFHandler"),
		HeaderName: xsrfHeader,
	}
}

//XSRFHandler is the handler responsible for Account operations
type XSRFHandler struct {
	*instrumented.Handler
	HeaderName string
}

//Routes returns the routes for the XSRFHandler
func (e *XSRFHandler) Routes() *chi.Mux {
	router := chi.NewRouter()
	router.Get(e.InstrumentChi("/", e.XSRF))
	router.Head(e.InstrumentChi("/", e.XSRF))
	return router
}

// XSRF performs XSRF for given account
func (e *XSRFHandler) XSRF(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(e.HeaderName, "deprecated")
}
