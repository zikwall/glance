package clickhouse

import (
	clickhousebuffer "github.com/zikwall/clickhouse-buffer"
	"github.com/zikwall/clickhouse-buffer/src/buffer"
	"github.com/zikwall/glance/pkg/scheduler/httpstat"
)

type writerImpl struct {
	writer clickhousebuffer.Writer
}

func NewHTTPStatWriter(writer clickhousebuffer.Writer) httpstat.StatusWriter {
	ch := &writerImpl{writer: writer}
	return ch
}

func (c *writerImpl) Write(bucket httpstat.Bucket) error {
	alias := BucketAlias(bucket)
	c.writer.WriteRow(&alias)
	return nil
}

type BucketAlias httpstat.Bucket

func (b *BucketAlias) Row() buffer.RowSlice {
	return buffer.RowSlice{
		b.StreamID, b.Code, b.InsertTS, b.InsertDate,
	}
}

func GetTableColumns() []string {
	return []string{"stream_id", "code", "insert_ts", "insert_date"}
}
