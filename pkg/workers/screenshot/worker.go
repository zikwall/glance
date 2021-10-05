package screenshot

import (
	"context"
	"fmt"
	"github.com/zikwall/glance"
	"github.com/zikwall/glance/pkg/workers/errorless"
	"os/exec"
)

type Worker struct {
	upload    string
	name      string
	formatter UrlFormatter
}

func New(name, upload string, formatter UrlFormatter) *Worker {
	worker := &Worker{
		upload:    upload,
		name:      name,
		formatter: formatter,
	}
	return worker
}

func (w *Worker) Name() string {
	return w.name
}

func (w *Worker) Label() string {
	return "screenshot "
}

func (w *Worker) Perform(ctx context.Context, stream glance.WorkerStream) {
	id := stream.ID()
	screenshot, err := w.execute(stream.URL(), w.upload, id)

	if err != nil {
		errorless.ErrorAsyncNoStarted(w.Name(), id, err)
		return
	}

	killFFMPEG := true
	defer func() {
		screenshot.clearResources()
		if killFFMPEG {
			screenshot.killProcesses(w.name, id)
		}
	}()

	FFMPEGKill := make(chan error, 1)
	go func() {
		select {
		case FFMPEGKill <- screenshot.cmd.Wait():
			return
		case <-ctx.Done():
			return
		}
	}()

	select {
	case <-ctx.Done():
		return
	case err = <-FFMPEGKill:
		if exitError, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("exit code is %d: %s", exitError.ExitCode(), exitError.Error())
		}

		errorless.ErrorProcessKilled(w.Name(), id, screenshot.cmd.Process.Pid, err)

		killFFMPEG = false
		return
	}
}
