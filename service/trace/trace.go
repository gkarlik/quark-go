package trace

import (
	"golang.org/x/net/context"
)

type Span interface {
	Start(name string) *Span
	StartFromContext(ctx context.Context, name string) *Span
	Finish()
}
