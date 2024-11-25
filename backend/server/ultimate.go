//go:build !minidemo

package server

import (
	// Drivers.
	_ "github.com/khulnasoft/devsecdb/backend/plugin/db/bigquery"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/db/clickhouse"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/db/cockroachdb"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/db/databricks"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/db/dm"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/db/dynamodb"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/db/elasticsearch"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/db/hive"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/db/mongodb"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/db/mssql"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/db/mysql"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/db/oracle"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/db/redis"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/db/redshift"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/db/risingwave"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/db/snowflake"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/db/spanner"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/db/sqlite"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/db/starrocks"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/db/tidb"

	// Parsers.
	_ "github.com/khulnasoft/devsecdb/backend/plugin/parser/bigquery"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/parser/mysql"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/parser/partiql"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/parser/plsql"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/parser/redis"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/parser/snowflake"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/parser/standard"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/parser/tidb"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/parser/tsql"

	// Advisors.
	_ "github.com/khulnasoft/devsecdb/backend/plugin/advisor/mssql"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/advisor/mysql"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/advisor/oceanbase"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/advisor/oracle"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/advisor/snowflake"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/advisor/tidb"

	// Schema designer.
	_ "github.com/khulnasoft/devsecdb/backend/plugin/schema/mysql"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/schema/oracle"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/schema/pg"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/schema/tidb"

	// Transformers.
	_ "github.com/khulnasoft/devsecdb/backend/plugin/parser/sql/transform/mysql"

	// IM webhooks.
	_ "github.com/khulnasoft/devsecdb/backend/plugin/webhook/dingtalk"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/webhook/feishu"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/webhook/slack"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/webhook/wecom"
)
