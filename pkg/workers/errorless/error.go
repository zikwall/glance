package errorless

import (
	"fmt"
	"github.com/zikwall/glance/pkg/log"
	"strings"
)

func info(worker, message string) {
	log.Info(WithWorkerLabel(worker, message))
}

func warning(worker, message string) {
	log.Warning(WithWorkerLabel(worker, message))
}

func WithWorkerLabel(worker string, message string) string {
	return fmt.Sprintf("%s %s", log.Colored(fmt.Sprintf("[%s]", worker), log.Yellow), message)
}

func ErrorCloseProcessStdoutReader(worker string, id string, err error) {
	warning(worker, fmt.Sprintf("reading from stdout for asynctask #%s completed (with error), cause %s", id, err))
}

func InfoCloseProcessStdoutReader(worker string, id string) {
	info(worker, fmt.Sprintf("eead from stdout for async task #%s completed, successfully!", id))
}

func InfoSuccessKillProcess(worker string, id string, pid int) {
	info(worker, fmt.Sprintf("async task #%s sub process PID %d successfully killed", id, pid))
}

func ErrorFailedKillProcess(worker string, id string, pid int, err error) {
	warning(worker, fmt.Sprintf("failed to kill async task #%s sub process PID %d %s", id, pid, err))
}

func ErrorAsyncNoStarted(worker string, id string, err error) {
	warning(worker, fmt.Sprintf("async task #%s will not be started, previous error: %s", id, err))
}

func ErrorProcessKilled(worker string, id string, pid int, err error) {
	warning(worker, fmt.Sprintf(`async task #%s process PID %d was terminated with an error, the task is removed from the pool and will be restarted in the future. Previous error '%s'`, id, pid, err))
}

func IsAlreadyFinished(err error) bool {
	return strings.EqualFold(err.Error(), "os: process already finished")
}
