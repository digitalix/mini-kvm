package pkg

import (
	"context"
	"fmt"
)

type MouseLocation struct {
	X, Y uint16
}

type MouseController struct {
	eventChan chan MouseLocation
}

func NewMouseController(ctx context.Context) *MouseController {
	c := &MouseController{
		eventChan: make(chan MouseLocation, 100),
	}

	go c.usbActionDispatcher(ctx)
	return c
}

func (m *MouseController) usbActionDispatcher(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-m.eventChan:
			fmt.Println("mouse event received")
		}
	}
}

func (m *MouseController) EventChan() chan MouseLocation {
	return m.eventChan
}
