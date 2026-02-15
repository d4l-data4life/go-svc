package migrate

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type testLog struct{}

func (l *testLog) InfoGeneric(_ context.Context, msg string) error {
	log.Println(msg)
	return nil
}

func TestMigration_parseFile(t *testing.T) {
	type fields struct {
		sourceFolder string
		log          logger
	}
	type args struct {
		filename     string
		templateData interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "success - with template",
			fields: fields{
				sourceFolder: "../../test/sql/fdw",
				log:          &testLog{},
			},
			args: args{
				filename: "fdw.up.sql",
				templateData: &ForeignDatabase{
					LocalUser: "myLocalUser",
					DBName:    "myDBName",
					Hostname:  "myHostname",
					Port:      42,
					User:      "myUser",
					Password:  "myPassword",
				},
			},
			want:    wantParsedFdwUp,
			wantErr: false,
		},
		{
			name: "success - no template",
			fields: fields{
				sourceFolder: "../../test/sql/test-setup",
				log:          &testLog{},
			},
			args: args{
				filename: "1_test.up.sql",
			},
			want:    wantParsedSetup,
			wantErr: false,
		},
		{
			name: "file not exists",
			fields: fields{
				sourceFolder: "../../test/sql/fdw",
				log:          &testLog{},
			},
			args: args{
				filename: "fdw.down.sql",
				templateData: &ForeignDatabase{
					LocalUser: "myLocalUser",
					DBName:    "myDBName",
					Hostname:  "myHostname",
					Port:      42,
					User:      "myUser",
					Password:  "myPassword",
				},
			},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Migration{
				sourceFolder: tt.fields.sourceFolder,
				log:          tt.fields.log,
			}
			got, err := m.parseFile(context.Background(), tt.args.filename, tt.args.templateData)
			if (err != nil) != tt.wantErr {
				t.Errorf("Migration.parseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Migration.parseFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindBeforeUpFile(t *testing.T) {
	dir := t.TempDir()
	writeMigrationFile(t, dir, "001_init.up.sql")
	writeMigrationFile(t, dir, "002_add.before.up.sql")
	writeMigrationFile(t, dir, "002_add.before.down.sql")
	writeMigrationFile(t, dir, "003_other.before.up.sql")
	writeMigrationFile(t, dir, "004_new.before.sql")

	found, err := findBeforeUpFile(dir, 2)
	if err != nil {
		t.Fatalf("findBeforeUpFile() error = %v", err)
	}
	if found != "002_add.before.up.sql" {
		t.Fatalf("findBeforeUpFile() got = %q, want %q", found, "002_add.before.up.sql")
	}

	found, err = findBeforeUpFile(dir, 5)
	if err != nil {
		t.Fatalf("findBeforeUpFile() error = %v", err)
	}
	if found != "" {
		t.Fatalf("findBeforeUpFile() got = %q, want empty", found)
	}

	found, err = findBeforeUpFile(dir, 4)
	if err != nil {
		t.Fatalf("findBeforeUpFile() error = %v", err)
	}
	if found != "004_new.before.sql" {
		t.Fatalf("findBeforeUpFile() got = %q, want %q", found, "004_new.before.sql")
	}
}

func TestFindBeforeUpFile_Conflicts(t *testing.T) {
	dir := t.TempDir()
	writeMigrationFile(t, dir, "004_new.before.sql")
	writeMigrationFile(t, dir, "004_new.before.up.sql")

	_, err := findBeforeUpFile(dir, 4)
	if err == nil {
		t.Fatalf("expected conflict error")
	}
}

func TestFindAfterUpFile(t *testing.T) {
	dir := t.TempDir()
	writeMigrationFile(t, dir, "001_init.after.up.sql")
	writeMigrationFile(t, dir, "002_add.after.up.sql")
	writeMigrationFile(t, dir, "003_other.after.up.sql")
	writeMigrationFile(t, dir, "004_new.after.sql")

	found, err := findAfterUpFile(dir, 2)
	if err != nil {
		t.Fatalf("findAfterUpFile() error = %v", err)
	}
	if found != "002_add.after.up.sql" {
		t.Fatalf("findAfterUpFile() got = %q, want %q", found, "002_add.after.up.sql")
	}

	found, err = findAfterUpFile(dir, 5)
	if err != nil {
		t.Fatalf("findAfterUpFile() error = %v", err)
	}
	if found != "" {
		t.Fatalf("findAfterUpFile() got = %q, want empty", found)
	}

	found, err = findAfterUpFile(dir, 4)
	if err != nil {
		t.Fatalf("findAfterUpFile() error = %v", err)
	}
	if found != "004_new.after.sql" {
		t.Fatalf("findAfterUpFile() got = %q, want %q", found, "004_new.after.sql")
	}
}

func TestFindAfterUpFile_Conflicts(t *testing.T) {
	dir := t.TempDir()
	writeMigrationFile(t, dir, "004_new.after.sql")
	writeMigrationFile(t, dir, "004_new.after.up.sql")

	_, err := findAfterUpFile(dir, 4)
	if err == nil {
		t.Fatalf("expected conflict error")
	}
}

func TestCreateAfterSourceFolder_excludesBefore(t *testing.T) {
	dir := t.TempDir()
	writeMigrationFile(t, dir, "001_init.up.sql")
	writeMigrationFile(t, dir, "002_add.before.up.sql")
	writeMigrationFile(t, dir, "002_add.before.down.sql")
	writeMigrationFile(t, dir, "setup.sql")

	afterDir, cleanup, err := CreateAfterSourceFolder(dir)
	if err != nil {
		t.Fatalf("CreateAfterSourceFolder() error = %v", err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	entries, err := os.ReadDir(afterDir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	found := map[string]bool{}
	for _, entry := range entries {
		found[entry.Name()] = true
	}
	if found["002_add.before.up.sql"] || found["002_add.before.down.sql"] {
		t.Fatalf("CreateAfterSourceFolder() should exclude before files")
	}
	if !found["001_init.up.sql"] || !found["setup.sql"] {
		t.Fatalf("CreateAfterSourceFolder() should keep non-before files")
	}
}

func TestCreateAfterSourceFolderForVersion(t *testing.T) {
	dir := t.TempDir()
	writeMigrationFile(t, dir, "001_init.after.up.sql")
	writeMigrationFile(t, dir, "002_add.after.up.sql")
	writeMigrationFile(t, dir, "003_other.after.up.sql")

	afterDir, cleanup, err := CreateAfterSourceFolderForVersion(dir, 2)
	if err != nil {
		t.Fatalf("CreateAfterSourceFolderForVersion() error = %v", err)
	}
	if cleanup != nil {
		defer cleanup()
	}
	if afterDir == "" {
		t.Fatalf("CreateAfterSourceFolderForVersion() returned empty dir")
	}

	entries, err := os.ReadDir(afterDir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	found := map[string]bool{}
	for _, entry := range entries {
		found[entry.Name()] = true
	}
	if !found["002_add.up.sql"] {
		t.Fatalf("CreateAfterSourceFolderForVersion() should include renamed after file")
	}
	if found["001_init.up.sql"] || found["003_other.up.sql"] {
		t.Fatalf("CreateAfterSourceFolderForVersion() should only include the target version")
	}

	afterDir, cleanup, err = CreateAfterSourceFolderForVersion(dir, 4)
	if err != nil {
		t.Fatalf("CreateAfterSourceFolderForVersion() error = %v", err)
	}
	if cleanup != nil {
		defer cleanup()
	}
	if afterDir == "" {
		t.Fatalf("CreateAfterSourceFolderForVersion() got empty dir")
	}
	entries, err = os.ReadDir(afterDir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	found = map[string]bool{}
	for _, entry := range entries {
		found[entry.Name()] = true
	}
	if !found["4_noop.up.sql"] {
		t.Fatalf("CreateAfterSourceFolderForVersion() should create noop migration")
	}
}

func TestCreateAfterSourceFolderForVersion_afterDotSql(t *testing.T) {
	dir := t.TempDir()
	writeMigrationFile(t, dir, "004_new.after.sql")

	afterDir, cleanup, err := CreateAfterSourceFolderForVersion(dir, 4)
	if err != nil {
		t.Fatalf("CreateAfterSourceFolderForVersion() error = %v", err)
	}
	if cleanup != nil {
		defer cleanup()
	}
	entries, err := os.ReadDir(afterDir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	found := map[string]bool{}
	for _, entry := range entries {
		found[entry.Name()] = true
	}
	if !found["004_new.up.sql"] {
		t.Fatalf("CreateAfterSourceFolderForVersion() should include renamed after.sql file")
	}
}

var wantParsedFdwUp = `BEGIN;

CREATE SERVER IF NOT EXISTS keymgmt_server FOREIGN DATA WRAPPER postgres_fdw OPTIONS (host 'myHostname', dbname 'myDBName', port '42');

CREATE USER MAPPING FOR myLocalUser SERVER keymgmt_server OPTIONS (user 'myUser', password 'myPassword');

COMMIT;
`

var wantParsedSetup = `CREATE TABLE IF NOT EXISTS test_setup.testtable (
    id text NOT NULL,
    PRIMARY KEY (id)
);
`

func writeMigrationFile(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("SELECT 1;"), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
