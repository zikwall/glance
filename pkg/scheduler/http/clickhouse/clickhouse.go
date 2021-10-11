package clickhouse

import (
	clickhousebuffer "github.com/zikwall/clickhouse-buffer"
	"github.com/zikwall/clickhouse-buffer/src/buffer"
)

type Clickhouse struct {
	writer clickhousebuffer.Writer
}

func NewWriter(writer clickhousebuffer.Writer) *Clickhouse {
	ch := &Clickhouse{writer: writer}
	return ch
}

func (c *Clickhouse) Write(bucket *Bucket) error {
	c.writer.WriteRow(bucket)
	return nil
}

type Bucket struct {
	StreamID   string
	Code       int
	InsertTS   string
	InsertDate string
}

func (b *Bucket) Row() buffer.RowSlice {
	return buffer.RowSlice{
		b.StreamID, b.Code, b.InsertTS, b.InsertDate,
	}
}
