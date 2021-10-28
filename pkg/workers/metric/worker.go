package metric

import (
	"bufio"
	"context"
	"fmt"
	"github.com/zikwall/glance"
	"github.com/zikwall/glance/pkg/log"
	"github.com/zikwall/glance/pkg/workers/errorless"
	"io"
	"math"
	"os/exec"
	"strings"
)

type Worker struct {
	name    string
	storage glance.Storage
	options *Options
}

type Options struct {
	HTTPHeaders map[string]string
}

func New(name string, storage glance.Storage, options *Options) *Worker {
	w := &Worker{name: name, storage: storage, options: options}
	return w
}

func (w *Worker) Name() string {
	return "metric"
}

func (w *Worker) Label() string {
	return "metric"
}

func (w *Worker) Perform(ctx context.Context, stream glance.WorkerStream) {
	id := stream.ID()

	process, err := w.execute(stream.URL(), id)
	if err != nil {
		errorless.Warning(w.Name(),
			fmt.Sprintf("[#%s] async process will not be started, previous error: %s", id, err),
		)

		return
	}

	// If an asynchronous task fails with an ffmpeg process error,
	// then there is no need to kill the process, since it has already been killed
	// example give error - os: process already finished
	NeedKillFFMPEG := true
	defer func() {
		process.clearResources()
		if NeedKillFFMPEG {
			process.killProcesses(w.name, id)
		}
	}()

	EventReceiveFFMPEG := make(chan string, 1000)
	EventKillFFMPEG := make(chan error, 1)
	// Runs a separate sub-thread, because when running in a single thread,
	// there is a lock while waiting for the buffer to be read.
	// In turn blocking by the reader will not allow the background task to finish gracefully
	go func() {
		buffer := bufio.NewReader(process.r)
		for {
			line, isPrefix, err := buffer.ReadLine()
			if err != nil {
				if err != io.EOF {
					errorless.Warning(w.Name(),
						fmt.Sprintf("[#%s] reading from stdout completed (with error), cause %s", id, err))
				}

				return
			}

			str := string(line)
			if isPrefix || str == "" {
				continue
			}

			EventReceiveFFMPEG <- str
		}
	}()

	// We listen to the FFMPEG process termination signal,
	// this will provide an opportunity to remove the task from the pool and restart it if necessary
	//
	// Note: We listen to the context so as not to leave active goroutines when the task is completed
	go func() {
		select {
		case EventKillFFMPEG <- process.cmd.Wait():
			return
		case <-ctx.Done():
			return
		}
	}()

	const keyframePos = 1
	const timePos = 2
	const bytesPos = 3
	const heightPos = 4

	frame := glance.Frame{}
	var lastTimestamp float64
	for {
		select {
		case <-ctx.Done():
			return
		case err = <-EventKillFFMPEG:
			NeedKillFFMPEG = false

			if exitError, ok := err.(*exec.ExitError); ok {
				err = fmt.Errorf("exit code is %d: %s", exitError.ExitCode(), exitError.Error())
			}
			errorless.Warning(w.Name(), fmt.Sprintf(errorless.ProcessIsDie, id, process.cmd.Process.Pid, err))

			return
		case csvPartials := <-EventReceiveFFMPEG:
			partials := strings.Split(csvPartials, ",")

			if len(partials) == partialSize || len(partials) == partialSizeWithBrokenSideData {
				frame.IncreasingContinue(
					stringToInt(partials[bytesPos]),
				)

				if isKeyframe(partials[keyframePos]) {
					frame.Height = stringToInt(partials[heightPos])
					pktPtsTime := math.Ceil((stringToFloat64(partials[timePos]))*1000000) / 1000000

					seconds := math.Ceil((pktPtsTime-lastTimestamp)*1000000) / 1000000
					frame.Seconds = seconds

					if frame.Frames != 1 {
						batch := glance.CreateBatch(id, frame)
						if err := w.storage.ProcessFrameBatch(batch); err != nil {
							log.Warning(err)
						}
					}

					frame.Cleanup()
					lastTimestamp = pktPtsTime
				}
			}
		}
	}
}

func isKeyframe(frame string) bool {
	return frame == "1"
}
