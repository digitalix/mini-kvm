package pkg

import (
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

	mouseChan chan KeyPress
	keyChan   chan MouseLocation

	isClosed atomic.Bool
}

func NewClient(id string, connection *webrtc.PeerConnection, logger zerolog.Logger, mouseChan chan KeyPress, keyChan chan MouseLocation) *Client {
	c := &Client{
		id:         id,
		connection: connection,
		mouseChan:  mouseChan,
		keyChan:    keyChan,
		logger:     logger,
	}

	c.connection.OnDataChannel(func(dc *webrtc.DataChannel) {
		logger.Println("data channel", dc.Label())
		switch dc.Label() {
		case "mouse":
			c.mouseChannel = dc
		case "keyboard":
			c.keyboardChannel = dc
		case "control":
			c.controlChannel = dc
		}
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
