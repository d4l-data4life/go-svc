package gormer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/d4l-data4life/go-svc/internal/testutils"
	"github.com/d4l-data4life/go-svc/pkg/db"
	"github.com/d4l-data4life/go-svc/pkg/gormer"
)

// Test gormer interface functionality using survey as examples

func TestGet(t *testing.T) {
	example1, example2, _ := testutils.InitDBWithExamples(t)
	defer db.Close()
	tests := []struct {
		name        string
		exampleName string
		expected    gormer.Example
		err         error
	}{
		{"Simple retrieve 1", example1.Name, *example1, nil},
		{"Simple retrieve 2", example2.Name, *example2, nil},
		{"Empty name", "", gormer.Example{}, gormer.ErrEmptyParams},
		{"Not found", "random", gormer.Example{}, gormer.ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &gormer.Example{Name: tt.exampleName}
			err := gormer.Get(got)

			if tt.err == nil {
				require.NoError(t, err, "Get() shouldn't return an error")
				assert.Equal(t, tt.expected.Name, got.Name)
				assert.Equal(t, tt.expected.Payload, got.Payload)
			} else {
				require.Error(t, err, "Get() should return an error")
				require.ErrorIs(t, err, tt.err, "wrong error returned")
			}
		})
	}
}

func TestUpsert(t *testing.T) {
	example1, _, _ := testutils.InitDBWithExamples(t)
	defer db.Close()
	tests := []struct {
		name    string
		example gormer.Example
		err     error
	}{
		{"Insert", gormer.Example{Name: "insert", Payload: "insert anything"}, nil},
		{"Update", gormer.Example{Name: example1.Name, Payload: "this is about chicken?"}, nil},
		{"Empty Name", gormer.Example{Name: "", Payload: "not gonna matter"}, gormer.ErrEmptyParams},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gormer.Upsert(&tt.example)

			if tt.err == nil {
				require.NoError(t, err, "Upsert() shouldn't return an error")
				got := &gormer.Example{Name: tt.example.Name, Payload: tt.example.Payload}
				err := gormer.Get(got)

				require.NoError(t, err, "Get() shouldn't return an error")
				assert.Equal(t, tt.example.Name, got.Name)
				assert.Equal(t, tt.example.Payload, got.Payload)
			} else {
				require.ErrorIs(t, err, tt.err, "wrong error returned")
			}
		})
	}
}

func TestDelete(t *testing.T) {
	example1, example2, _ := testutils.InitDBWithExamples(t)
	defer db.Close()
	tests := []struct {
		name        string
		exampleName string
		err         error
	}{
		{"Frist delete", example1.Name, nil},
		{"Second delete", example2.Name, nil},
		{"Empty name", "", gormer.ErrEmptyParams},
		{"Not found", "random", gormer.ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			survey := &gormer.Example{Name: tt.exampleName}
			err := gormer.Delete(survey)

			if tt.err == nil {
				require.NoError(t, err, "Delete() shouldn't return an error")
				getErr := gormer.Get(survey)
				require.ErrorIs(t, getErr, gormer.ErrNotFound)
			} else {
				require.ErrorIs(t, err, tt.err, "wrong error returned")
			}
		})
	}
}

func TestGetFiltered(t *testing.T) {
	example1, example2, example3 := testutils.InitDBWithExamples(t)
	wantAll := []*gormer.Example{example1, example2, example3}
	wantSingle := []*gormer.Example{example1}
	defer db.Close()
	tests := []struct {
		name        string
		exampleName string
		expected    []*gormer.Example
		err         error
	}{
		{"All", "", wantAll, nil},
		{"Single Program", example1.Name, wantSingle, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := gormer.GetFiltered(gormer.Example{Name: tt.exampleName})

			if tt.err == nil {
				require.NoError(t, err, "Get() shouldn't return an error")
				require.Len(t, got, len(tt.expected))
				for i := range tt.expected {
					assert.Equal(t, tt.expected[i].Name, got[i].Name)
				}
			} else {
				require.Error(t, err, "Get() should return an error")
				require.ErrorIs(t, err, tt.err, "wrong error returned")
			}
		})
	}
}
