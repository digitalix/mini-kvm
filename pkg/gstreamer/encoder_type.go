package gstreamer

type EncoderType uint8

const (
	EncoderTypeUNKOWN EncoderType = iota
	EncoderTypeHEVC_MPP
	EncoderTypeOPUS
)

func (t EncoderType) String() string {
	switch t {
	case EncoderTypeHEVC_MPP:
		return "HEVC_MPP"
	case EncoderTypeOPUS:
		return "OPUS"
	default:
		panic("unknown encoder type")
	}
}
