// Package traces defines types and interfaces for collecting and publishing
// GCI workflow traces.
package traces

import (
	"context"
)

// Tracer publishes traces to a remote service.
type Tracer interface {
	// Trace publishes any trace in the given context.
	Trace(context.Context)
}
