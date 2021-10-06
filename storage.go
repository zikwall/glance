package glance

type Storage interface {
	ProcessFrameBatch(id string, frame Frame) error
}
