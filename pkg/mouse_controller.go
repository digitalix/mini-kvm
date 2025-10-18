package pkg

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

type MouseEventKind uint8

const (
	MouseMovedEventKind MouseEventKind = iota
	MouseButtonEventKind
	MouseWheelEventKind
)

type MouseEvent struct {
	Kind MouseEventKind `json:"a"`

	//MouseMovedEventKind
	X uint16 `json:"x"`
	Y uint16 `json:"y"`

	//MouseButtonEventKind
	Button JSMouseButton `json:"b"`
	IsDown bool          `json:"d"`

	WheelX int8 `json:"wx"`
	WheelY int8 `json:"wy"`
}

type MouseController struct {
	device                    *os.File
	eventChan                 chan MouseEvent
	screenWidth, screenHeight int
}

func NewMouseController(ctx context.Context, devicePath string, screenWidth, screenHeight int) *MouseController {
	c := &MouseController{
		eventChan:    make(chan MouseEvent, 100),
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
	}

	device, err := os.OpenFile(devicePath, os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal().Err(err).Str("path", devicePath).Msg("failed to open device")
	}

	c.device = device

	go c.usbActionDispatcher(ctx)
	return c
}

func (m *MouseController) screenToHID(screenX, screenY uint16) (uint16, uint16) {
	hidX := uint16((float64(screenX) / float64(m.screenWidth)) * 32767)
	hidY := uint16((float64(screenY) / float64(m.screenHeight)) * 32767)
	return hidX, hidY
}
func (m *MouseController) usbActionDispatcher(ctx context.Context) {
	defer m.device.Close()
	lastX, lastY := uint16(0), uint16(0)
	pressedButtons := make(map[MouseButton]bool)
	buttons := ButtonNone
	for {
		select {
		case <-ctx.Done():
			return
		case ml := <-m.eventChan:
			switch ml.Kind {
			case MouseMovedEventKind:
				ml.X, ml.Y = m.screenToHID(ml.X, ml.Y)
				lastX, lastY = ml.X, ml.Y
				if err := m.sendReport(ml.X, ml.Y, buttons, 0); err != nil {
					log.Error().Err(err).Msg("failed to sendReport mouse location")
				}
			case MouseButtonEventKind:
				if ml.IsDown {
					pressedButtons[ml.Button.ToMouseButton()] = true
				} else {
					delete(pressedButtons, ml.Button.ToMouseButton())
				}

				buttons = ButtonNone
				for k, v := range pressedButtons {
					if v {
						buttons |= k
					}
				}

				if err := m.sendReport(lastX, lastY, buttons, 0); err != nil {
					log.Error().Err(err).Msg("failed to sendReport mouse location")
				}
			case MouseWheelEventKind:
				if err := m.sendReport(ml.X, ml.Y, buttons, ml.WheelX); err != nil {
					log.Error().Err(err).Msg("failed to sendReport wheel")
				}
			}
			fmt.Println("mouse event received <- ", ml.X, ml.Y)
		}
	}
}

/*
Report Structure for HID Touch Screen

Byte 0: Report ID (always 0x01)
Byte 1: Button state (bit flags)
Byte 2: X coordinate (low byte)
Byte 3: X coordinate (high byte)
Byte 4: Y coordinate (low byte)
Byte 5: Y coordinate (high byte)
Byte 6: Wheel (signed byte)
*/
func (m *MouseController) sendReport(x, y uint16, buttons MouseButton, wheel int8) error {
	report := make([]byte, 7)
	report[0] = 0x01
	report[1] = byte(buttons)
	binary.LittleEndian.PutUint16(report[2:4], x)
	binary.LittleEndian.PutUint16(report[4:6], y)
	report[6] = byte(wheel)

	_, err := m.device.Write(report)
	return err
}

func (m *MouseController) EventChan() chan MouseEvent {
	return m.eventChan
}
