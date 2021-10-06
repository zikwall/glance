package metric

import (
	"bufio"
	"fmt"
	"github.com/zikwall/glance/pkg/log"
	"github.com/zikwall/glance/pkg/workers/errorless"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
)

type metricWatcher struct {
	cmd *exec.Cmd
	r   *io.PipeReader
	w   *io.PipeWriter
	f   *os.File
}

func execute(rtmp string, id string) (*metricWatcher, error) {
	file, err := ioutil.TempFile("./tmp", fmt.Sprintf("%s_go_tmp_stream_err_*.log", id))
	if err != nil {
		return nil, err
	}

	rt, err := url.Parse(rtmp)
	if err != nil {
		return nil, err
	}

	args := []string{
		"-loglevel", "error",
		"-threads", "1",
		"-select_streams", "v:0",
		"-show_frames",
		"-show_entries", "frame=key_frame,pkt_pts_time,pkt_size,height,repeat_pict",
		"-of", "csv",
		rt.String(),
	}

	r, w := io.Pipe()
	cmd := exec.Command("ffprobe", args...)
	cmd.Stdout = w
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &metricWatcher{cmd: cmd, r: r, w: w, f: file}, nil
}

func (w *metricWatcher) Reader() io.Reader {
	return bufio.NewReader(w.r)
}

func (w *metricWatcher) clearResources() {
	if err := w.f.Close(); err != nil {
		log.Warning(err)
	}

	if err := w.w.Close(); err != nil {
		log.Warning(err)
	}

	if err := os.Remove(w.f.Name()); err != nil {
		log.Warning(err)
	}
}

func (w *metricWatcher) killProcesses(name, id string) {
	if err := w.cmd.Process.Kill(); err != nil && !errorless.IsAlreadyFinished(err) {
		errorless.ErrorFailedKillProcess(name, id, w.cmd.Process.Pid, err)
	} else {
		errorless.InfoSuccessKillProcess(name, id, w.cmd.Process.Pid)
	}
}
