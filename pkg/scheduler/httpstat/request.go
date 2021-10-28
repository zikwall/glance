package httpstat

import (
	"context"
	"fmt"
	"github.com/zikwall/glance/pkg/log"
	"math"
	"net/http"
	"sync"
	"time"
)

type future struct {
	id   string
	code int
	err  error
}

type request struct {
	id  string
	url string
}

func (r *request) RequestContext(ctx context.Context, url string, headers map[string]string) (int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(req.Context(), 1000*time.Millisecond)
	defer cancel()

	req = req.WithContext(ctx)
	client := http.DefaultClient

	for name, value := range headers {
		req.Header.Set(name, value)
	}

	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}

	return res.StatusCode, nil
}

func asyncRequests(ctx context.Context, count int, headers map[string]string, requests ...[]request) []Status {
	ctx, cancel := context.WithTimeout(ctx, 60_000*time.Millisecond)
	pool := make(chan future, 5)

	wg := &sync.WaitGroup{}
	for i, r := range requests {
		n := i
		wg.Add(1)
		go func(n int, requesters []request) {
			defer wg.Done()
			log.Info(fmt.Sprintf("request group #%d is pending", n))
			defer log.Info(fmt.Sprintf("request group #%d is done", n))

			for _, request := range requesters {
				select {
				case <-ctx.Done():
					return
				default:
				}
				code, err := request.RequestContext(ctx, request.url, headers)
				pool <- future{
					id:   request.id,
					code: code,
					err:  err,
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
				ID:    value.id,
				Code:  value.code,
				Error: value.err,
			})
		}
	}

	return values
}

func parts(streams int) int {
	return int(math.Round(float64(streams)/float64(threads) + 0.49))
}
