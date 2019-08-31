package chunk

import (
	"io"
	"strings"

	"github.com/but80/go-smaf/v2/enums"
	"github.com/but80/go-smaf/v2/internal/util"
	"github.com/pkg/errors"
)

type OptionalDataChunk struct {
	*ChunkHeader
	SubChunks []Chunk `json:"sub_chunks"`
}

func (c *OptionalDataChunk) Traverse(fn func(Chunk)) {
	fn(c)
	for _, sub := range c.SubChunks {
		sub.Traverse(fn)
	}
}

func (c *OptionalDataChunk) String() string {
	result := "OptionalDataChunk: " + c.ChunkHeader.String()
	sub := []string{}
	for _, chunk := range c.SubChunks {
		sub = append(sub, chunk.String())
	}
	return result + "\n" + util.Indent(strings.Join(sub, "\n"), "\t")
}

// Read は、バイト列を読み取ってパースした結果をこの構造体に格納します。
func (c *OptionalDataChunk) Read(rdr io.Reader) error {
	rest := int(c.ChunkHeader.Size)
	c.SubChunks = []Chunk{}
	for 8 <= rest {
		var hdr ChunkHeader
		err := hdr.Read(rdr, &rest)
		if err != nil {
			return errors.WithStack(err)
		}
		sub, err := hdr.CreateChunk(rdr, enums.ScoreTrackFormatTypeDefault)
		if err != nil {
			return errors.WithStack(err)
		}
		c.SubChunks = append(c.SubChunks, sub)
	}
	return nil
}
