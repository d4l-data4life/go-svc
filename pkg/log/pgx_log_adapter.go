package log

import (
	"context"

	"github.com/jackc/pgx/v4"
)

type PgxLogger struct {
	l *Logger
}

func PgxLoggerAdapter(l *Logger) *PgxLogger {
	return &PgxLogger{l: l}
}

func (l *PgxLogger) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
	_ = l.l.SqlLog(ctx, level.String(), msg, data)
}
