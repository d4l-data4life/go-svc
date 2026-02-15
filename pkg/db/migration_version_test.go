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
	forced   []int
	forceErr error
}

func (f *fakeVersionSetter) Version() (uint, bool, error) {
	return f.version, f.dirty, f.err
}

func (f *fakeVersionSetter) Force(v int) error {
	f.forced = append(f.forced, v)
	return f.forceErr
}

func TestCurrentMigrationVersion_StartFromZero(t *testing.T) {
	setter := &fakeVersionSetter{err: migrate.ErrNilVersion}
	version, dirty, needsRecordTarget, err := currentMigrationVersion(setter, 5, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dirty {
		t.Fatalf("expected dirty=false")
	}
	if needsRecordTarget {
		t.Fatalf("expected needsRecordTarget=false")
	}
	if version != 0 {
		t.Fatalf("expected version 0, got %d", version)
	}
	if len(setter.forced) != 0 {
		t.Fatalf("expected no Force calls, got %v", setter.forced)
	}
}

func TestCurrentMigrationVersion_NeedsRecordTarget(t *testing.T) {
	setter := &fakeVersionSetter{err: migrate.ErrNilVersion}
	version, dirty, needsRecordTarget, err := currentMigrationVersion(setter, 7, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dirty {
		t.Fatalf("expected dirty=false")
	}
	if !needsRecordTarget {
		t.Fatalf("expected needsRecordTarget=true")
	}
	if version != 7 {
		t.Fatalf("expected version 7, got %d", version)
	}
	if len(setter.forced) != 0 {
		t.Fatalf("expected no Force calls, got %v", setter.forced)
	}
}

func TestCurrentMigrationVersion_PropagatesDirty(t *testing.T) {
	setter := &fakeVersionSetter{version: 3, dirty: true}
	version, dirty, needsRecordTarget, err := currentMigrationVersion(setter, 7, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !dirty {
		t.Fatalf("expected dirty=true")
	}
	if needsRecordTarget {
		t.Fatalf("expected needsRecordTarget=false")
	}
	if version != 3 {
		t.Fatalf("expected version 3, got %d", version)
	}
}

func TestCurrentMigrationVersion_PropagatesError(t *testing.T) {
	boom := errors.New("boom")
	setter := &fakeVersionSetter{err: boom}
	version, dirty, needsRecordTarget, err := currentMigrationVersion(setter, 7, false)
	_ = version
	_ = dirty
	_ = needsRecordTarget
	if !errors.Is(err, boom) {
		t.Fatalf("expected error %v, got %v", boom, err)
	}
}
