package errorless

import (
	"fmt"
	"strings"

	"github.com/zikwall/glance/pkg/log"
)

func Warning(worker, message string) {
	log.Warning(Labeled(worker, message))
}

const ProcessIsDie = "[#%s] async process PID %d was terminated with an error, task is removed from the pool" +
	" and will be restarted in the future. Previous error '%s'"

func Labeled(worker, message string) string {
	return fmt.Sprintf("%s %s", log.Colored(fmt.Sprintf("[%s]", worker), log.Yellow), message)
}

func IsFinished(err error) bool {
	return strings.EqualFold(err.Error(), "os: process already finished")
}
