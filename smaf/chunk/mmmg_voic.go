package chunk

import (
	"io"
	"strings"

	"github.com/mersenne-sister/smaf825/smaf/enums"
	"github.com/mersenne-sister/smaf825/smaf/util"
)

type MMMGVoiceChunk struct {
	*ChunkHeader
	SubChunks []Chunk
}

func (c *MMMGVoiceChunk) Traverse(fn func(Chunk)) {
	fn(c)
	for _, sub := range c.SubChunks {
		sub.Traverse(fn)
	}
}

func (c *MMMGVoiceChunk) String() string {
	result := c.ChunkHeader.String()
	sub := []string{}
	for _, chunk := range c.SubChunks {
		sub = append(sub, chunk.String())
	}
	return result + "\n" + util.Indent(strings.Join(sub, "\n"), "\t")
}

func (c *MMMGVoiceChunk) Read(rdr io.Reader) error {
	rest := int(c.ChunkHeader.Size)
	for 8 <= rest {
		var hdr ChunkHeader
		err := hdr.Read(rdr, &rest)
		if err != nil {
			return err
		}
		sub, err := hdr.CreateChunk(rdr, enums.ScoreTrackFormatType_Default)
		if err != nil {
			return err
		}
		c.SubChunks = append(c.SubChunks, sub)
	}
	return nil
}
