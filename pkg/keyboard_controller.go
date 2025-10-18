package pkg

import (
	"context"
	"fmt"
	"os"
	"slices"

	"github.com/rs/zerolog/log"
)

type KeyPressEvent struct {
	KeyCode JSKeyCode `json:"key_code"`
	IsDown  bool      `json:"is_down"`
}

type KeyboardController struct {
	device      *os.File
	pressedKeys map[JSKeyCode]bool
	eventChan   chan KeyPressEvent
}

func NewKeyboardController(ctx context.Context, devicePath string) *KeyboardController {
	c := &KeyboardController{
		eventChan:   make(chan KeyPressEvent, 100),
		pressedKeys: make(map[JSKeyCode]bool, 6),
	}

	device, err := os.OpenFile(devicePath, os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal().Err(err).Str("path", devicePath).Msg("failed to open device")
	}

	c.device = device
	go c.usbActionDispatcher(ctx)
	return c
}

func (m *KeyboardController) usbActionDispatcher(ctx context.Context) {
	defer m.device.Close()
	pressedKeysArr := make([]JSKeyCode, 0, 6)
	prevPressedKeysArr := make([]JSKeyCode, 0, 6)
	for {
		select {
		case <-ctx.Done():
			return
		case keyPress := <-m.eventChan:
			if keyPress.IsDown {
				m.pressedKeys[keyPress.KeyCode] = true
			} else {
				delete(m.pressedKeys, keyPress.KeyCode)
			}

			clear(pressedKeysArr)
			for k := range m.pressedKeys {
				pressedKeysArr = append(pressedKeysArr, k)
			}

			slices.Sort(pressedKeysArr)
			if !slices.Equal(pressedKeysArr, prevPressedKeysArr) {
				clear(prevPressedKeysArr)
				for _, k := range pressedKeysArr {
					prevPressedKeysArr = append(prevPressedKeysArr, k)
				}

				fmt.Println("selected keys:", prevPressedKeysArr)
				if err := m.release(); err != nil {
					log.Error().Err(err).Msg("failed to release keys")
				}

				if err := m.sendReport(prevPressedKeysArr); err != nil {
					log.Error().Err(err).Msg("failed to press keys")
				}
			}
			fmt.Println("keyboard event received")
		}
	}
}

func (m *KeyboardController) EventChan() chan KeyPressEvent {
	return m.eventChan
}

func (m *KeyboardController) release() error {
	return m.sendReport([]JSKeyCode{})
}

/*
Report Structure for HID Keyboard

Byte 0: Modifier keys (bit flags)
Byte 1: Reserved (always 0x00)
Byte 2: Key code 1
Byte 3: Key code 2
Byte 4: Key code 3
Byte 5: Key code 4
Byte 6: Key code 5
Byte 7: Key code 6
*/
func (m *KeyboardController) sendReport(keys []JSKeyCode) error {
	report := make([]byte, 8)
	report[0] = byte(ModNone)
	report[1] = 0x00

	for i := 0; i < len(keys) && i < 6; i++ {
		key := keys[i].ToKey()
		if key.IsModifier() {
			report[0] = byte(Key(report[0]) | key)
			continue
		}

		report[2+i] = byte(key)
	}

	_, err := m.device.Write(report)
	return err
}
