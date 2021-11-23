package glance

import (
	"context"
	"fmt"
	"github.com/zikwall/glance/pkg/log"
	"runtime"
	"sort"
	"sync"
	"time"
)

type WorkerItem struct {
	ID  string
	URL string
}

func (wi WorkerItem) GetID() string {
	return wi.ID
}

func (wi WorkerItem) GetURL() string {
	return wi.URL
}

type Workstation struct {
	spaces    map[string]*Workspace
	mu        sync.RWMutex
	startedAt time.Time
}

type Process struct {
	ctx       context.Context
	cancel    context.CancelFunc
	startedAt time.Time
}

func New(ctx context.Context, workers ...Worker) *Workstation {
	w := &Workstation{}
	w.mu = sync.RWMutex{}
	w.startedAt = time.Now()
	w.spaces = map[string]*Workspace{}

	for _, worker := range workers {
		w.spaces[worker.Name()] = NewWorkspace(ctx, worker)
	}

	return w
}

func (w *Workstation) Workspace(name string) (*Workspace, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	workspace, ok := w.spaces[name]

	if !ok {
		return nil, fmt.Errorf("workspace '%s' not found", name)
	}

	return workspace, nil
}

type (
	Info struct {
		Workspaces map[string]WorkspaceInfo `json:"workspaces"`
		Runtime    RuntimeInfo              `json:"runtime"`
	}
	WorkspaceInfo struct {
		Name           string        `json:"name"`
		Label          string        `json:"label"`
		TotalProcesses int           `json:"total_processes"`
		Processes      []ProcessInfo `json:"processes"`
	}
	ProcessInfo struct {
		Name      string `json:"name"`
		StartedAt string `json:"started_at"`
	}
	RuntimeInfo struct {
		NumGC       uint32  `json:"num_gc"`
		MemoryAlloc uint64  `json:"memory_alloc"`
		Gorutines   int     `json:"num_gorutines"`
		Uptime      float64 `json:"uptime"`
	}
)

func (w *Workstation) WorkstationInformation() Info {
	info := Info{
		Workspaces: map[string]WorkspaceInfo{},
	}

	memory := runtime.MemStats{}
	runtime.ReadMemStats(&memory)

	kb := func(b uint64) uint64 {
		return b / 1024
	}

	info.Runtime = RuntimeInfo{
		Uptime:      time.Since(w.startedAt).Seconds(),
		MemoryAlloc: kb(memory.Alloc),
		Gorutines:   runtime.NumGoroutine(),
		NumGC:       memory.NumGC,
	}

	w.mu.RLock()
	defer w.mu.RUnlock()

	for _, workspace := range w.spaces {
		workspaceInfo := WorkspaceInfo{
			Name:           workspace.worker.Name(),
			TotalProcesses: workspace.NumberOfActiveAsyncTasks(),
			Processes:      []ProcessInfo{},
		}

		for id, process := range workspace.tasks {
			workspaceInfo.Processes = append(workspaceInfo.Processes, ProcessInfo{
				StartedAt: process.startedAt.Format("2006-01-02 15:04:05"),
				Name:      fmt.Sprintf("%s%s", workspace.worker.Label(), id),
			})
		}

		sort.Slice(workspaceInfo.Processes, func(i, j int) bool {
			return workspaceInfo.Processes[i].StartedAt > workspaceInfo.Processes[j].StartedAt
		})

		info.Workspaces[workspace.worker.Name()] = workspaceInfo
	}

	return info
}

func (w *Workstation) Drop() error {
	for _, space := range w.spaces {
		if err := space.Drop(); err != nil {
			log.Warning(err)
		}
	}

	return nil
}

func (w *Workstation) DropMsg() string {
	return "glance completed successfully"
}
