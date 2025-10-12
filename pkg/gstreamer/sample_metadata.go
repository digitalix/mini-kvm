package gstreamer

type SampleMetadata struct {
	IsKeyFrame bool
	Source     string
	MediaType  MediaType
}
