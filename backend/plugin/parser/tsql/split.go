package tsql

import (
	"github.com/khulnasoft/devsecdb/backend/plugin/parser/base"
	"github.com/khulnasoft/devsecdb/backend/plugin/parser/tokenizer"
	storepb "github.com/khulnasoft/devsecdb/proto/generated-go/store"
)

func init() {
	base.RegisterSplitterFunc(storepb.Engine_MSSQL, SplitSQL)
}

// SplitSQL splits the given SQL statement into multiple SQL statements.
func SplitSQL(statement string) ([]base.SingleSQL, error) {
	t := tokenizer.NewTokenizer(statement)
	list, err := t.SplitStandardMultiSQL()
	if err != nil {
		return nil, err
	}
	return list, nil
}
