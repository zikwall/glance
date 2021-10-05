package glance

import (
	"context"
)

type Fetcher interface {
	FetchStreams(ctx context.Context) (Collection, error)
}
