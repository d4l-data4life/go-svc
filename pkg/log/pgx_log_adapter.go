package log

import (
	"context"

	"github.com/jackc/pgx/v4"
)

type PgxLogger struct {
	l           *Logger
	obfuscators []SqlObfuscator
}

func PgxLoggerAdapter(l *Logger, obfuscators ...SqlObfuscator) *PgxLogger {
	return &PgxLogger{l: l, obfuscators: obfuscators}
}

func (l *PgxLogger) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
	_ = l.l.SqlLog(ctx, level.String(), msg, data, l.obfuscators...)
}
