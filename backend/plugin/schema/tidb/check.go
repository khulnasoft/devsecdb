package tidb

import (
	"fmt"

	tidbparser "github.com/khulnasoft/devsecdb/backend/plugin/parser/tidb"
	"github.com/khulnasoft/devsecdb/backend/plugin/schema"
	storepb "github.com/khulnasoft/devsecdb/proto/generated-go/store"
)

func init() {
	schema.RegisterCheckColumnType(storepb.Engine_TIDB, checkColumnType)
}

func checkColumnType(tp string) bool {
	_, err := tidbparser.ParseTiDB(fmt.Sprintf("CREATE TABLE t (a %s NOT NULL)", tp), "", "")
	return err == nil
}
