package glance

import (
	"context"
	"fmt"
	"github.com/zikwall/glance/pkg/log"
	"github.com/zikwall/glance/pkg/workers/errorless"
	"sync"
	"time"
)

type Workspace struct {
	mu      sync.RWMutex
	tasks   map[string]Process
	worker  Worker
	context context.Context
	// This property simultaneously serves as a counter for asynchronous tasks
	// and a mechanism for waiting/completing the task, for successful completion
	wg sync.WaitGroup
	// This property serves as a flag for successful completion of all asynchronous tasks
	done chan struct{}
}

func NewWorkspace(context context.Context, worker Worker) *Workspace {
	return &Workspace{
		mu:      sync.RWMutex{},
		tasks:   map[string]Process{},
		worker:  worker,
		context: context,
		wg:      sync.WaitGroup{},
		done:    make(chan struct{}),
	}
}

// PerformAsync Initializes the task and runs it in the background.
// The task is handled by a worker defined by the worker interface, where the Perform method is defined
func (w *Workspace) PerformAsync(stream WorkerStream) error {
	id := stream.ID()
	if w.lookupAsyncTask(id) {
		return errorless.TaskAlreadyExists(id)
	}

	ctx, cancel := context.WithCancel(w.context)

	w.attach(id, Process{
		ctx:       ctx,
		cancel:    cancel,
		startedAt: time.Now(),
	})

	w.wg.Add(1)
	go func(id string, str WorkerStream) {
		w.launchAsyncTaskMsg(id)
		defer func() {
			w.tryCancelAndDetach(id)
			w.wg.Done()
			w.shutdownAsyncTaskMsg(id)
		}()

		// The method must work synchronously, otherwise it will be completed
		w.worker.Perform(ctx, str)
	}(id, stream)

	return nil
}

// FinishAsyncTask The method terminates a specific asynchronous task by removing it from the task pool.
func (w *Workspace) FinishAsyncTask(id string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, ok := w.tasks[id]; !ok {
		return errorless.TaskNotFound(id)
	}

	w.tasks[id].cancel()
	delete(w.tasks, id)

	return nil
}

func (w *Workspace) NumberOfActiveAsyncTasks() int {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return len(w.tasks)
}

func (w *Workspace) ActiveTasks() Collection {
	w.mu.RLock()
	collection := Collection{
		Streams: make(map[string]WorkerItem, len(w.tasks)),
	}

	for id := range w.tasks {
		collection.Streams[id] = WorkerItem{
			Id: id,
		}
	}
	w.mu.RUnlock()
	return collection
}

func (w *Workspace) lookupAsyncTask(id string) bool {
	w.mu.RLock()
	_, ok := w.tasks[id]
	w.mu.RUnlock()

	return ok
}

func (w *Workspace) attach(id string, process Process) {
	w.mu.Lock()
	w.tasks[id] = process
	w.mu.Unlock()
}

// Safe deletion from the pool
func (w *Workspace) tryCancelAndDetach(id string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if _, ok := w.tasks[id]; ok && w.tasks[id].ctx.Err() == nil {
		w.tasks[id].cancel()
	}

	delete(w.tasks, id)
}

// messages
func (w *Workspace) launchAsyncTaskMsg(id string) {
	log.Info(errorless.WithWorkerLabel(w.worker.Name(), fmt.Sprintf("Launch async task ID: %s, attach to pool", id)))
}

func (w *Workspace) shutdownAsyncTaskMsg(id string) {
	log.Info(errorless.WithWorkerLabel(w.worker.Name(), fmt.Sprintf("Finished async task ID: %s, removed from pool...", id)))
}

func (w *Workspace) exitMsg(success bool) {
	if success {
		log.Info(errorless.WithWorkerLabel(w.worker.Name(), "Exit from workspace without error"))
	} else {
		log.Warning(errorless.WithWorkerLabel(w.worker.Name(), "Exited after a long wait of 10 seconds from workspace"))
	}
}

func (w *Workspace) doneAllAsyncTasksMsg() {
	log.Info(errorless.WithWorkerLabel(w.worker.Name(), "All asynchronous tasks in workspace completed successfully!"))
}

func (w *Workspace) Context() context.Context {
	return w.context
}

// The method waits for graceful completion or crashes after a certain amount of time
func (w *Workspace) await() {
	select {
	case <-w.done:
		w.exitMsg(true)
	case <-time.After(10 * time.Second):
		w.exitMsg(false)
	}
}

// Drop This and subsequent methods implement the Notifier interface,
// which is automatically terminated when the server is stopped.
// Completion occurs synchronously,
// which represents the possibility of waiting for the completion of all asynchronous tasks,
// or an emergency termination
func (w *Workspace) Drop() error {
	go func() {
		// wait all async tasks
		w.wg.Wait()
		// to inform about the successful completion of the task
		w.done <- struct{}{}
		w.doneAllAsyncTasksMsg()
	}()

	// waiting for a message about the completion of tasks, or completing
	w.await()

	return nil
}

func (w *Workstation) Drop() error {
	for _, space := range w.spaces {
		if err := space.Drop(); err != nil {
			log.Warning(err)
		}
	}

	return nil
}

func (w *Workstation) DropMsg() string {
	return "glance completed successfully"
}
