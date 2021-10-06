package glance

import (
	"context"
	"math"
	"time"
)

// Storage basic interface that implements data saving
type Storage interface {
	ProcessFrameBatch(batch Batch) error
}

// Fetcher interface that implements formatting of screenshot links
type Fetcher interface {
	FetchStreams(ctx context.Context) (Collection, error)
}

// Worker A worker interface that provides a synchronous Perform method
// for the ability to implement custom processing of an asynchronous task.
type Worker interface {
	Perform(context.Context, WorkerStream)
	Name() string
	Label() string
}

type WorkerStream interface {
	ID() string
	URL() string
}

// Batch type is the main structure for generating and sending data to the storage
type Batch struct {
	Date             string  `json:"date"`
	InsertTs         string  `json:"insert_ts"`
	Fps              float64 `json:"fps"`
	Bitrate          float64 `json:"bitrate"`
	StreamId         string  `json:"stream_id"`
	Seconds          float64 `json:"seconds"`
	Bytes            uint64  `json:"bytes"`
	Frames           uint64  `json:"frames"`
	Height           uint64  `json:"height"`
	KeyframeInterval uint64  `json:"keyframe_interval"`
}

// Frame structure for counting frames and their parameters
type Frame struct {
	Frames           int
	Bytes            int
	Seconds          float64
	Height           int
	KeyframeInterval int
}

func (f *Frame) IncreasingContinue(bytes int) {
	f.Bytes += bytes
	f.Frames += 1
	f.KeyframeInterval += 1
}

func (f *Frame) Cleanup() {
	f.Bytes = 0
	f.Frames = 0
	f.Seconds = 0
	f.Height = 0
	f.KeyframeInterval = 0
}

const bitsInBytes = 8
const bytesInKb = 1024

func CreateBatch(id string, frame Frame) Batch {
	batch := Batch{
		StreamId:         id,
		Seconds:          frame.Seconds,
		Bytes:            uint64(frame.Bytes),
		Frames:           uint64(frame.Frames),
		Height:           uint64(frame.Height),
		KeyframeInterval: uint64(frame.KeyframeInterval),
	}

	batch.Date = date(time.Now())
	batch.InsertTs = datetime(time.Now())

	// calculate
	fps := float64(batch.Frames) / batch.Seconds
	batch.Fps = math.Round(fps*100) / 100
	bitrate := float64(frame.Bytes*bitsInBytes) / (frame.Seconds * bytesInKb)
	batch.Bitrate = math.Round(bitrate*1000) / 1000
	return batch
}

// Collection structure for unification of transmitted data between workers
type Collection struct {
	Streams map[string]WorkerItem
}

func (i Collection) Exist(key string) bool {
	_, ok := i.Streams[key]
	return ok
}

const DateTimeFormat = "2006-01-02 15:04:05"
const DateFormat = "2006-01-02"

func date(t time.Time) string {
	return format(t, DateFormat)
}

func datetime(t time.Time) string {
	return format(t, DateTimeFormat)
}

func format(t time.Time, format string) string {
	location, _ := time.LoadLocation("Europe/Moscow")
	return t.In(location).Format(format)
}
