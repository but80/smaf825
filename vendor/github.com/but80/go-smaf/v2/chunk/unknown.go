package chunk

import (
	"io"

	"github.com/pkg/errors"
)

type UnknownChunk struct {
	*ChunkHeader
	Stream []uint8 `json:"stream"`
}

func (c *UnknownChunk) Traverse(fn func(Chunk)) {
	fn(c)
}

func (c *UnknownChunk) String() string {
	result := "UnknownChunk: " + c.ChunkHeader.String()
	return result
}

// Read は、バイト列を読み取ってパースした結果をこの構造体に格納します。
func (c *UnknownChunk) Read(rdr io.Reader) error {
	c.Stream = make([]uint8, c.ChunkHeader.Size)
	n, err := rdr.Read(c.Stream)
	if err != nil {
		return err
	}
	if n < len(c.Stream) {
		return errors.Errorf("Cannot read enough byte length specified in chunk header (%d < %d)", n, len(c.Stream))
	}
	return nil
}
