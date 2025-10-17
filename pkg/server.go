package pkg

import (
	"context"
	"fmt"
	"mini-kvm/pkg/concurrents"
	"mini-kvm/pkg/gstreamer"

	"github.com/google/uuid"
	"github.com/pion/interceptor"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/rs/zerolog/log"
)

var peerConnectionConfiguration = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	},
}

func toPtr[T any](t T) *T {
	return &t
}

type Server struct {
	webrtcAPI *webrtc.API
	clients   concurrents.Map[string, *Client]

	keyboardController *KeyboardController
	mouseController    *MouseController

	videoTrack *webrtc.TrackLocalStaticSample
	audioTrack *webrtc.TrackLocalStaticSample
}

func NewServer(ctx context.Context, mediaChan chan *media.Sample) (*Server, error) {
	api, err := configureWebRTCApi()
	if err != nil {
		return nil, fmt.Errorf("failed to configure webrtc api: %w", err)
	}

	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
		MimeType: webrtc.MimeTypeH265,
	}, "video", "mkvm")
	if err != nil {
		return nil, fmt.Errorf("failed to create video track: %w", err)
	}

	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
		MimeType: webrtc.MimeTypeOpus,
	}, "audio", "mkvm")
	if err != nil {
		return nil, fmt.Errorf("failed to create audio track: %w", err)
	}

	keyboardController := NewKeyboardController(ctx, "/dev/hidg0")
	mouseController := NewMouseController(ctx)

	server := &Server{
		webrtcAPI:          api,
		keyboardController: keyboardController,
		mouseController:    mouseController,
		videoTrack:         videoTrack,
		audioTrack:         audioTrack,
	}

	go server.mediaDistribution(ctx, mediaChan)
	return server, nil
}

func (s *Server) mediaDistribution(ctx context.Context, mediaChan chan *media.Sample) {
	var err error
	for {
		select {
		case <-ctx.Done():
			return
		case media := <-mediaChan:
			metadata := media.Metadata.(gstreamer.SampleMetadata)
			switch metadata.MediaType {
			case gstreamer.MediaTypeVideo:
				err = s.videoTrack.WriteSample(*media)
			case gstreamer.MediaTypeAudio:
				err = s.audioTrack.WriteSample(*media)
			default:
				log.Panic().Msg("unknown media type")
			}

			if err != nil {
				log.Error().Err(err).Msg("failed to write sample")
			}
		}
	}
}

func configureWebRTCApi() (api *webrtc.API, err error) {
	media := &webrtc.MediaEngine{}
	if err = media.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeH265,
			ClockRate:    90000,
			Channels:     0,
			SDPFmtpLine:  "",
			RTCPFeedback: nil,
		},
		PayloadType: 96,
	}, webrtc.RTPCodecTypeVideo); err != nil {
		return nil, fmt.Errorf("failed to register video codec: %w", err)
	}

	if err = media.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeOpus,
			ClockRate:    48000,
			Channels:     2,
			SDPFmtpLine:  "",
			RTCPFeedback: nil,
		},
		PayloadType: 97,
	}, webrtc.RTPCodecTypeAudio); err != nil {
		return nil, fmt.Errorf("failed to register audio codec: %w", err)
	}

	ir := &interceptor.Registry{}
	if err = webrtc.RegisterDefaultInterceptors(media, ir); err != nil {
		return nil, fmt.Errorf("failed to register default interceptors: %w", err)
	}

	if err = webrtc.ConfigureTWCCHeaderExtensionSender(media, ir); err != nil {
		return nil, fmt.Errorf("failed to configure twcc: %w", err)
	}

	return webrtc.NewAPI(webrtc.WithMediaEngine(media), webrtc.WithInterceptorRegistry(ir)), nil
}

func (s *Server) CreateClient(offer webrtc.SessionDescription) (string, *webrtc.SessionDescription, error) {
	peerConnection, err := s.webrtcAPI.NewPeerConnection(peerConnectionConfiguration)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create client: %w", err)
	}

	for _, track := range []webrtc.TrackLocal{s.videoTrack, s.audioTrack} {
		sender, err := peerConnection.AddTrack(track)
		if err != nil {
			return "", nil, fmt.Errorf("failed to add track: %w", err)
		}

		go rtcpDummyReader(sender)
	}

	{
		_, err = peerConnection.CreateDataChannel("dummy", nil)
		if err != nil {
			return "", nil, fmt.Errorf("failed to create data channel for control: %w", err)
		}
	}

	id := uuid.NewString()
	logger := log.With().Str("id", id).Logger()
	client := NewClient(id, peerConnection, logger, s.mouseController.EventChan(), s.keyboardController.EventChan())
	s.clients.Set(id, client)
	peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		logger.Info().Str("state", state.String()).Msg("connection state changed")
		switch state {
		case webrtc.PeerConnectionStateConnected:

		case webrtc.PeerConnectionStateDisconnected, webrtc.PeerConnectionStateFailed:
			if err := client.Close(); err != nil {
				logger.Error().Err(err).Msg("failed to close client")
			}

			s.clients.Delete(id)
		}
	})

	if err = peerConnection.SetRemoteDescription(offer); err != nil {
		return "", nil, fmt.Errorf("failed to set remote desc: %w", err)
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create answer: %w", err)
	}

	if err = peerConnection.SetLocalDescription(answer); err != nil {
		return "", nil, fmt.Errorf("failed to set local desc: %w", err)
	}

	<-gatherComplete
	return client.Id(), &answer, nil
}

func rtcpDummyReader(sender *webrtc.RTPSender) {
	rtcpBuf := make([]byte, 1500)
	for {
		if _, _, rtcpErr := sender.Read(rtcpBuf); rtcpErr != nil {
			return
		}
	}
}
