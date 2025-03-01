package tsql

import (
	"log/slog"
	"sort"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	parser "github.com/bytebase/tsql-parser"

	"github.com/khulnasoft/devsecdb/backend/plugin/parser/base"
	storepb "github.com/khulnasoft/devsecdb/proto/generated-go/store"
)

func init() {
	base.RegisterExtractResourceListFunc(storepb.Engine_MSSQL, ExtractResourceList)
	base.RegisterQueryValidator(storepb.Engine_MSSQL, ValidateSQLForEditor)
}

func ValidateSQLForEditor(statement string) (bool, bool, error) {
	parseResult, err := ParseTSQL(statement)
	if err != nil {
		return false, false, err
	}
	if parseResult == nil {
		return false, false, nil
	}

	l := &queryValidateListener{
		valid: true,
	}
	antlr.ParseTreeWalkerDefault.Walk(l, parseResult.Tree)
	return l.valid, l.valid, nil
}

type queryValidateListener struct {
	*parser.BaseTSqlParserListener

	valid bool
}

func (q *queryValidateListener) EnterBatch_without_go(ctx *parser.Batch_without_goContext) {
	if !q.valid {
		return
	}
	if ctx.Batch_level_statement() != nil {
		q.valid = false
		return
	}
}

func (q *queryValidateListener) EnterSql_clauses(ctx *parser.Sql_clausesContext) {
	if !q.valid {
		return
	}
	if ctx.Dml_clause() == nil {
		q.valid = false
		return
	}
}

func (q *queryValidateListener) EnterDml_clause(ctx *parser.Dml_clauseContext) {
	if !q.valid {
		return
	}
	_, ok := ctx.GetParent().(*parser.Sql_clausesContext)
	if !ok {
		return
	}
	if ctx.Select_statement_standalone() == nil {
		q.valid = false
		return
	}
}

func (q *queryValidateListener) EnterSelect_statement_standalone(ctx *parser.Select_statement_standaloneContext) {
	if !q.valid {
		return
	}
	_, ok := ctx.GetParent().(*parser.Dml_clauseContext)
	if !ok {
		return
	}
	if ctx.Select_statement() == nil {
		q.valid = false
		return
	}
}

func (q *queryValidateListener) EnterQuery_specification(ctx *parser.Query_specificationContext) {
	if !q.valid {
		return
	}
	if ctx.INTO() != nil {
		// For Into clause, we only select into temporary table, likes "SELECT ... INTO #temp FROM ...".
		isValid := false
		// NOTE: normal mode is not in single session mode, so temporary table is meaningless.
		// if tableName := ctx.Table_name(); tableName != nil {
		// 	if allID := tableName.AllId_(); len(allID) == 1 {
		// 		if id := allID[0].TEMP_ID(); id != nil {
		// 			isValid = true
		// 		}
		// 	}
		// }
		q.valid = isValid
		return
	}
}

// ExtractResourceList extracts the list of resources from the SELECT statement, and normalizes the object names with the NON-EMPTY currentNormalizedDatabase and currentNormalizedSchema.
func ExtractResourceList(currentNormalizedDatabase string, currentNormalizedSchema string, selectStatement string) ([]base.SchemaResource, error) {
	parseResult, err := ParseTSQL(selectStatement)
	if err != nil {
		return nil, err
	}
	if parseResult == nil {
		return nil, nil
	}

	l := &reasourceExtractListener{
		currentDatabase: currentNormalizedDatabase,
		currentSchema:   currentNormalizedSchema,
		resourceMap:     make(map[string]base.SchemaResource),
	}

	var result []base.SchemaResource
	antlr.ParseTreeWalkerDefault.Walk(l, parseResult.Tree)
	for _, resource := range l.resourceMap {
		result = append(result, resource)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].String() < result[j].String()
	})

	return result, nil
}

type reasourceExtractListener struct {
	*parser.BaseTSqlParserListener

	currentDatabase string
	currentSchema   string
	resourceMap     map[string]base.SchemaResource
}

// EnterTable_source_item is called when the parser enters the table_source_item production.
func (l *reasourceExtractListener) EnterTable_source_item(ctx *parser.Table_source_itemContext) {
	if fullTableName := ctx.Full_table_name(); fullTableName != nil {
		name, err := NormalizeFullTableName(fullTableName)
		if err != nil {
			slog.Debug("Failed to normalize full table name", "error", err)
			return
		}
		var parts []string
		var linkedServer string
		if name.LinkedServer != "" {
			linkedServer = name.LinkedServer
			parts = append(parts, linkedServer)
		}

		database := l.currentDatabase
		if name.Database != "" {
			database = name.Database
		}
		parts = append(parts, database)

		schema := l.currentSchema
		if name.Schema != "" {
			schema = name.Schema
		}
		parts = append(parts, schema)

		var table string
		if name.Table != "" {
			table = name.Table
		}
		parts = append(parts, table)
		normalizedObjectName := strings.Join(parts, ".")
		l.resourceMap[normalizedObjectName] = base.SchemaResource{
			LinkedServer: linkedServer,
			Database:     database,
			Schema:       schema,
			Table:        table,
		}
	}

	if rowsetFunction := ctx.Rowset_function(); rowsetFunction != nil {
		return
	}

	// https://simonlearningsqlserver.wordpress.com/tag/changetable/
	// It seems that the CHANGETABLE is only return some statistics, so we ignore it.
	if changeTable := ctx.Change_table(); changeTable != nil {
		return
	}

	// other...
}
