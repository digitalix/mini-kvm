package pkg

import (
	"context"
	"errors"
	"fmt"
	"mini-kvm/pkg/gstreamer"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/go-gst/go-gst/gst"
	"github.com/pion/webrtc/v4/pkg/media"
)

var httpServer = &http.Server{
	Addr:         ":8080",
	ReadTimeout:  10 * time.Second,
	WriteTimeout: 10 * time.Second,
	IdleTimeout:  60 * time.Second,
}

func Run(ctx context.Context) error {
	inputChan := make(chan *gst.Buffer, 30)
	outputChan := make(chan *media.Sample, 100)
	captureSettings := gstreamer.V4L2CaptureSettings{
		VideoCaptureSettings: gstreamer.VideoCaptureSettings{
			Width:     1920,
			Height:    1080,
			Framerate: 30,
		},
		Mode:   gstreamer.VideoCaptureModeMJPEG,
		Device: "/dev/video0",
	}
	videoCapture, err := gstreamer.NewV4L2Capturer(captureSettings)
	if err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	videoEncoder, err := gstreamer.NewVideoEncoder(gstreamer.VideoEncoderSettings{
		Name:        "out",
		EncoderType: gstreamer.EncoderTypeHEVC_MPP,
		Width:       captureSettings.Width,
		Height:      captureSettings.Height,
		Framerate:   captureSettings.Framerate,
		Bitrate:     2_000_000,
		EncoderOptions: map[string]string{
			"rc-mode":     "vbr",
			"max-pending": "4",
			"header-mode": "each-idr",
			"gop":         "60",
		},
	}, captureSettings.VideoCaptureSettings, inputChan, outputChan)
	if err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	videoCapture.AddEncoder(videoEncoder)
	if err := videoCapture.Start(); err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	videoEncoder.Start()
	server, err := NewServer(ctx, outputChan)

	httpHandler := HttpHandler{
		server: server,
	}
	http.HandleFunc("/connect", httpHandler.whepHandler)

	go func() {
		log.Printf("Server starting on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("ctx done")
			return nil
		case sample := <-outputChan:
			if err := server.videoTrack.WriteSample(*sample); err != nil {
				log.Error().Err(err).Msg("failed to write sample")
			}
		}
	}

}
