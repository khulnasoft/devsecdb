package bigquery

import (
	"context"

	"github.com/khulnasoft/devsecdb/backend/plugin/parser/base"
	storepb "github.com/khulnasoft/devsecdb/proto/generated-go/store"
)

func init() {
	base.RegisterGetQuerySpan(storepb.Engine_BIGQUERY, GetQuerySpan)
}

// GetQuerySpan returns the query span for the given statement.
func GetQuerySpan(
	ctx context.Context,
	gCtx base.GetQuerySpanContext,
	statement, database, _ string,
	// getDatabaseMetadata base.GetDatabaseMetadataFunc,
	// listDatabaseFunc base.ListDatabaseNamesFunc,
	ignoreCaseSensitive bool,
) (*base.QuerySpan, error) {
	q := newQuerySpanExtractor(database, gCtx, ignoreCaseSensitive)
	querySpan, err := q.getQuerySpan(ctx, statement)
	if err != nil {
		return nil, err
	}
	return querySpan, nil
}
