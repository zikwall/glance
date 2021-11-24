package httpstat

import (
	"context"
	"testing"

	"github.com/zikwall/glance"
)

func TestRequestContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("it should be successfully send requests", func(t *testing.T) {
		mockRequests := glance.Collection{
			Streams: map[string]glance.WorkerItem{
				"1": {ID: "1", URL: "https://github.com/"},
				"2": {ID: "2", URL: "https://google.com/"},
				"3": {ID: "3", URL: "https://news.yahoo.com/"},
			},
		}

		statuses := getHTTPStatuses(ctx, mockRequests, map[string]string{})
		if len(statuses) == 0 {
			t.Fatal("Failed. give empty response")
		}

		for _, status := range statuses {
			if status.Error != nil {
				t.Fatal(status.Error)
			}
			if status.Code != 200 {
				t.Fatalf("Failed, expect HTTP code 200 (OK), give %d", status.Code)
			}
		}
	})

	t.Run("parts test", func(t *testing.T) {
		n := parts(3)
		if n != 1 {
			t.Fatalf("Failed, expect 1 give %d", n)
		}

		n = parts(10)
		if n != 4 {
			t.Fatalf("Failed, expect 4 give %d", n)
		}

		n = parts(12)
		if n != 4 {
			t.Fatalf("Failed, expect 4 give %d", n)
		}

		n = parts(16)
		if n != 6 {
			t.Fatalf("Failed, expect 6 give %d", n)
		}
	})
}
