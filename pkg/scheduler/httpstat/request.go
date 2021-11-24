package httpstat

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/zikwall/glance/pkg/log"
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
	reqCtx, cancel := context.WithTimeout(ctx, 1000*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", url, http.NoBody)
	if err != nil {
		return 0, err
	}

	for name, value := range headers {
		req.Header.Set(name, value)
	}

	res, err := (http.DefaultClient).Do(req)
	if err != nil {
		return 0, err
	}

	defer func() {
		_ = res.Body.Close()
	}()

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
