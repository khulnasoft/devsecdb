package server

import (
	// This includes the first-class database, Postgres.

	// Drivers.
	_ "github.com/khulnasoft/devsecdb/backend/plugin/db/pg"

	// Parsers.
	_ "github.com/khulnasoft/devsecdb/backend/plugin/parser/pg"

	// Schema designer.
	_ "github.com/khulnasoft/devsecdb/backend/plugin/schema/pg"

	// Advisors.
	_ "github.com/khulnasoft/devsecdb/backend/plugin/advisor/pg"

	// Editors.
	_ "github.com/khulnasoft/devsecdb/backend/plugin/parser/sql/engine/pg"

	// IM webhooks.
	_ "github.com/khulnasoft/devsecdb/backend/plugin/webhook/dingtalk"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/webhook/feishu"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/webhook/slack"
	_ "github.com/khulnasoft/devsecdb/backend/plugin/webhook/wecom"
)
