package glance

import "context"

// Worker A worker interface that provides a synchronous Perform method
// for the ability to implement custom processing of an asynchronous task.
type Worker interface {
	Perform(context.Context, WorkerStream)
	Name() string
	Label() string
}

type WorkerStream interface {
	ID() string
	URL() string
}
