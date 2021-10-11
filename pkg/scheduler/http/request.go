package http

import (
	"context"
	"math"
	"net/http"
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
	pool := make(chan FutureResponse, count)

	defer func() {
		cancel()
		close(pool)
	}()

	for i, r := range requests {
		n := i

		go func(n int, requesters []Request) {
			for _, request := range requesters {
				code, err := request.RequestContext(ctx, request.ID)

				pool <- FutureResponse{
					ID:       request.URL,
					HTTPCode: code,
					Error:    err,
				}
			}
		}(n, r)
	}

	values := make([]Status, 0, count)
	for value := range pool {
		values = append(values, Status{
			ID:    value.ID,
			Code:  value.HTTPCode,
			Error: value.Error,
		})
	}

	return values
}

func parts(streams int) int {
	return int(math.Round(float64(streams)/float64(threads) + 0.49))
}
