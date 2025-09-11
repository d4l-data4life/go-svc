package gormer

import (
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/d4l-data4life/go-svc/pkg/db"
	"github.com/d4l-data4life/go-svc/pkg/logging"
)

// define errors
var (
	ErrEmptyParams = errors.New("required parameter(s) empty")
	ErrNotFound    = errors.New("resource not found")
	ErrGet         = errors.New("failed retrieving resource")
	ErrGetFilered  = errors.New("failed retrieving filtered resources")
	ErrUpsert      = errors.New("failed upserting resource")
	ErrDelete      = errors.New("failed deleting resource")
)

type Gormer interface {
	Validate() error
	UpdateableColumns() []string
	ConflictClauseColumns() string
	OrderString() string
	Preloads() []string
}

// Get fetches a resource from the database (always use pointers)
func Get[T Gormer](g *T) error {
	if err := (*g).Validate(); err != nil {
		logging.LogErrorf(ErrEmptyParams, "cannot get resource")
		return err
	}

	query := db.Get().Where(g)
	for _, el := range (*g).Preloads() {
		query = query.Preload(el)
	}

	err := query.Take(g).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrNotFound
		}
		logging.LogErrorf(err, "%s", ErrGet.Error())
		return ErrGet
	}

	return nil
}

func GetFiltered[T Gormer](g T) (result []T, err error) {
	err = db.Get().Order(g.OrderString()).Where(&g).Find(&result).Error
	if err != nil {
		logging.LogErrorf(err, "%s", ErrGetFilered.Error())
		return nil, ErrGetFilered
	}

	return result, nil
}

// Upsert creates/updates a resource in the database (always use pointers)
func Upsert[T Gormer](g *T) error {
	if err := (*g).Validate(); err != nil {
		logging.LogErrorf(ErrEmptyParams, "cannot upsert resource")
		return err
	}

	err := db.Get().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: (*g).ConflictClauseColumns(), Raw: true}},
		DoUpdates: clause.AssignmentColumns((*g).UpdateableColumns())}).
		Create(g).Error

	if err != nil {
		logging.LogErrorf(err, "%s", ErrUpsert.Error())
		return ErrUpsert
	}

	return nil
}

// Delete deletes resource from the database (always use pointers)
func Delete[T Gormer](g *T) error {
	if err := (*g).Validate(); err != nil {
		logging.LogErrorf(ErrEmptyParams, "cannot delete resource")
		return err
	}

	result := db.Get().Delete(g)

	if result.Error != nil {
		logging.LogErrorf(result.Error, "%s", ErrDelete.Error())
		return ErrDelete
	}

	if result.RowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
