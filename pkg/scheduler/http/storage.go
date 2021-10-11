package http

type StatusWriter interface {
	Write(bucket *Bucket) error
}

type Bucket struct {
	StreamID   string
	Code       int
	InsertTS   string
	InsertDate string
}
