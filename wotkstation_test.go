package glance

import (
	"context"
	"errors"
	"github.com/zikwall/glance/pkg/workers/errorless"
	"testing"
	"time"
)

type MockWorker struct{}
type MockWorkerStream struct {
	id  string
	app string
}

func (w MockWorkerStream) ID() string {
	return w.id
}

func (w MockWorkerStream) URL() string {
	return w.id
}

func (w *MockWorker) Perform(ctx context.Context, _ WorkerStream) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 1):
		}
	}
}

func (w *MockWorker) Name() string {
	return "mock_worker"
}

func (w *MockWorker) Label() string {
	return "mock_worker/"
}

func TestNewWorkstation(t *testing.T) {
	ctx, cancelFunc := context.WithCancel(context.Background())

	workstation := New(ctx, &MockWorker{})
	workspace, err := workstation.Workspace("mock_worker")

	if err != nil {
		t.Fatal(err)
	}

	t.Run("it should be work correctly", func(t *testing.T) {
		if err := workspace.PerformAsync(MockWorkerStream{"1", "in"}); err != nil {
			t.Fatal(err)
		}

		if err := workspace.PerformAsync(MockWorkerStream{"2", "out"}); err != nil {
			t.Fatal(err)
		}

		if err := workspace.PerformAsync(MockWorkerStream{"3", "outn"}); err != nil {
			t.Fatal(err)
		}

		t.Run("it should be active task", func(t *testing.T) {
			if active := workspace.lookupAsyncTask("3"); !active {
				t.Fatal("Fail, expect active task #3")
			}
		})

		if err := workspace.PerformAsync(MockWorkerStream{"3", "in"}); err == nil {
			t.Fatal("Fail, expect error")
		} else {
			e := &errorless.TaskAlreadyExistsError{}

			if errors.As(err, &e) == false {
				t.Fatal("Fail, expect typed error")
			}
		}

		if err := workspace.FinishAsyncTask("2"); err != nil {
			t.Fatal(err)
		}

		if err := workspace.FinishAsyncTask("1"); err != nil {
			t.Fatal(err)
		}

		<-time.After(time.Millisecond * 100)

		if err := workspace.FinishAsyncTask("1"); err == nil {
			t.Fatal("Fail, expect error")
		} else {
			e := &errorless.TaskNotFoundError{}

			if errors.As(err, &e) == false {
				t.Fatal("Fail, expect typed error")
			}
		}

		t.Run("it should be get correct size of active tasks in worker pool", func(t *testing.T) {
			if workspace.NumberOfActiveAsyncTasks() != 1 {
				t.Fatal("Fail, expect 1 active task in pool")
			}

			if err := workspace.FinishAsyncTask("3"); err != nil {
				t.Fatal(err)
			}

			<-time.After(time.Millisecond * 100)

			if workspace.NumberOfActiveAsyncTasks() != 0 {
				t.Fatal("Fail, expect empty worker pool")
			}

			t.Run("it should be close all tasks", func(t *testing.T) {
				if err := workspace.PerformAsync(MockWorkerStream{"1", "in"}); err != nil {
					t.Fatal(err)
				}

				if err := workspace.PerformAsync(MockWorkerStream{"2", "out"}); err != nil {
					t.Fatal(err)
				}

				if err := workspace.PerformAsync(MockWorkerStream{"3", "outn"}); err != nil {
					t.Fatal(err)
				}

				cancelFunc()

				if err := workspace.Drop(); err != nil {
					t.Fatal(err)
				}

				if workspace.NumberOfActiveAsyncTasks() != 0 {
					t.Fatal("Fail, expect empty worker pool")
				}
			})
		})
	})
}
