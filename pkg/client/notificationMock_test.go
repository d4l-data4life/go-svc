package client

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestTestNotfifcationMock_SendTemplatedReturningInfo(t *testing.T) {
	user1 := uuid.NewV4()
	user2 := uuid.NewV4()
	user3 := uuid.NewV4()

	type call struct {
		templateKey string
		language    string
		subscribers []uuid.UUID
	}
	tests := []struct {
		name  string
		calls []call
		want  NotifiedUsers
	}{
		{
			"Initial empty state",
			[]call{},
			map[string]map[string][]uuid.UUID{},
		},
		{
			"Single key two languages",
			[]call{
				{"key1", "en", []uuid.UUID{user1, user2}},
				{"key1", "de", []uuid.UUID{user3}},
			},
			map[string]map[string][]uuid.UUID{
				"key1": {
					"en": []uuid.UUID{user1, user2},
					"de": []uuid.UUID{user3},
				},
			},
		},
		{
			"Two keys two languages",
			[]call{
				{"key1", "en", []uuid.UUID{user1, user2}},
				{"key1", "de", []uuid.UUID{user3}},
			},
			map[string]map[string][]uuid.UUID{
				"key1": {
					"en": []uuid.UUID{user1, user2},
					"de": []uuid.UUID{user3},
				},
			},
		},
		{
			"Null conditions",
			[]call{{"key1", "", []uuid.UUID{}}},
			map[string]map[string][]uuid.UUID{
				"key1": {
					"": []uuid.UUID{},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			c := NewNotificationMock()
			for i := 0; i < len(tt.calls); i++ {
				_ = c.SendTemplated(tt.calls[i].templateKey, tt.calls[i].language, nil, tt.calls[i].subscribers...)
			}
			got := c.GetNotifiedUsers()
			assert.EqualValuesf(t, got, tt.want, "notification result should match")
		})
	}
}
