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
}

func New(name string, storage glance.Storage) *Worker {
	w := &Worker{name: name, storage: storage}
	return w
}

func (w *Worker) Name() string {
	return "metric"
}

func (w *Worker) Label() string {
	return "metric "
}

func (w *Worker) Perform(ctx context.Context, stream glance.WorkerStream) {
	id := stream.ID()
	watcher, err := execute(stream.URL(), id)

	if err != nil {
		errorless.ErrorAsyncNoStarted(w.Name(), id, err)
		return
	}

	// If an asynchronous task fails with an ffmpeg process error,
	// then there is no need to kill the process, since it has already been killed
	// example give error - os: process already finished
	killFFMPEG := true
	defer func() {
		watcher.clearResources()
		if killFFMPEG {
			watcher.killProcesses(w.name, id)
		}
	}()

	FFMPEGOut := make(chan string, 1000)
	FFMPEGKill := make(chan error, 1)
	// Runs a separate sub-thread, because when running in a single thread,
	// there is a lock while waiting for the buffer to be read.
	// In turn blocking by the reader will not allow the background task to finish gracefully
	go func() {
		bufioReader := bufio.NewReader(watcher.r)
		for {
			line, isPrefix, err := bufioReader.ReadLine()
			if err != nil {
				if err != io.EOF {
					errorless.ErrorCloseProcessStdoutReader(w.Name(), id, err)
				} else {
					errorless.InfoCloseProcessStdoutReader(w.Name(), id)
				}

				return
			}

			str := string(line)
			if isPrefix || str == "" {
				continue
			}

			FFMPEGOut <- str
		}
	}()

	// We listen to the FFMPEG process termination signal,
	// this will provide an opportunity to remove the task from the pool and restart it if necessary
	//
	// Note: We listen to the context so as not to leave active goroutines when the task is completed
	go func() {
		select {
		case FFMPEGKill <- watcher.cmd.Wait():
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
		case err = <-FFMPEGKill:
			if exitError, ok := err.(*exec.ExitError); ok {
				err = fmt.Errorf("exit code is %d: %s", exitError.ExitCode(), exitError.Error())
			}

			errorless.ErrorProcessKilled(w.Name(), id, watcher.cmd.Process.Pid, err)
			killFFMPEG = false

			return
		case outPartial := <-FFMPEGOut:
			parts := strings.Split(outPartial, ",")

			if len(parts) == partialSize || len(parts) == partialSizeWithBrokenSideData {
				frame.IncreasingContinue(
					stringToInt(parts[bytesPos]),
				)

				if isKeyframe(parts[keyframePos]) {
					frame.Height = stringToInt(parts[heightPos])
					pktPtsTime := math.Ceil((stringToFloat64(parts[timePos]))*1000000) / 1000000

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
