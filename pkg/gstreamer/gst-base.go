package gstreamer

import "C"
import (
	"errors"
	"fmt"
	"mini-kvm/pkg/concurrents"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-gst/go-gst/gst"
	"github.com/go-gst/go-gst/gst/app"
	"github.com/rs/zerolog"
)

var ErrPipeAlreadyRunning = errors.New("pipe is already running")

type BaseOption func(bc *gstBase) error

func WithOnEOSHandler(onEndOfStreamHandler func()) BaseOption {
	return func(bc *gstBase) error {
		bc.onEOSFunc = onEndOfStreamHandler
		return nil
	}
}

func WithAppSink(appSink *app.Sink) BaseOption {
	return func(bc *gstBase) error {
		bc.appSink = appSink
		return nil
	}
}

func WithAppSource(appSource *app.Source, inputChan chan *gst.Buffer) BaseOption {
	return func(bc *gstBase) error {
		cancelChan := make(chan struct{}, 1)
		bc.cancelResults = append(bc.cancelResults, cancelChan)
		bc.onStartForAppSrcFunc = func() {
			go func() {
				defer func() {
					if !bc.isStopping.Load() {
						bc.logger.Println("gstbase input routine rip")
					}
					cancelChan <- struct{}{}
				}()
				for {
					buffer := <-inputChan
					if buffer == nil {
						bc.logger.Println("nil frame received, stopping")
						break
					}

					//fmt.Println("audio buffer pushed")
					appSource.PushBuffer(buffer)
				}
			}()
		}

		return nil
	}
}

type gstBase struct {
	logger            zerolog.Logger
	pipeline          *gst.Pipeline
	appSink           *app.Sink
	appSrc            *app.Source
	pipelineMediaType MediaType

	encoders   *concurrents.Slice[Encoder]
	isStopping atomic.Bool
	isStarting atomic.Bool
	hasFailed  atomic.Bool
	cleanedUp  atomic.Bool

	onEOSFunc            func()
	onStartForAppSrcFunc func()
	onFailure            func(err error)

	GeneratedFramesCount int
	LastFrameTimestamp   time.Time
	FirstFrameTimestamp  time.Time

	cancelResults []chan struct{}
}

var id int

func newGstBase(logger zerolog.Logger, pipeline *gst.Pipeline, pipelineMediaType MediaType, options ...BaseOption) (*gstBase, error) {
	baseId := id
	id += 1
	bc := &gstBase{
		logger:            logger.With().Int("id", baseId).Logger(),
		encoders:          concurrents.NewSlice[Encoder](),
		pipeline:          pipeline,
		pipelineMediaType: pipelineMediaType,
		cancelResults:     make([]chan struct{}, 0),
	}

	for _, op := range options {
		if err := op(bc); err != nil {
			return nil, fmt.Errorf("failed to apply options: %w", err)
		}
	}

	return bc, nil
}

func (e *gstBase) cleanUp() {
	for e.pipeline.GetCurrentState() != gst.StateNull {
		time.Sleep(time.Second)
		e.logger.Println("waiting for pipeline to clean up...", e.pipeline.GetCurrentState().String())
	}

	mainElements, err := e.pipeline.GetElements()
	if err != nil {
		e.logger.Error().Err(err).Msg("failed to get elements to stop properly")
	} else {
		for _, el := range mainElements {
			if err := el.BlockSetState(gst.StateNull); err != nil {
				panic(err)
			}

			if err := e.pipeline.Remove(el); err != nil {
				panic(err)
			}
		}
	}

	e.pipeline = nil
	e.cleanedUp.Store(true)
	e.logger.Println("pipe destroyed")
}

func (e *gstBase) Stop() {
	if e.isStopping.Swap(true) {
		return
	}

	for _, cc := range e.cancelResults {
		<-cc
		close(cc)
	}

	if e.hasFailed.Load() {
		e.logger.Error().Msg("this pipeline has failed, no need for EOS")
		if e.pipeline != nil {
			if err := e.pipeline.SetState(gst.StateNull); err != nil {
				e.logger.Error().Err(err).Msg("pipe to nil err")
			}
		}
	} else {
		e.pipeline.SendEvent(gst.NewEOSEvent())
	}

	stopTime := time.Now()
	for !e.cleanedUp.Load() {
		time.Sleep(time.Millisecond * 100)
		e.logger.Println("waiting for clean up to complete")
		if time.Now().Sub(stopTime).Seconds() > 15 {
			e.logger.Fatal().Msg("stop timed out")
		}
	}

	e.logger.Println("clean up complete")
}

func (e *gstBase) AddEncoder(encoder Encoder) {
	if index := e.encoders.Index(encoder); index > 0 {
		return
	}

	e.encoders.Add(encoder)
}

func (e *gstBase) RemoveOutput(encoder Encoder) {
	index := e.encoders.Index(encoder)
	if index > 0 {
		e.encoders.Remove(index)
	}
}

func (e *gstBase) sendBuffer(buffer *gst.Buffer) {
	for encoder := range e.encoders.Iterate {
		if !encoder.IsRunning() {
			continue
		}

		if len(encoder.InputChan()) == cap(encoder.InputChan()) {
			e.logger.Warn().Str("mediaType", e.pipelineMediaType.String()).Int("len", len(encoder.InputChan())).Msg("inputChan is full")
			return
		}

		encoder.InputChan() <- buffer.Copy().Ref()
	}
}

func (e *gstBase) Start() error {
	if e.isStarting.Swap(true) {
		return ErrPipeAlreadyRunning
	}

	go e.gstBusLoop()
	if err := e.pipeline.SetState(gst.StatePlaying); err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	if e.appSink != nil {
		go e.appSinkPuller()
	}

	if e.onStartForAppSrcFunc != nil {
		e.onStartForAppSrcFunc()
	}

	return nil
}

func (e *gstBase) gstBusLoop() {
	defer e.cleanUp()
	bus := e.pipeline.GetPipelineBus()
	for {
		msg := bus.Pop()
		if msg == nil {
			time.Sleep(time.Millisecond * 10)
			continue
		}

		switch msg.Type() {
		case gst.MessageStateChanged:
			if !strings.HasPrefix(msg.Source(), "pipeline") {
				break
			}

			prevState, newState := msg.ParseStateChanged()
			e.logger.Info().Str("source", msg.Source()).Str("new", newState.String()).Str("old", prevState.String()).Msg("state changed")
			break
		case gst.MessageEOS:
			e.logger.Println("gstbase eos")
			e.pipeline.SetState(gst.StateNull)
			if handler := e.onEOSFunc; handler != nil {
				handler()
			}

			return
		case gst.MessageError:
			e.pipeline.SetState(gst.StateNull)
			if e.hasFailed.Swap(true) {
				return
			}

			err := msg.ParseError()
			e.logger.Error().Err(err).Msg("pipeline failed")
			if handler := e.onFailure; handler != nil {
				go e.onFailure(err)
			}

			return
		default:
			//e.logger.Println("unknown event received:", msg.Type().String())
			break
		}
	}
}

func (e *gstBase) appSinkPuller() {
	startTime := time.Now()
	pipeline := e.pipeline
	if pipeline == nil {
		e.logger.Panic().Bool("cleanedUp", e.cleanedUp.Load()).Bool("hasFailed", e.hasFailed.Load()).Bool("isStopping", e.isStopping.Load()).Msg("appSinkPuller used nil pipeline")
		return
	}

	for pipeline.GetCurrentState() != gst.StatePlaying {
		time.Sleep(time.Second / 10)
		if e.cleanedUp.Load() {
			return
		}

		if time.Now().Sub(startTime).Seconds() > 15 && !e.hasFailed.Load() {
			e.logger.Panic().Bool("cleanedUp", e.cleanedUp.Load()).Bool("hasFailed", e.hasFailed.Load()).Bool("isStopping", e.isStopping.Load()).Msg("timed out with startup...")
		}
	}

	defer func() {
		if !e.isStopping.Load() && !e.hasFailed.Load() {
			e.logger.Println("output routine rip")
		}
	}()
	firstFrame := true
	var now time.Time
	for {
		now = time.Now()
		sample := e.appSink.PullSample()
		if sec := time.Now().Sub(now).Seconds(); sec > 1 {
			e.logger.Println("took", sec, "to generate frame")
		}

		if e.appSink.IsEOS() || sample == nil {
			if !e.isStopping.Load() && !e.hasFailed.Load() {
				e.logger.Println("pull err", e.appSink.IsEOS(), e.appSink.GetCurrentState(), sample == nil)
				panic("this is not supposed to happen")
			}

			break
		}

		buffer := sample.GetBuffer()
		if buffer == nil {
			break
		}

		e.LastFrameTimestamp = time.Now()
		e.GeneratedFramesCount += 1
		if firstFrame {
			e.FirstFrameTimestamp = time.Now()
			firstFrame = false
			e.logger.Println("first frame generated")
		}

		e.sendBuffer(buffer.Ref())
		buffer.Unref()
	}
}

func (e *gstBase) SetOnFailureHandler(handler func(err error)) {
	e.onFailure = handler
}
