package glance

type Frame struct {
	Frames           int
	Bytes            int
	Seconds          float64
	Height           int
	KeyframeInterval int
}

func (f *Frame) IncreasingContinue(bytes int) {
	f.Bytes += bytes
	f.Frames += 1
	f.KeyframeInterval += 1
}

func (f *Frame) Cleanup() {
	f.Bytes = 0
	f.Frames = 0
	f.Seconds = 0
	f.Height = 0
	f.KeyframeInterval = 0
}
