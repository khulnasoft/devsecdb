package pg

import (
	"fmt"

	pgrawparser "github.com/khulnasoft/devsecdb/backend/plugin/parser/sql/engine/pg"
	"github.com/khulnasoft/devsecdb/backend/plugin/schema"
	storepb "github.com/khulnasoft/devsecdb/proto/generated-go/store"
)

func init() {
	schema.RegisterCheckColumnType(storepb.Engine_POSTGRES, checkColumnType)
}

func checkColumnType(tp string) bool {
	_, err := pgrawparser.Parse(pgrawparser.ParseContext{}, fmt.Sprintf("CREATE TABLE t (a %s NOT NULL)", tp))
	return err == nil
}
