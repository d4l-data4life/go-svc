package db

import (
	"errors"
	"testing"

	"github.com/d4l-data4life/go-svc/pkg/migrate"
)

type fakeVersionSetter struct {
	version  uint
	dirty    bool
	err      error
	forced   []uint
	forceErr error
}

func (f *fakeVersionSetter) Version() (uint, bool, error) {
	return f.version, f.dirty, f.err
}

func (f *fakeVersionSetter) Force(v int) error {
	f.forced = append(f.forced, uint(v))
	return f.forceErr
}

func TestCurrentMigrationVersion_StartFromZero(t *testing.T) {
	setter := &fakeVersionSetter{err: migrate.ErrNilVersion}
	version, dirty, err := currentMigrationVersion(setter, 5, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dirty {
		t.Fatalf("expected dirty=false")
	}
	if version != 0 {
		t.Fatalf("expected version 0, got %d", version)
	}
	if len(setter.forced) != 0 {
		t.Fatalf("expected no Force calls, got %v", setter.forced)
	}
}

func TestCurrentMigrationVersion_RecordsTarget(t *testing.T) {
	setter := &fakeVersionSetter{err: migrate.ErrNilVersion}
	version, dirty, err := currentMigrationVersion(setter, 7, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dirty {
		t.Fatalf("expected dirty=false")
	}
	if version != 7 {
		t.Fatalf("expected version 7, got %d", version)
	}
	if len(setter.forced) != 1 || setter.forced[0] != 7 {
		t.Fatalf("expected Force(7), got %v", setter.forced)
	}
}

func TestCurrentMigrationVersion_PropagatesDirty(t *testing.T) {
	setter := &fakeVersionSetter{version: 3, dirty: true}
	version, dirty, err := currentMigrationVersion(setter, 7, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !dirty {
		t.Fatalf("expected dirty=true")
	}
	if version != 3 {
		t.Fatalf("expected version 3, got %d", version)
	}
}

func TestCurrentMigrationVersion_PropagatesError(t *testing.T) {
	boom := errors.New("boom")
	setter := &fakeVersionSetter{err: boom}
	_, _, err := currentMigrationVersion(setter, 7, false)
	if !errors.Is(err, boom) {
		t.Fatalf("expected error %v, got %v", boom, err)
	}
}
