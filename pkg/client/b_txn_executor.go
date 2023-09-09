package executor

import (
	batch "colexecdb/pkg/query_engine/c_batch"
	parser "colexecdb/pkg/query_engine/d_parser"
	process "colexecdb/pkg/query_engine/e_process"
	catalog "colexecdb/pkg/query_engine/f_catalog"
	queryplan "colexecdb/pkg/query_engine/g_query_plan"
	logicalplan "colexecdb/pkg/query_engine/h_logical_plan"
	"context"
)

type txnExecutor struct {
	s   *sqlExecutor
	ctx context.Context
}

func newTxnExecutor(
	ctx context.Context,
	s *sqlExecutor) (*txnExecutor, error) {
	return &txnExecutor{s: s, ctx: ctx}, nil
}

func (exec *txnExecutor) Exec(sql string) (result Result, err error) {

	// parse sql to ast statements
	stmt, err := parser.Parse(sql)
	if err != nil {
		return Result{}, err
	}

	// get table def from catalog
	schema := catalog.MockTableDef(2)
	ctx := catalog.NewMockSchemaContext()
	ctx.AppendTableDef("tbl1", schema)

	// create query plan
	qp, err := queryplan.BuildPlan(stmt, ctx)
	if err != nil {
		return Result{}, err
	}

	// TODO: implement later
	qp.Optimize(nil)

	// init logical_plan
	p := process.New(exec.ctx)
	lp := logicalplan.New(sql, exec.ctx, p, stmt)

	// compiles query plan to logical plan
	var batches []*batch.Batch
	fillFn := func(a any, bat *batch.Batch) error {
		if bat != nil {
			rows, _ := bat.Dup()
			batches = append(batches, rows)
		}
		return nil
	}
	err = lp.Compile(exec.ctx, qp, fillFn)
	if err != nil {
		return Result{}, err
	}

	// run the logical plan
	runResult, err := lp.Run()
	if err != nil {
		return Result{}, err
	}

	// set output
	result.Batches = batches
	result.AffectedRows = runResult.AffectedRows

	return
}

func (exec *txnExecutor) commit() error {
	return nil
}

func (exec *txnExecutor) rollback(err error) error {
	return nil
}