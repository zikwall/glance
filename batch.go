package glance

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
