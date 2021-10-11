package httpstat

import (
	"context"
	"github.com/zikwall/glance"
	"github.com/zikwall/glance/pkg/log"
	"time"
)

const threads = 3

type Status struct {
	ID    string
	Code  int
	Error error
}

type scheduler struct {
	fetcher         glance.Fetcher
	storage         StatusWriter
	refreshInterval time.Duration
}

func NewScheduler(fetcher glance.Fetcher, storage StatusWriter, refresh time.Duration) *scheduler {
	scheduler := &scheduler{fetcher: fetcher, storage: storage, refreshInterval: refresh}
	return scheduler
}

func (s *scheduler) RunContext(ctx context.Context) {
	ticker := time.NewTicker(s.refreshInterval)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			streams, err := s.fetcher.FetchStreams(ctx)
			if err != nil {
				log.Warning(err)
				continue
			}

			now := glance.Datetime(time.Now())
			dat := glance.Date(time.Now())
			statuses := getHTTPStatuses(ctx, streams)

			for _, status := range statuses {
				if status.Error != nil {
					log.Warning(err)
					continue
				}

				err = s.storage.Write(Bucket{
					StreamID:   "",
					Code:       0,
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

func getHTTPStatuses(ctx context.Context, streams glance.Collection) []Status {
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
			id:  stream.ID(),
			url: stream.URL(),
		})
	}

	return asyncRequests(ctx, cn*threads, th...)
}
