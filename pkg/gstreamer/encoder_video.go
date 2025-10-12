package gstreamer

import "C"
import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
	"github.com/go-gst/go-gst/gst/video"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type VideoEncoderSettings struct {
	Name           string
	EncoderType    EncoderType
	Width          int
	Height         int
	Framerate      int
	Bitrate        int64
	EncoderOptions map[string]string
}

func newKeyFrameEvent() *gst.Event {
	structure := gst.NewStructure("GstForceKeyUnit")
	if err := structure.SetValue("all-headers", true); err != nil {
		log.Panic().Err(err).Send()
	}

	return gst.NewCustomEvent(gst.EventTypeCustomUpstream, structure)
}

type VideoEncoder struct {
	logger zerolog.Logger

	encoderSettings VideoEncoderSettings
	captureSettings VideoCaptureSettings

	pipeline *gst.Pipeline

	appSink        *app.Sink
	appSrc         *app.Source
	encoderElement *gst.Element

	inputChan       chan *gst.Buffer
	outputChan      chan *media.Sample
	framesTillReset int

	producedFirstFrame atomic.Bool
	isStopping         atomic.Bool
	isRunning          atomic.Bool

	ctx       context.Context
	ctxCancel func()
}

func configureVideoEncoder(settings VideoEncoderSettings, captureSettings VideoCaptureSettings) (*gst.Pipeline, *app.Source, *app.Sink, error) {
	var sb strings.Builder
	if settings.EncoderOptions != nil {
		for k, v := range settings.EncoderOptions {
			if k == "num-ref-frames" {
				k = "num-Ref-Frames"
			}
			sb.WriteString(k)
			sb.WriteString("=")
			sb.WriteString(v)
			sb.WriteString(" ")
		}
	}

	videoInfo := video.NewInfo().
		WithFormat(video.FormatNV12, uint(captureSettings.Width), uint(captureSettings.Height)).
		WithFPS(gst.Fraction(int(settings.Framerate), 1))
	var pipeStr string
	switch settings.EncoderType {
	case EncoderTypeHEVC_MPP:
		pipeStr = fmt.Sprintf("appsrc is-live=True format=3 name=appsrc ! mpph265enc name=enc %s bps=%v ! h265parse ! appsink name=appsink sync=false", strings.TrimSpace(sb.String()), settings.Bitrate)
		break
	default:
		panic("unsupported codec")
	}
	fmt.Println(pipeStr)

	pipeline, err := gst.NewPipelineFromString(pipeStr)
	if err != nil {
		return nil, nil, nil, err
	}

	appsrcElement, err := pipeline.GetElementByName("appsrc")
	if err != nil {
		return nil, nil, nil, err
	}

	appsinkElement, err := pipeline.GetElementByName("appsink")
	if err != nil {
		return nil, nil, nil, err
	}

	appsrc := app.SrcFromElement(appsrcElement)
	appsink := app.SinkFromElement(appsinkElement)

	appsrc.SetCaps(videoInfo.ToCaps())

	return pipeline, appsrc, appsink, nil
}

// NewVideoEncoder
func NewVideoEncoder(settings VideoEncoderSettings, captureSettings VideoCaptureSettings, inputChan chan *gst.Buffer, outputChan chan *media.Sample) (*VideoEncoder, error) {
	pipeline, appsrc, appsink, err := configureVideoEncoder(settings, captureSettings)
	if err != nil {
		return nil, fmt.Errorf("failed to create videoEncoder: %w", err)
	}

	var encoderElement *gst.Element
	if settings.EncoderType == EncoderTypeHEVC_MPP {
		element, err := pipeline.GetElementByName("enc")
		if err != nil {
			return nil, err
		}

		encoderElement = element
	}

	return &VideoEncoder{
		logger:          log.With().Str("encoderType", "video").Str("encoderName", settings.Name).Logger(),
		encoderSettings: settings,
		captureSettings: captureSettings,
		pipeline:        pipeline,
		appSink:         appsink,
		appSrc:          appsrc,
		encoderElement:  encoderElement,
		inputChan:       inputChan,
		outputChan:      outputChan,
		framesTillReset: 14000,
	}, nil
}

func (e *VideoEncoder) RequestKeyframe() error {
	e.logger.Debug().Interface("encType", e.encoderSettings.EncoderType).Msg("requested keyframe")

	for !e.producedFirstFrame.Load() {
		time.Sleep(time.Millisecond * 100)
	}

	pad := e.encoderElement.GetStaticPad("src")
	if !pad.SendEvent(newKeyFrameEvent()) {
		panic("failed to send kf event")
	}

	return nil
}

func (e *VideoEncoder) Stop() {
	if e.isStopping.Swap(true) {
		return
	}

	defer e.isRunning.Store(false)
	e.logger.Println("stopping videoEncoder...")
	e.ctxCancel()
	e.pipeline.SendEvent(gst.NewEOSEvent())

	stopTime := time.Now()
	for e.pipeline.GetCurrentState() != gst.StateNull {
		time.Sleep(time.Millisecond * 100)
		e.logger.Warn().Str("state", e.pipeline.GetCurrentState().String()).Msg("waiting to stop...")
		if time.Now().Sub(stopTime).Seconds() > 30 {
			e.logger.Fatal().Msg("stop timed out")
		}
	}

	mainElements, err := e.pipeline.GetElements()
	if err != nil {
		panic(err)
	}

	for _, el := range mainElements {
		if err := el.BlockSetState(gst.StateNull); err != nil {
			panic(err)
		}

		if err := e.pipeline.Remove(el); err != nil {
			panic(err)
		}

	}

	e.pipeline = nil
	e.logger.Println("stopped videoEncoder...")
}

func (e *VideoEncoder) Start() {
	if e.isStopping.Load() {
		e.logger.Panic().Msg("something tried to start stopped videoEncoder")
	}

	defer e.isRunning.Store(true)

	go func() {
		defer e.logger.Warn().Msg("videoEncoder bus exit")
		bus := e.pipeline.GetPipelineBus()
		for {
			msg := bus.BlockPopMessage()
			switch msg.Type() {
			case gst.MessageStateChanged:
				if !strings.HasPrefix(msg.Source(), "pipeline") {
					break
				}

				oldState, newState := msg.ParseStateChanged()
				e.logger.Info().
					Str("source", msg.Source()).
					Str("oldState", oldState.String()).
					Str("newState", newState.String()).
					Msg("state changed")
				break
			case gst.MessageEOS:
				e.logger.Debug().Msg("videoEncoder pipeline EOS")
				if err := e.pipeline.SetState(gst.StateNull); err != nil {
					e.logger.Panic().Msg("failed to set videoEncoder state to null")
				}
				return
			case gst.MessageError:
				e.logger.Fatal().Err(msg.ParseError()).Msg("videoEncoder pipeline failed")
				return
			}
		}
	}()
	if err := e.pipeline.SetState(gst.StatePlaying); err != nil {
		e.logger.Panic().Err(err).Msg("failed to start videoEncoder")
	}

	e.ctx, e.ctxCancel = context.WithCancel(context.Background())
	stopChan := make(chan struct{}, 1)

	go func() {
		defer func() {
			stopChan <- struct{}{}
			if !e.isStopping.Load() {
				e.logger.Println("videoEncoder output routine rip")
			}
		}()
		firstFrame := true
		lastFrameTime := time.Now()
		for {
			sample := e.appSink.PullSample()
			if e.appSink.IsEOS() || sample == nil {
				if !e.isStopping.Load() {
					e.logger.Println("pull err", e.appSink.IsEOS(), e.appSink.GetCurrentState(), sample == nil)
				}
				break
			}

			now := time.Now()
			diff := now.Sub(lastFrameTime).Milliseconds()
			lastFrameTime = now
			_ = diff

			buffer := sample.GetBuffer()
			if buffer == nil {
				break
			}

			if firstFrame {
				firstFrame = false
				e.producedFirstFrame.Store(true)
			}

			duration := time.Duration(0)
			if buffer.Duration() != gst.ClockTimeNone {
				duration = *buffer.Duration().AsDuration()
			}

			if len(e.outputChan) == cap(e.outputChan) {
				e.logger.Warn().Msg("videoEncoder videoOutputChan is full")
				continue
			}

			flags := buffer.GetFlags()
			isKeyframe := (flags&gst.BufferFlagHeader) != 0 || !((flags & gst.BufferFlagDeltaUnit) != 0)
			if (e.encoderSettings.Name == "med" || e.encoderSettings.Name == "high") && isKeyframe {
				e.logger.Println(">>>>>>>> kf generated", e.encoderSettings.Name)
			}

			if e.isStopping.Load() {
				continue
			}

			outputChan := e.outputChan
			if len(e.outputChan) == cap(e.outputChan) {
				e.logger.Println("outputchan full", len(e.outputChan))
			}

			outputChan <- &media.Sample{Data: buffer.Bytes(), Timestamp: time.Unix(int64(buffer.PresentationTimestamp()), 0), Duration: duration, Metadata: SampleMetadata{IsKeyFrame: isKeyframe, Source: e.encoderSettings.Name, MediaType: MediaTypeVideo}}
		}
	}()

	go func() {
		defer func() {
			if !e.isStopping.Load() {
				e.logger.Println("videoEncoder input routine rip")
			}
			fmt.Println("input dead")
		}()
		for {
			select {
			case <-e.ctx.Done():
				return
			case buffer := <-e.inputChan:
				if buffer == nil {
					e.logger.Println("nil frame received, stopping")
					return
				}

				e.appSrc.PushBuffer(buffer)
				buffer.Unref()
				break
			}

		}
	}()

	for e.pipeline.GetCurrentState() != gst.StatePaused {
		e.logger.Println("waiting for state change...", e.pipeline.GetCurrentState().String())
		time.Sleep(time.Millisecond * 100)
	}
}

func (e *VideoEncoder) IsRunning() bool {
	return e.isRunning.Load()
}

func (e *VideoEncoder) InputChan() chan *gst.Buffer {
	return e.inputChan
}
