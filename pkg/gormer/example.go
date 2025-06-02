package gormer

import (
	"fmt"
	"time"
)

var _ Gormer = (*Example)(nil)

type Example struct {
	Name      string    `json:"name"      gorm:"primaryKey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Payload   string    `json:"payload"`
}

func (e Example) String() string {
	return fmt.Sprintf("%s - %s", e.Name, e.Payload)
}

func (Example) UpdateableColumns() []string {
	return []string{"updated_at", "payload"}
}

func (Example) ConflictClauseColumns() string {
	return "name"
}

func (Example) OrderString() string {
	return "name ASC"
}

func (e Example) Validate() error {
	if e.Name == "" {
		return ErrEmptyParams
	}
	return nil
}

func (Example) Preloads() []string {
	return nil
}
