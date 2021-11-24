package clickhouse

import (
	clickhousebuffer "github.com/zikwall/clickhouse-buffer"
	"github.com/zikwall/clickhouse-buffer/src/buffer"

	"github.com/zikwall/glance"
)

type Clickhouse struct {
	writer clickhousebuffer.Writer
}

func New(writer clickhousebuffer.Writer) *Clickhouse {
	ch := &Clickhouse{writer: writer}
	return ch
}

func (c *Clickhouse) ProcessFrameBatch(batch *glance.Batch) error {
	bucket := Batch(*batch)
	c.writer.WriteRow(&bucket)

	return nil
}

type Batch glance.Batch

// nolint:(typecheck) // its OK
func (b *Batch) Row() buffer.RowSlice {
	return buffer.RowSlice{
		b.StreamID,
		b.Bitrate,
		b.Frames,
		b.Height,
		b.Fps,
		b.Bytes,
		b.Seconds,
		b.KeyframeInterval,
		b.InsertTS,
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
