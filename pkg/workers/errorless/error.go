package errorless

import (
	"fmt"
	"github.com/zikwall/glance/pkg/log"
	"strings"
)

func Info(worker, message string) {
	log.Info(labeled(worker, message))
}

func Warning(worker, message string) {
	log.Warning(labeled(worker, message))
}

const ProcessIsDie = "[#%s] async process PID %d was terminated with an error, task is removed from the pool" +
	" and will be restarted in the future. Previous error '%s'"

func labeled(worker string, message string) string {
	return fmt.Sprintf("%s %s", log.Colored(fmt.Sprintf("[%s]", worker), log.Yellow), message)
}

func IsFinished(err error) bool {
	return strings.EqualFold(err.Error(), "os: process already finished")
}
