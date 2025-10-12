package gstreamer

import "github.com/go-gst/go-gst/gst"

type Encoder interface {
	IsRunning() bool
	InputChan() chan *gst.Buffer
}
