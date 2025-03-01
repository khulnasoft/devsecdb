package mysql

import (
	"fmt"

	mysqlparser "github.com/khulnasoft/devsecdb/backend/plugin/parser/mysql"
	"github.com/khulnasoft/devsecdb/backend/plugin/schema"
	storepb "github.com/khulnasoft/devsecdb/proto/generated-go/store"
)

func init() {
	schema.RegisterCheckColumnType(storepb.Engine_MYSQL, checkColumnType)
}

func checkColumnType(tp string) bool {
	_, err := mysqlparser.ParseMySQL(fmt.Sprintf("CREATE TABLE t (a %s NOT NULL)", tp))
	return err == nil
}
