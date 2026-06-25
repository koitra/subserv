package humaext

import (
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
)

func SlogMiddleware(ctx huma.Context, next func(huma.Context)) {
	method := ctx.Operation().Method
	path := ctx.Operation().Path

	next(ctx)

	status := ctx.Status()

	params := []any{
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status", status),
	}

	if status >= 400 {
		slog.Error("API request completed with error", params...)
	} else {
		slog.Info("API request completed", params...)
	}
}
