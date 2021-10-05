package glance

import "context"

// Worker A worker interface that provides a synchronous Perform method
// for the ability to implement custom processing of an asynchronous task.
//
// func (w *MockWorker) Perform(ctx context.Context, id int) {
//	for { <- Observable
//		select {
//		case <-ctx.Done(): <- Providable
//			return
//		}
//
//		do stuff...
//	}
// }
type Worker interface {
	Perform(context.Context, WorkerStream)
	Name() string
	Label() string
}

type WorkerStream interface {
	ID() string
	URL() string
}
