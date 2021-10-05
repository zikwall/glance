### Glance

Glance is a portable, event-driven, package for building monitoring applications for your streams (HLS, RTMP, SRS, MPEG-DASH, RTSP, WebRTC).

#### Example usage

```go
func main() {
	...
	// create async Clickhouse writer
	writerApi := buffer.Client().Writer(
		clickhousebuffer.View{
			Name:    "stream.metrics",
			Columns: metric.GetTableColumns(),
		},
		memory.NewBuffer(
			buffer.Client().Options().BatchSize(),
		),
	)

	workers := make([]glance.Worker, 0, 2)
	// add screenshot worker
	workers = append(workers, screenshot.New(
		"screenshot", "user:pass@webdav-server.com", &screenshot.SimpleUrlFormatter{},
	))
	// add metric writer to Clickhouse worker
	workers = append(workers, metric.New("metric", writerApi))
	workstation := glance.New(monitoring.Context(), workers...)

	go func() {
		screenWorker, _ := workstation.Workspace("screenshot")
		metricWorker, _ := workstation.Workspace("metric")

		scheduler.New(mockfetcher.New()).RunContext(monitoring.Context(),
			scheduler.Options{
				WorkspaceScreenshot: screenWorker,
				WorkspaceMetrics:    metricWorker,
				RefreshInterval:     time.Duration(10 * time.Minute),
			},
		)
	}()
	...
	await()
}
```