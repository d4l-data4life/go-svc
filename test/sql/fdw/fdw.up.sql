BEGIN;

CREATE SERVER IF NOT EXISTS keymgmt_server FOREIGN DATA WRAPPER postgres_fdw OPTIONS (host '{{.Hostname}}', dbname '{{.DBName}}', port '{{.Port}}');

CREATE USER MAPPING FOR {{.LocalUser}} SERVER keymgmt_server OPTIONS (user '{{.User}}', password '{{.Password}}');

COMMIT;
