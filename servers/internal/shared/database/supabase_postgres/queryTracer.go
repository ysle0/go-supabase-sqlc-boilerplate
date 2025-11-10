package supabase_postgres

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type QueryTracer struct {
	pgx.QueryTracer
	logger *slog.Logger
}

type queryTracingKey struct{}

type queryTracingValue struct {
	start time.Time
	query string
	args  []any
}

func (q *queryTracingValue) elapsed() time.Duration {
	return time.Since(q.start)
}

// TraceQueryStart is called at the beginning of Query, QueryRow, and Exec calls. The returned context is used for the
// rest of the call and will be passed to TraceQueryEnd.
func (t *QueryTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {

	return context.WithValue(ctx, queryTracingKey{}, queryTracingValue{
		start: time.Now(),
		query: data.SQL,
		args:  data.Args,
	})
}

func (t *QueryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	val, _ := ctx.Value(queryTracingKey{}).(queryTracingValue)
	elapsed := val.elapsed()

	if data.Err != nil {
		t.logger.Error("Query End ERROR>",
			"command tag", data.CommandTag.String(),
			"elapsed", elapsed,
			"error", data.Err,
		)
		return
	}

	queryTokens := strings.Split(val.query, "\n")
	query := strings.Join(queryTokens, " ")
	// query := queryTokens[0] + "\n" + strings.Join(queryTokens[1:], " ")

	t.logger.Debug("",
		// "command tag", data.CommandTag.String(),
		"elapsed", elapsed,
		"query", query,
		"args", val.args,
	)
}
