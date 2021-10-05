package metric

import (
	"github.com/zikwall/clickhouse-buffer/src/buffer"
)

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
