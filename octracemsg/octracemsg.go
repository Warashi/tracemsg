package octracemsg

import (
	"context"
	"fmt"
	"go/constant"

	"github.com/Warashi/ssautil"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ssa"
)

const doc = "octracemsg is ..."

// Analyzer is ...
var Analyzer = &analysis.Analyzer{
	Name: "octracemsg",
	Doc:  doc,
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
		buildssa.Analyzer,
	},
}

func StartSpan(f *ssa.Function) *ssa.Call {
	for _, b := range f.Blocks {
		for _, instr := range b.Instrs {
			switch instr := instr.(type) {
			case *ssa.Call:
				if instr.Common().Value.Name() == "StartSpan" {
					return instr
				}
			}
		}
	}
	return nil
}

func IsTarget(f *ssa.Function) bool {
	if !ssautil.IsExported(f) {
		return false
	}
	params := f.Signature.Params()
	for i := 0; i < params.Len(); i++ {
		if ssautil.IsContext(params.At(i)) {
			return true
		}
	}
	return false
}

func report(ctx context.Context, f *ssa.Function) {
	if !IsTarget(f) {
		return
	}
	pass := ssautil.Pass(ctx)
	call := StartSpan(f)
	if call == nil {
		pass.Reportf(f.Pos(), "%s should call trace.StartSpan", f)
		return
	}
	for _, o := range ssautil.Operands(call) {
		switch o := o.(type) {
		case *ssa.Const:
			if o.Value == nil || o.Value.Kind() != constant.String {
				continue
			}
			actual := constant.StringVal(o.Value)
			want := f.String()
			if actual == want {
				return
			}
			node, ok := ssautil.CallExpr(ctx, call)
			if !ok {
				return
			}
			pos, end := node.Pos(), node.End()
			pass.Report(analysis.Diagnostic{
				Pos:     pos,
				End:     end,
				Message: fmt.Sprintf("span name should be %q", want),
				SuggestedFixes: []analysis.SuggestedFix{{TextEdits: []analysis.TextEdit{{
					Pos:     pos,
					End:     end,
					NewText: ssautil.PrettyPrint(ctx, ssautil.ReplaceConstArg(node, actual, want)),
				}}}},
			})
			return
		}
	}
}

func run(pass *analysis.Pass) (interface{}, error) {
	ctx := ssautil.Prepare(context.Background(), pass)
	s := ssautil.SSA(ctx)
	for _, f := range s.SrcFuncs {
		report(ctx, f)
	}
	return nil, nil
}
