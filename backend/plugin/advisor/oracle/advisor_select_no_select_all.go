// Package oracle is the advisor for oracle database.
package oracle

import (
	"github.com/antlr4-go/antlr/v4"
	parser "github.com/khulnasoft/go-parser/parsers/plsql"
	"github.com/pkg/errors"

	"github.com/khulnasoft/devsecdb/backend/plugin/advisor"
	storepb "github.com/khulnasoft/devsecdb/proto/generated-go/store"
)

var (
	_ advisor.Advisor = (*SelectNoSelectAllAdvisor)(nil)
)

func init() {
	advisor.Register(storepb.Engine_ORACLE, advisor.OracleNoSelectAll, &SelectNoSelectAllAdvisor{})
	advisor.Register(storepb.Engine_DM, advisor.OracleNoSelectAll, &SelectNoSelectAllAdvisor{})
	advisor.Register(storepb.Engine_OCEANBASE_ORACLE, advisor.OracleNoSelectAll, &SelectNoSelectAllAdvisor{})
}

// SelectNoSelectAllAdvisor is the advisor checking for no select all.
type SelectNoSelectAllAdvisor struct {
}

// Check checks for no select all.
func (*SelectNoSelectAllAdvisor) Check(ctx advisor.Context, _ string) ([]*storepb.Advice, error) {
	tree, ok := ctx.AST.(antlr.Tree)
	if !ok {
		return nil, errors.Errorf("failed to convert to Tree")
	}

	level, err := advisor.NewStatusBySQLReviewRuleLevel(ctx.Rule.Level)
	if err != nil {
		return nil, err
	}

	listener := &selectNoSelectAllListener{
		level:           level,
		title:           string(ctx.Rule.Type),
		currentDatabase: ctx.CurrentDatabase,
	}

	antlr.ParseTreeWalkerDefault.Walk(listener, tree)

	return listener.generateAdvice()
}

// selectNoSelectAllListener is the listener for no select all.
type selectNoSelectAllListener struct {
	*parser.BasePlSqlParserListener

	level           storepb.Advice_Status
	title           string
	currentDatabase string
	adviceList      []*storepb.Advice
}

func (l *selectNoSelectAllListener) generateAdvice() ([]*storepb.Advice, error) {
	return l.adviceList, nil
}

// EnterSelected_list is called when production selected_list is entered.
func (l *selectNoSelectAllListener) EnterSelected_list(ctx *parser.Selected_listContext) {
	if ctx.ASTERISK() != nil {
		l.adviceList = append(l.adviceList, &storepb.Advice{
			Status:  l.level,
			Code:    advisor.StatementSelectAll.Int32(),
			Title:   l.title,
			Content: "Avoid using SELECT *.",
			StartPosition: &storepb.Position{
				Line: int32(ctx.GetStart().GetLine()),
			},
		})
	}
}
