package donortoken

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A fetch by the cache's background worker goes through limitedClient as well, so it counts for
// the refresh rate limit: an unknown kid right after it must not fetch again.
func TestRefresh_BackgroundFetchCountsForRateLimit(t *testing.T) {
	var fetches atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fetches.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"keys":[]}`))
	}))
	t.Cleanup(server.Close)

	now := time.Date(2026, 9, 29, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	v, err := NewVerifier(context.Background(), server.URL, WithClock(clock), WithMinRefreshInterval(time.Minute))
	require.NoError(t, err)
	t.Cleanup(v.Close)
	require.EqualValues(t, 1, fetches.Load(), "initial fetch")

	// Past the interval of the initial fetch, the cache worker refreshes the set.
	now = now.Add(2 * time.Minute)
	res, err := limitedClient{client: http.DefaultClient, fetches: v.fetches}.Get(v.jwksURL)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())
	require.EqualValues(t, 2, fetches.Load())

	// An unknown kid within the interval of that background fetch uses the cached set.
	_, err = v.refresh(context.Background())
	require.NoError(t, err)
	assert.EqualValues(t, 2, fetches.Load(), "no refetch right after the background refresh")

	now = now.Add(time.Minute)
	_, err = v.refresh(context.Background())
	require.NoError(t, err)
	assert.EqualValues(t, 3, fetches.Load(), "refetch once the interval has passed")
}
