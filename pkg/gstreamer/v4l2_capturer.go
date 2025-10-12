package gstreamer

import "C"
import (
	"errors"
	"fmt"

	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
	"github.com/rs/zerolog/log"
)

type V4L2CaptureSettings struct {
	VideoCaptureSettings

	Mode   VideoCaptureMode
	Device string
}

type V4L2Capturer struct {
	*gstBase
	settings V4L2CaptureSettings
}

func configureV4L2Capturer(settings V4L2CaptureSettings) (*gst.Pipeline, *app.Sink, error) {
	var pipelineStr string
	if settings.Mode == VideoCaptureModeMJPEG {
		pipelineStr = fmt.Sprintf("v4l2src device=%s ! image/jpeg, width=%d, height=%d, framerate=%v/1 ! jpegparse ! mppjpegdec ! video/x-raw, format=NV12 ! appsink name=appsink", settings.Device, settings.Width, settings.Height, settings.Framerate)
	} else {
		return nil, nil, errors.New("unknown capture mode")
	}

	pipeline, err := gst.NewPipelineFromString(pipelineStr)
	if err != nil {
		return nil, nil, err
	}

	appsinkElement, err := pipeline.GetElementByName("appsink")
	if err != nil {
		return nil, nil, err
	}

	appsink := app.SinkFromElement(appsinkElement)

	return pipeline, appsink, nil
}

func NewV4L2Capturer(settings V4L2CaptureSettings) (*V4L2Capturer, error) {
	pipeline, appsink, err := configureV4L2Capturer(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create v4l2 videoCapturer: %w", err)
	}

	logger := log.With().Str("type", "v4l2").Str("source", settings.Device).Logger()
	base, err := newGstBase(logger, pipeline, MediaTypeVideo, WithAppSink(appsink))
	if err != nil {
		return nil, fmt.Errorf("failed to create v4l2 videoCapturer: %w", err)
	}

	return &V4L2Capturer{
		gstBase:  base,
		settings: settings,
	}, nil
}

func (e *V4L2Capturer) CaptureSettings() VideoCaptureSettings {
	return e.settings.VideoCaptureSettings
}

func (e *V4L2Capturer) Stop() {
	e.gstBase.Stop()
}

func (e *V4L2Capturer) Start() error {
	if err := e.gstBase.Start(); err != nil {
		return err
	}

	return nil
}
