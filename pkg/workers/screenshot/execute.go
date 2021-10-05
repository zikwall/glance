package screenshot

import (
	"fmt"
	"github.com/zikwall/glance/pkg/log"
	"github.com/zikwall/glance/pkg/workers/errorless"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path"
)

type captureWatcher struct {
	cmd    *exec.Cmd
	temp   *os.File
	layout string
}

func execute(rtmp, upload, id string) (*captureWatcher, error) {
	file, err := ioutil.TempFile("./tmp", fmt.Sprintf("%s_go_tmp_capture_*.log", id))

	if err != nil {
		return nil, err
	}

	u, err := url.Parse(upload)

	if err != nil {
		return nil, err
	}

	imagePushURL := ""
	useUpdate := "1"
	useStrftime := "0"

	u.Path = path.Join(u.Path, fmt.Sprintf("%s.jpg", id))
	imagePushURL = u.String()

	rt, err := url.Parse(rtmp)

	if err != nil {
		return nil, err
	}

	args := []string{
		"-y",
		"-nostdin",
		"-threads", "1",
		"-skip_frame", "nokey",
		"-i", rt.String(),
		"-vsync", "0",
		"-r", "30",
		"-f", "image2",
		"-strftime", useStrftime,
		"-update", useUpdate,
		"-protocol_opts", "method=PUT",
		imagePushURL,
	}

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = file
	cmd.Stderr = file

	if err := cmd.Start(); err != nil {
		_ = file.Close()
		_ = os.Remove(file.Name())

		return nil, err
	}

	return &captureWatcher{cmd: cmd, temp: file}, nil
}

func (c *captureWatcher) clearResources() {
	if err := c.temp.Close(); err != nil {
		log.Warning(err)
	}

	if err := os.Remove(c.temp.Name()); err != nil {
		log.Warning(err)
	}
}

func (c *captureWatcher) killProcesses(name, id string) {
	if err := c.cmd.Process.Kill(); err != nil && !errorless.IsAlreadyFinished(err) {
		errorless.ErrorFailedKillProcess(name, id, c.cmd.Process.Pid, err)
	} else {
		errorless.InfoSuccessKillProcess(name, id, c.cmd.Process.Pid)
	}
}
