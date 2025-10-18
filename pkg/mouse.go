package pkg

type MouseButton uint8

const (
	ButtonNone   MouseButton = 0x00
	ButtonLeft   MouseButton = 0x01
	ButtonRight  MouseButton = 0x02
	ButtonMiddle MouseButton = 0x04
)

type JSMouseButton uint8

func (b JSMouseButton) ToMouseButton() MouseButton {
	switch b {
	case 0:
		return ButtonLeft
	case 1:
		return ButtonMiddle
	case 2:
		return ButtonRight
	default:
		panic("unknown mouse button")
	}
}
