package migrate

import (
	"context"
	"log"
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
		tt := tt
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
