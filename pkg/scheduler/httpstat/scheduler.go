package httpstat

import (
	"context"
	"time"

	"github.com/zikwall/glance"
	"github.com/zikwall/glance/pkg/log"
)

const threads = 3

type Status struct {
	ID    string
	Code  int
	Error error
}

type Scheduler struct {
	fetcher glance.Fetcher
	storage StatusWriter
	options *Options
}

type Options struct {
	HTTPHeaders map[string]string
	Refresh     time.Duration
}

func NewScheduler(fetcher glance.Fetcher, storage StatusWriter, options *Options) *Scheduler {
	scheduler := &Scheduler{fetcher: fetcher, storage: storage, options: options}
	return scheduler
}

func (s *Scheduler) RunContext(ctx context.Context) {
	log.Info("run HTTP stats scheduler")
	defer log.Info("stop HTTP stats scheduler")
	ticker := time.NewTicker(s.options.Refresh)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			log.Info("get HTTP statuses")

			streams, err := s.fetcher.FetchStreams(ctx)
			if err != nil {
				log.Warning(err)
				continue
			}

			now := glance.Datetime(time.Now())
			dat := glance.Date(time.Now())
			statuses := getHTTPStatuses(ctx, streams, s.options.HTTPHeaders)

			for _, status := range statuses {
				if status.Error != nil {
					log.Warning(err)
					continue
				}

				err = s.storage.Write(Bucket{
					StreamID:   status.ID,
					Code:       status.Code,
					InsertTS:   now,
					InsertDate: dat,
				})
				if err != nil {
					log.Warning(err)
				}
			}
		}
	}
}

func getHTTPStatuses(ctx context.Context, streams glance.Collection, headers map[string]string) []Status {
	th := make([][]request, threads)
	cn := parts(len(streams.Streams))

	for i := 1; i < threads; i++ {
		th[i] = make([]request, 0, cn)
	}

	index := 0
	for _, stream := range streams.Streams {
		if len(th[index]) >= cn {
			index++
		}

		th[index] = append(th[index], request{
			id:  stream.GetID(),
			url: stream.GetURL(),
		})
	}

	return asyncRequests(ctx, cn*threads, headers, th...)
}
