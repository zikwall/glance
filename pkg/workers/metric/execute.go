package metric

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"

	"github.com/zikwall/glance/pkg/log"
	"github.com/zikwall/glance/pkg/workers/errorless"
)

type process struct {
	cmd *exec.Cmd
	r   *io.PipeReader
	w   *io.PipeWriter
	f   *os.File
}

func (a *Worker) execute(rtmp, id string) (*process, error) {
	file, err := ioutil.TempFile("./tmp", fmt.Sprintf("%s_go_tmp_stream_err_*.log", id))
	if err != nil {
		return nil, err
	}

	rt, err := url.Parse(rtmp)
	if err != nil {
		return nil, err
	}

	var args []string
	for _, value := range a.options.HTTPHeaders {
		args = append(args, "-headers", value)
	}
	args = append(args, []string{
		"-loglevel", "error",
		"-threads", "1",
		"-select_streams", "v:0",
		"-show_frames",
		"-show_entries", "frame=key_frame,pkt_pts_time,pkt_size,height,repeat_pict",
		"-of", "csv",
		rt.String(),
	}...)

	r, w := io.Pipe()
	cmd := exec.Command("ffprobe", args...)
	cmd.Stdout = w
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &process{cmd: cmd, r: r, w: w, f: file}, nil
}

func (p *process) Reader() io.Reader {
	return bufio.NewReader(p.r)
}

func (p *process) clearResources() {
	if err := p.f.Close(); err != nil {
		log.Warning(err)
	}

	if err := p.w.Close(); err != nil {
		log.Warning(err)
	}

	if err := os.Remove(p.f.Name()); err != nil {
		log.Warning(err)
	}
}

func (p *process) killProcesses(name, id string) {
	if err := p.cmd.Process.Kill(); err != nil && !errorless.IsFinished(err) {
		errorless.Warning(name,
			fmt.Sprintf("[#%s] failed to kill async process PID %d %s", id, p.cmd.Process.Pid, err),
		)
	}
}
