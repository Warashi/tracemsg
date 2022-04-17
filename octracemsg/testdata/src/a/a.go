package a

import (
	"context"

	"go.opencensus.io/trace"
)

func F(ctx context.Context) {
	trace.StartSpan(ctx, "a.F")
}

func G(ctx context.Context) { // want `a\.G should call trace\.StartSpan`
}

func H(ctx context.Context) {
	trace.StartSpan(ctx, "a.F") // want `span name should be .*`
}

func NoCtx() {
}

func noExported(ctx context.Context) {
}

type Hoge struct{}

func (h *Hoge) Func(ctx context.Context) {
	trace.StartSpan(ctx, "a.Hoge#Func") // want `span name should be .*`
}

func (h *Hoge) Fail(ctx context.Context) { // want `\(\*a\.Hoge\)\.Fail should call trace\.StartSpan`
}

func (h *Hoge) NoCtx() {
}

func (h *Hoge) noExported() {
}