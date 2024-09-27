package traces

import (
	"context"
)

type Tracer interface {
	Trace(context.Context)
}
