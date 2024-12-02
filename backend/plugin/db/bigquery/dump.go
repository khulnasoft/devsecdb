package bigquery

import (
	"context"
	"io"

	storepb "github.com/khulnasoft/devsecdb/proto/generated-go/store"
)

// Dump dumps the database.
func (*Driver) Dump(_ context.Context, _ io.Writer, _ *storepb.DatabaseSchemaMetadata) error {
	return nil
}
