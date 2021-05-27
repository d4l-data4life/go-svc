package d4lhandler

import (
	"errors"
	"net/http"

	"github.com/gesundheitscloud/go-svc/pkg/d4lcontext"
	"github.com/gesundheitscloud/go-svc/pkg/logging"
	"github.com/gofrs/uuid"
)

// ParseRequesterID returns the requester account id from context (only for protected endpoints).
// It logs the error and adds an error to the response in case the requester ID cannot be
// found in the context.
func ParseRequesterID(w http.ResponseWriter, r *http.Request) (requesterID uuid.UUID, err error) {
	requester := r.Context().Value(d4lcontext.UserIDContextKey)
	if requester == nil {
		err := errors.New("missing account id")
		logging.LogErrorfCtx(r.Context(), err, "error parsing Requester UUID")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return uuid.Nil, err
	}

	switch id := requester.(type) {
	case string:
		requesterID, err = uuid.FromString(id)
		if err != nil || requesterID == uuid.Nil {
			err := errors.New("malformed Account ID")
			logging.LogErrorfCtx(r.Context(), err, "error parsing Requester UUID")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return uuid.Nil, err
		}
	case uuid.UUID:
		requesterID = id
	}

	return requesterID, nil
}
