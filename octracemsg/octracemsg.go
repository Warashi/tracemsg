package octracemsg

import (
	"context"
	"fmt"
	"go/ast"
	"go/constant"
	"go/types"
	"strings"

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

func IsTarget(ctx context.Context, f *ssa.Function) bool {
	file, ok := ssautil.Node[*ast.File](ctx, f)
	if !ok || ssautil.IsGenerated(file) {
		return false
	}
	if f.Object() == nil || !f.Object().Exported() {
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

func Name(typ types.Type) string {
	switch typ := typ.(type) {
	case *types.Pointer:
		return Name(typ.Elem())
	case *types.Named:
		return typ.Obj().Name()
	}
	panic(fmt.Errorf("unknown type: %T", typ))
}

func want(f *ssa.Function) string {
	var builder strings.Builder
	builder.WriteString(f.Package().Pkg.Name())
	builder.WriteString(".")
	if recv := f.Signature.Recv(); recv != nil {
		builder.WriteString(Name(recv.Type()))
		builder.WriteString("#")
	}
	builder.WriteString(f.Name())
	return builder.String()
}

func report(ctx context.Context, f *ssa.Function) {
	if !IsTarget(ctx, f) {
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
			want := want(f)
			if actual == want {
				return
			}
			node, ok := ssautil.Node[*ast.CallExpr](ctx, call)
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
