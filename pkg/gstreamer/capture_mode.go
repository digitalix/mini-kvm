package gstreamer

type VideoCaptureMode uint8

const (
	VideoCaptureModeUNKOWN VideoCaptureMode = iota
	VideoCaptureModeMJPEG
	VideoCaptureModeXRAW
)

type AudioCaptureMode uint8

const (
	AudioCaptureModeUNKOWN AudioCaptureMode = iota
	AudioCaptureModeXRAW
)
