package voice

import "fmt"

type VoiceLib interface {
	fmt.Stringer
}

type chunkHeader struct {
	Signature uint32
	Size      uint32
}
