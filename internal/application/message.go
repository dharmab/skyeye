package application

import (
	"context"
)

// Message binds a context to data. It should only be used for passing a
// request context and data together through channels. The receiver should
// immediately extract the context and data from the message and use them
// independently. If you're passing a Message as a function parameter, you are
// misusing it.
type Message[T any] struct {
	Context context.Context
	Data    T
}

// AsMessage creates a Message from a context and data. The message should be
// immediately sent through a channel.
func AsMessage[T any](ctx context.Context, data T) Message[T] {
	return Message[T]{Context: ctx, Data: data}
}
