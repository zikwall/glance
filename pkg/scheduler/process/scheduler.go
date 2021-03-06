package process

import (
	"context"
	"fmt"
	"time"

	"github.com/zikwall/glance"
	"github.com/zikwall/glance/pkg/log"
)

type Options struct {
	RefreshInterval     time.Duration
	WorkspaceScreenshot *glance.Workspace
	WorkspaceMetrics    *glance.Workspace
}

type Scheduler struct {
	fetcher glance.Fetcher
}

func NewScheduler(fetcher glance.Fetcher) *Scheduler {
	scheduler := &Scheduler{fetcher: fetcher}
	return scheduler
}

func (s *Scheduler) RunContext(ctx context.Context, options Options) {
	s.justRun(ctx, options.WorkspaceMetrics, options.WorkspaceScreenshot)

	defer log.Info("monitoring thread update scheduler is being terminated")
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(options.RefreshInterval):
			log.Info("monitoring thread update scheduler is started")

			fetchedJobs, err := s.fetcher.FetchStreams(ctx)
			if err != nil {
				log.Warning(err)
				continue
			}
			if options.WorkspaceScreenshot != nil {
				refresh("[SCHEDULER][SCREENSHOT WORKER]", options.WorkspaceScreenshot, fetchedJobs)
			}
			if options.WorkspaceMetrics != nil {
				refresh("[SCHEDULER][METRICS WORKER]", options.WorkspaceMetrics, fetchedJobs)
			}
		}
	}
}

func refresh(t string, space *glance.Workspace, fetched glance.Collection) {
	spaced := space.ActiveTasks()
	stopped, started := 0, 0

	// if is active
	// if is not fetched
	// stop
	for activeID := range spaced.Streams {
		if !fetched.Exist(activeID) {
			if err := space.FinishAsyncTask(activeID); err != nil {
				log.Warning(fmt.Sprintf("%s STOPING  %s", t, err))
			} else {
				stopped++
			}
		}
	}

	// if is not active
	// if is fetched
	// start
	for fetchedID, stream := range fetched.Streams {
		// if not exists -> run async
		if !spaced.Exist(fetchedID) {
			if err := space.PerformAsync(stream); err != nil {
				log.Warning(fmt.Sprintf("%s STARTING %s", t, err))
			} else {
				started++
			}
		}
	}

	if started+stopped > 0 {
		log.Info(fmt.Sprintf("%s started %d stoped %d", t, started, stopped))
	} else {
		log.Info(fmt.Sprintf("%s nothing to update", t))
	}
}

func (s *Scheduler) justRun(ctx context.Context, monitoring, screenshot *glance.Workspace) {
	streams, err := s.fetcher.FetchStreams(ctx)
	if err != nil {
		log.Warning(err)
		return
	}
	for _, stream := range streams.Streams {
		if monitoring != nil {
			if err := monitoring.PerformAsync(stream); err != nil {
				log.Warning(err)
			}
		}

		if screenshot != nil {
			if err := screenshot.PerformAsync(stream); err != nil {
				log.Warning(err)
			}
		}
	}
}
