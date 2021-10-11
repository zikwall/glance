package http

import (
	"context"
	"fmt"
	"github.com/zikwall/glance/pkg/log"
	"math"
	"net/http"
	"sync"
	"time"
)

type FutureResponse struct {
	ID       string
	HTTPCode int
	Error    error
}

type Request struct {
	ID  string
	URL string
}

func (r *Request) RequestContext(ctx context.Context, url string) (int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(req.Context(), 500*time.Millisecond)
	defer cancel()

	req = req.WithContext(ctx)
	client := http.DefaultClient

	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}

	return res.StatusCode, nil
}

func asyncRequests(ctx context.Context, count int, requests ...[]Request) []Status {
	ctx, cancel := context.WithTimeout(ctx, 60_000*time.Millisecond)
	pool := make(chan FutureResponse, 5)

	wg := &sync.WaitGroup{}
	for i, r := range requests {
		n := i
		wg.Add(1)
		go func(n int, requesters []Request) {
			defer wg.Done()
			log.Info(fmt.Sprintf("request #%d is pending", n))
			defer log.Info(fmt.Sprintf("request #%d is done", n))

			for _, request := range requesters {
				select {
				case <-ctx.Done():
					return
				default:
				}
				code, err := request.RequestContext(ctx, request.ID)
				pool <- FutureResponse{
					ID:       request.URL,
					HTTPCode: code,
					Error:    err,
				}
			}
		}(n, r)
	}

	go func() {
		wg.Wait()
		cancel()
		close(pool)
	}()

	values := make([]Status, 0, count)
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case value := <-pool:
			values = append(values, Status{
				ID:    value.ID,
				Code:  value.HTTPCode,
				Error: value.Error,
			})
		}
	}

	return values
}

func parts(streams int) int {
	return int(math.Round(float64(streams)/float64(threads) + 0.49))
}
