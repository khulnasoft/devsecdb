package mssql

import (
	"github.com/antlr4-go/antlr/v4"
	rawparser "github.com/bytebase/tsql-parser"
	"github.com/pkg/errors"

	parser "github.com/khulnasoft/devsecdb/backend/plugin/parser/tsql"
)

type stmtType = int

const (
	stmtTypeUnknown stmtType = iota
	stmtTypeResultSetGenerating
	stmtTypeResultSetRowCountGenerating
)

type stmtTypeListener struct {
	*rawparser.BaseTSqlParserListener
	stmtType stmtType
	err      error
}

func getStmtType(stmt string) (stmtType, error) {
	parseResult, err := parser.ParseTSQL(stmt)
	if err != nil {
		return stmtTypeUnknown, err
	}
	l := &stmtTypeListener{}
	antlr.ParseTreeWalkerDefault.Walk(l, parseResult.Tree)
	if l.err != nil {
		return stmtTypeUnknown, l.err
	}
	return l.stmtType, nil
}

func (l *stmtTypeListener) EnterBatch_without_go(ctx *rawparser.Batch_without_goContext) {
	switch {
	case len(ctx.AllSql_clauses()) > 0:
		if len(ctx.AllSql_clauses()) > 1 {
			l.err = errors.Errorf("unexpected multiple SQL clauses")
		}
		l.stmtType, l.err = getStmtTypeFromSQLClauses(ctx.AllSql_clauses()[0])
	case ctx.Batch_level_statement() != nil:
		l.stmtType = stmtTypeUnknown
	case ctx.Execute_body_batch() != nil:
		l.err = errors.Errorf("unsupported execute func proc")
	}
}

func getStmtTypeFromSQLClauses(ctx rawparser.ISql_clausesContext) (stmtType, error) {
	switch {
	case ctx.Dml_clause() != nil:
		if ctx.Dml_clause().Select_statement_standalone() != nil {
			return stmtTypeResultSetGenerating, nil
		}
		return stmtTypeResultSetRowCountGenerating, nil
	case ctx.Cfl_statement() != nil:
		return stmtTypeUnknown, errors.Errorf("unsupported control flow statement")
	case ctx.Another_statement() != nil:
		return stmtTypeUnknown, nil
	case ctx.Ddl_clause() != nil:
		return stmtTypeUnknown, nil
	case ctx.Dbcc_clause() != nil:
		return stmtTypeUnknown, nil
	case ctx.Backup_statement() != nil:
		return stmtTypeUnknown, nil
	}
	return stmtTypeUnknown, nil
}
