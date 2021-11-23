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
	formatter URLFormatter
	options   *Options
}

type Options struct {
	HTTPHeaders []string
}

func New(name, upload string, formatter URLFormatter, options *Options) *Worker {
	worker := &Worker{
		upload:    upload,
		name:      name,
		formatter: formatter,
		options:   options,
	}
	return worker
}

func (w *Worker) Name() string {
	return w.name
}

func (w *Worker) Label() string {
	return "screenshot"
}

func (w *Worker) Perform(ctx context.Context, stream glance.WorkerStream) {
	id := stream.GetID()

	process, err := w.execute(stream.GetURL(), w.upload, id)
	if err != nil {
		errorless.Warning(w.Name(),
			fmt.Sprintf("[#%s] async process will not be started, previous error: %s", id, err),
		)

		return
	}

	NeedKillFFMPEG := true
	defer func() {
		process.clearResources()
		if NeedKillFFMPEG {
			process.killProcesses(w.name, id)
		}
	}()

	EventKillFFMPEG := make(chan error, 1)
	go func() {
		select {
		case EventKillFFMPEG <- process.cmd.Wait():
			return
		case <-ctx.Done():
			return
		}
	}()

	select {
	case <-ctx.Done():
		return
	case err = <-EventKillFFMPEG:
		NeedKillFFMPEG = false

		if exitError, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("exit code is %d: %s", exitError.ExitCode(), exitError.Error())
		}
		errorless.Warning(w.Name(),
			fmt.Sprintf(errorless.ProcessIsDie, id, process.cmd.Process.Pid, err),
		)

		return
	}
}
