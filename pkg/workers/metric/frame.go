package metric

type Frame struct {
	frames            int
	bytes             int
	seconds           float64
	height            int
	keyframe_interval int
}

func (f *Frame) increasingContinue(bytes int) {
	f.bytes += bytes
	f.frames += 1
	f.keyframe_interval += 1
	//f.seconds = math.Ceil((f.seconds+time)*1000000) / 1000000
}

func (f *Frame) cleanup() {
	f.bytes = 0
	f.frames = 0
	f.seconds = 0
	f.height = 0
	f.keyframe_interval = 0
}
