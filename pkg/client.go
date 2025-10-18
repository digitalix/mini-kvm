package pkg

import (
	"encoding/json"
	"fmt"
	"sync/atomic"

	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog"
)

type Client struct {
	logger zerolog.Logger

	id         string
	connection *webrtc.PeerConnection

	mouseChannel    *webrtc.DataChannel
	keyboardChannel *webrtc.DataChannel
	controlChannel  *webrtc.DataChannel

	mouseChan chan MouseEvent
	keyChan   chan KeyPressEvent

	isClosed atomic.Bool
}

func NewClient(id string, connection *webrtc.PeerConnection, logger zerolog.Logger, mouseChan chan MouseEvent, keyChan chan KeyPressEvent) *Client {
	c := &Client{
		id:         id,
		connection: connection,
		mouseChan:  mouseChan,
		keyChan:    keyChan,
		logger:     logger,
	}

	c.connection.OnDataChannel(func(dc *webrtc.DataChannel) {
		logger.Println("on data channel", dc.Label())
		switch dc.Label() {
		case "mouse":
			c.mouseChannel = dc
		case "keyboard":
			c.keyboardChannel = dc
		case "control":
			c.controlChannel = dc
		}
		dc.OnOpen(func() {
			logger.Println("on open data channel", dc.Label())
		})
		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			c.onDataChannelMessage(dc, msg)
		})
	})

	return c
}

func (c *Client) Id() string {
	return c.id
}

func (c *Client) onDataChannelMessage(dc *webrtc.DataChannel, message webrtc.DataChannelMessage) {
	switch dc.Label() {
	case "mouse":
		var m MouseEvent
		if err := json.Unmarshal(message.Data, &m); err != nil {
			c.logger.Error().Err(err).Msg("failed to unmarshal mouse location")
			break
		}

		c.mouseChan <- m
		break
	case "keyboard":
		var k KeyPressEvent
		if err := json.Unmarshal(message.Data, &k); err != nil {
			c.logger.Error().Err(err).Msg("failed to unmarshal key press")
		}

		c.keyChan <- k
		break
	}
	c.logger.Println("onDataChannelMessage", dc.Label())
}

func (c *Client) Close() error {
	if c.isClosed.Swap(true) {
		return nil
	}

	if err := c.connection.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	return nil
}
