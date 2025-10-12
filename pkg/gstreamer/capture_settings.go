package gstreamer

type VideoCaptureSettings struct {
	Width     int
	Height    int
	Framerate int
}

type AudioCaptureSettings struct {
	SampleRate int
	Channels   int
	Format     string
}
