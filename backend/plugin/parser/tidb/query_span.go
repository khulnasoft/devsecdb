package tidb

import (
	"context"

	"github.com/pkg/errors"

	"github.com/khulnasoft/devsecdb/backend/plugin/parser/base"
	storepb "github.com/khulnasoft/devsecdb/proto/generated-go/store"
)

func init() {
	base.RegisterGetQuerySpan(storepb.Engine_TIDB, GetQuerySpan)
}

func GetQuerySpan(ctx context.Context, gCtx base.GetQuerySpanContext, statement, database, _ string, _ bool) (*base.QuerySpan, error) {
	extractor := newQuerySpanExtractor(database, gCtx)

	querySpan, err := extractor.getQuerySpan(ctx, statement)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get query span from statement: %s", statement)
	}
	return querySpan, nil
}
