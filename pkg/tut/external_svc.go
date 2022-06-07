package tut

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi"
)

// EndpointMock represents the mock and the expectations for one endpoint.
// For an endpoint, the following options can be specified:
// RequestChecks: how the request is expected to look. This allows to check that the request is formed
// as expected. Optional, defaults to no checks
// ResponseBuilder: how the response should look like. Optional, defaults to a successful (status 200) empty response.
// ProcessingTime: how long the processing should take (until the handler returns). Optional, defaults to 0
// MultipleCalls: if this is set to true, the mock will be used for any number of calls to the same endpoint.
// In case it is false, it will only be used once. Optional, defaults to 'false'.
type EndpointMock struct {
	reqError error

	RequestChecks   RequestCheckFunc
	ResponseBuilder ResponseBuilder
	ProcessingTime  time.Duration
	MultipleCalls   bool
}

// ExternalServiceMock represents the mock of an external HTTP service that the
// test under code is expected to interact with.
type ExternalServiceMock struct {
	// unmetExpectations will hold the endpoints calls that are still expected to be made
	// this is a map from the endpoint URL to the (potentially multiple) endpoint mock(s)
	unmetExpectations map[string][]*EndpointMock

	// metExpectations will hold the endpoints calls that were made and could be mapped to an expected call
	// This doesn't necessarily mean that the checks were successful, just that the call was expected and happened.
	metExpectations map[string][]*EndpointMock

	unexpectedCalls []error

	// mutex protects both the expectations map and the unexpectedCalls array
	mutex sync.Mutex
}

// NewExternalService returns a new instance of a ExternalServiceMock.
// By default the mock doesn't expect any calls and returns 404.
// Expected endpoints calls can be added using .On
func NewExternalService() ExternalServiceMock {
	return ExternalServiceMock{
		unmetExpectations: make(map[string][]*EndpointMock),
		metExpectations:   make(map[string][]*EndpointMock),
		unexpectedCalls:   make([]error, 0),
	}
}

// On allows to add endpoints that are expected to be called by the code under test.
// Each endpoint is identified by a chi-compatible URL string. This will be passed to a chi.Router, so
// it needs to be formatted as expected by chi.
// On needs to be called before calling .Handler to create a handler.
// On can be called multiple times if the endpoint is expected to be called multiple times. In that case, the
// calls (for the endpoint) are expected to be in the same order. Calls for multiple endpoints can be in any order.
func (m *ExternalServiceMock) On(url string, em EndpointMock) {
	if _, ok := m.unmetExpectations[url]; !ok {
		m.unmetExpectations[url] = make([]*EndpointMock, 0)
	}
	m.unmetExpectations[url] = append(m.unmetExpectations[url], &em)
}

// Handler builds a http.Handler given the behavior that was already added using the .On
// method previously.
func (m *ExternalServiceMock) Handler() http.Handler {
	r := chi.NewRouter()

	for url := range m.unmetExpectations {
		u := url // the loop variable needs to be pinned here
		r.HandleFunc(u, func(w http.ResponseWriter, r *http.Request) {
			m.mutex.Lock()
			defer m.mutex.Unlock()
			exps, ok := m.unmetExpectations[u]
			if !ok || len(exps) == 0 {
				m.unexpectedCalls = append(m.unexpectedCalls, fmt.Errorf("got an unexpected extra call to endpoint %v", u))
				return
			}

			currentExp := m.unmetExpectations[u][0]
			if currentExp.RequestChecks != nil {
				if err := currentExp.RequestChecks(r); err != nil {
					currentExp.reqError = err
				}
			}

			m.metExpectations[u] = append(m.metExpectations[u], currentExp)

			// if MultipleCalls is true, the expectation is kept in the unmet map
			if !currentExp.MultipleCalls {
				m.unmetExpectations[u] = m.unmetExpectations[u][1:]
			}

			if currentExp.ResponseBuilder != nil {
				currentExp.ResponseBuilder(w)
			}

			time.Sleep(currentExp.ProcessingTime)
		})
	}

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		m.mutex.Lock()
		m.unexpectedCalls = append(m.unexpectedCalls, fmt.Errorf("unexpected call to unknown URL: '%s'", r.URL))
		m.mutex.Unlock()
		http.Error(w, "", http.StatusNotFound)
	})

	return r
}

// AssertExpectations allows to collect the request validation errors, as well as the unexpected endpoints calls.
// It will return the first error encountered.
func (m *ExternalServiceMock) AssertExpectations() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if len(m.unexpectedCalls) > 0 {
		return fmt.Errorf("unexpected URL(s) were called: %v", m.unexpectedCalls)
	}

	// check if all the expected calls passed the checks
	for endpoint, exps := range m.metExpectations {
		for i, exp := range exps {
			if exp.reqError != nil {
				return fmt.Errorf("checks for %v-th call to endpoint '%v' failed: %v", i, endpoint, exp.reqError)
			}
		}
	}

	// check that there were no expected calls that didn't happen
	for endpoint, exps := range m.unmetExpectations {
		for _, e := range exps {
			if !e.MultipleCalls {
				// endpoint did not allow multiple calls, this means that the endpoint was never called
				return fmt.Errorf("expected at least one more call to endpoint '%v'", endpoint)
			}
			if e.MultipleCalls && len(m.metExpectations[endpoint]) == 0 {
				// endpoint allowed multiple calls but none was made
				return fmt.Errorf("expected at least one more call to endpoint '%v'", endpoint)
			}

			// else: multiple calls are allowed, at least one was made: all fine
		}
	}

	return nil
}
