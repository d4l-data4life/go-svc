package client

import (
	"sort"

	uuid "github.com/gofrs/uuid"
)

// NotifiedUsers is a structured map that represents a tree: templateID->language->userID
type NotifiedUsers map[string]map[string][]uuid.UUID

type notifiedUsersCounter struct {
	notifiedUsers NotifiedUsers
}

func newNotifiedUsersCounter() *notifiedUsersCounter {
	return &notifiedUsersCounter{
		notifiedUsers: make(NotifiedUsers),
	}
}

func (nuc *notifiedUsersCounter) Count(templateKey, language string, subscribers ...uuid.UUID) {
	langToUsers, exists := nuc.notifiedUsers[templateKey]
	if !exists {
		langToUsers = make(map[string][]uuid.UUID)
		nuc.notifiedUsers[templateKey] = langToUsers
	}
	users, exists := langToUsers[language]
	if !exists {
		users = []uuid.UUID{}
	}
	langToUsers[language] = append(users, subscribers...)
	// sort array of user-ids for test assertions
	sorted := langToUsers[language]
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].String() < sorted[j].String()
	})
	langToUsers[language] = sorted
}
func (nuc *notifiedUsersCounter) GetStatus() NotifiedUsers {
	return nuc.notifiedUsers
}
