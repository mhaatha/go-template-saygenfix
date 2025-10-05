package helper

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

func CommitOrRollback(ctx context.Context, tx pgx.Tx) {
	if err := tx.Commit(ctx); err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			slog.Error("failed to rollback", "err", err)
		} else {
			slog.Error("rollback success")
		}
	}
}
