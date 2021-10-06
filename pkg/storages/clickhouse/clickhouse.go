package clickhouse

import (
	clickhousebuffer "github.com/zikwall/clickhouse-buffer"
	"github.com/zikwall/clickhouse-buffer/src/buffer"
	"github.com/zikwall/glance"
	"math"
	"time"
)

type Clickhouse struct {
	writer clickhousebuffer.Writer
}

func New(writer clickhousebuffer.Writer) *Clickhouse {
	ch := &Clickhouse{writer: writer}
	return ch
}

func (c *Clickhouse) ProcessFrameBatch(id string, frame glance.Frame) error {
	batch := &Batch{
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

	c.writer.WriteRow(batch)
	return nil
}

type Batch glance.Batch

func (b *Batch) Row() buffer.RowSlice {
	return buffer.RowSlice{
		b.StreamId,
		b.Bitrate,
		b.Frames,
		b.Height,
		b.Fps,
		b.Bytes,
		b.Seconds,
		b.KeyframeInterval,
		b.InsertTs,
		b.Date,
	}
}

func GetDefaultTableName() string {
	return "stream.metrics"
}

func GetTableColumns() []string {
	return []string{
		"stream_id",
		"bitrate",
		"frames",
		"height",
		"fps",
		"bytes",
		"seconds",
		"keyframe_interval",
		"insert_ts",
		"date",
	}
}
