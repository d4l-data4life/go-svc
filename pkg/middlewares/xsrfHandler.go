package middlewares

import (
	"net/http"

	"github.com/go-chi/chi"
	"golang.org/x/net/xsrftoken"

	"github.com/gesundheitscloud/go-svc/pkg/d4lcontext"
	"github.com/gesundheitscloud/go-svc/pkg/instrumented"
)

//NewXSRFHandler initializes a new handler
func NewXSRFHandler(xsrfSecret string, xsrfHeader string, handlerFactory *instrumented.HandlerFactory) *XSRFHandler {
	return &XSRFHandler{
		Handler:    handlerFactory.NewHandler("XSRFHandler"),
		secret:     xsrfSecret,
		HeaderName: xsrfHeader,
	}
}

//XSRFHandler is the handler responsible for Account operations
type XSRFHandler struct {
	*instrumented.Handler
	secret     string
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
	// Get account id from the request
	accountID, err := d4lcontext.ParseRequesterID(w, r)
	if err != nil {
		WriteHTTPErrorCode(w, err, http.StatusBadRequest)
		return
	}

	xsrfToken := xsrftoken.Generate(e.secret, accountID.String(), "")
	w.Header().Set(e.HeaderName, xsrfToken)
}
