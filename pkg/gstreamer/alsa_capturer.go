package gstreamer

import "C"
import (
	"errors"
	"fmt"

	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
	"github.com/rs/zerolog/log"
)

type AlsaCaptureSettings struct {
	AudioCaptureSettings

	Mode   AudioCaptureMode
	Device string
}

type AlsaCapturer struct {
	*gstBase
	settings AlsaCaptureSettings
}

func configureAlsaCapturer(settings AlsaCaptureSettings) (*gst.Pipeline, *app.Sink, error) {
	var pipelineStr string
	if settings.Mode == AudioCaptureModeXRAW {
		pipelineStr = fmt.Sprintf("alsasrc device=%s ! audio/x-raw, rate=%d, format=%s, layout=interleaved, channels=%d ! audioconvert ! appsink name=appsink", settings.Device, settings.SampleRate, settings.Format, settings.Channels)
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

func NewAlsaCapturer(settings AlsaCaptureSettings) (*AlsaCapturer, error) {
	pipeline, appsink, err := configureAlsaCapturer(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create alsa audioCapturer: %w", err)
	}

	logger := log.With().Str("type", "alsa").Str("source", settings.Device).Logger()
	base, err := newGstBase(logger, pipeline, MediaTypeAudio, WithAppSink(appsink))
	if err != nil {
		return nil, fmt.Errorf("failed to create alsa audioCapturer: %w", err)
	}

	return &AlsaCapturer{
		gstBase:  base,
		settings: settings,
	}, nil
}

func (e *AlsaCapturer) CaptureSettings() AudioCaptureSettings {
	return e.settings.AudioCaptureSettings
}

func (e *AlsaCapturer) Stop() {
	e.gstBase.Stop()
}

func (e *AlsaCapturer) Start() error {
	if err := e.gstBase.Start(); err != nil {
		return err
	}

	return nil
}
