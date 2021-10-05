package glance

type Collection struct {
	Streams map[string]WorkerItem
}

func (i Collection) Exist(key string) bool {
	_, ok := i.Streams[key]
	return ok
}
