package db

import (
	"strings"

	"github.com/lib/pq"

	"github.com/gesundheitscloud/go-svc/pkg/logging"
	"github.com/gesundheitscloud/go-svc/pkg/probe"
)

func handlePostgresError(pqErr *pq.Error, migFn MigrationFunc) {
	switch {
	case pqErr.Code[0:2] == "08": // connection issues
		probe.Liveness().SetDead()
	case pqErr.Code == "42P01": // relation does not exist (observed by DB restart)
		// DB has crashed and a new instance was brought up - autoconnect will connect, but migration may be necessary
		_ = migrate(Get(), migFn)
	case pqErr.Code == "53300": // sorry, too many clients already
		// the service managed to exhaust all DB connections - should restart
		probe.Liveness().SetDead()
	default:
		logging.LogInfof("unexpected pg error = '%s', code = '%s'", pqErr, pqErr.Code)
	}
}

func handleGormError(err error) {
	logging.LogDebugf("gormError '%s'", err)
	if strings.Contains(err.Error(), "connect: connection refused") {
		// service runs fine, but DB has crashed
		probe.Liveness().SetDead()
	}
}

func HandleDatabaseError(migFn MigrationFunc, dbErrors ...error) {
	for _, err := range dbErrors {
		if err != nil {
			if pqerr, ok := err.(*pq.Error); ok {
				handlePostgresError(pqerr, migFn)
			} else {
				handleGormError(err)
			}
		}
	}
}
