package pkg

import (
	"context"
	"fmt"
)

type KeyPress struct {
	Key    rune
	IsDown bool
}

type KeyboardController struct {
	eventChan chan KeyPress
}

func NewKeyboardController(ctx context.Context) *KeyboardController {
	c := &KeyboardController{
		eventChan: make(chan KeyPress, 100),
	}

	go c.usbActionDispatcher(ctx)
	return c
}

func (m *KeyboardController) usbActionDispatcher(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-m.eventChan:
			fmt.Println("keyboard event received")
		}
	}
}

func (m *KeyboardController) EventChan() chan KeyPress {
	return m.eventChan
}
