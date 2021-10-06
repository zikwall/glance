package screenshot

import (
	"fmt"
	"github.com/zikwall/glance/pkg/log"
	"github.com/zikwall/glance/pkg/workers/errorless"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
)

type process struct {
	cmd    *exec.Cmd
	temp   *os.File
	layout string
}

func (w *Worker) execute(rtmp, upload, id string) (*process, error) {
	file, err := ioutil.TempFile("./tmp", fmt.Sprintf("%s_go_tmp_capture_*.log", id))
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(upload)
	if err != nil {
		return nil, err
	}

	rt, err := url.Parse(rtmp)
	if err != nil {
		return nil, err
	}

	useUpdate := "1"
	useStrftime := "0"
	u.Path = w.formatter.Format(u, nil, id, false)

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
		u.String(),
	}

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = file
	cmd.Stderr = file

	if err := cmd.Start(); err != nil {
		_ = file.Close()
		_ = os.Remove(file.Name())

		return nil, err
	}

	return &process{cmd: cmd, temp: file}, nil
}

func (p *process) clearResources() {
	if err := p.temp.Close(); err != nil {
		log.Warning(err)
	}

	if err := os.Remove(p.temp.Name()); err != nil {
		log.Warning(err)
	}
}

func (p *process) killProcesses(name, id string) {
	if err := p.cmd.Process.Kill(); err != nil && !errorless.IsFinished(err) {
		errorless.Warning(name,
			fmt.Sprintf("[#%s] failed to kill async proccess PID %d %s", id, p.cmd.Process.Pid, err),
		)
	}
}
