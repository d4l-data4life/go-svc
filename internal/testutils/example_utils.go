package testutils

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/d4l-data4life/go-svc/pkg/gormer"
)

func InitDBWithExamples(t *testing.T) (example1, example2, example3 *gormer.Example) {
	InitializeTestDB(t)
	example1 = &gormer.Example{Name: "chicken", Payload: "chicken is good for you"}
	err := gormer.Upsert(example1)
	require.NoError(t, err, "Error in test setup")

	example2 = &gormer.Example{Name: "cookies", Payload: "cookies are delicious"}
	err = gormer.Upsert(example2)
	require.NoError(t, err, "Error in test setup")

	example3 = &gormer.Example{Name: "dark-side", Payload: "we have cookies"}
	err = gormer.Upsert(example3)
	require.NoError(t, err, "Error in test setup")

	return
}
