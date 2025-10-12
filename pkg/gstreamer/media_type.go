package gstreamer

type MediaType uint8

const (
	MediaTypeUnknown MediaType = iota
	MediaTypeVideo
	MediaTypeAudio
)

func (mt MediaType) String() string {
	switch mt {
	default:
		return "unknown"
	case MediaTypeVideo:
		return "video"
	case MediaTypeAudio:
		return "audio"
	}
}
