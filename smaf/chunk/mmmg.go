package chunk

import (
	"io"
	"strings"

	"unsafe"

	"github.com/mersenne-sister/smaf825/smaf/enums"
	"github.com/mersenne-sister/smaf825/smaf/util"
)

type MMMGChunk struct {
	*ChunkHeader
	Enigma    uint16
	SubChunks []Chunk
}

func (c *MMMGChunk) Traverse(fn func(Chunk)) {
	fn(c)
	for _, sub := range c.SubChunks {
		sub.Traverse(fn)
	}
}

func (c *MMMGChunk) String() string {
	result := "MMMGChunk: " + c.ChunkHeader.String()
	sub := []string{}
	for _, chunk := range c.SubChunks {
		sub = append(sub, chunk.String())
	}
	return result + "\n" + util.Indent(strings.Join(sub, "\n"), "\t")
}

func (c *MMMGChunk) Read(rdr io.Reader) error {
	rest := int(c.ChunkHeader.Size)
	enigma := make([]byte, 2)
	rdr.Read(enigma)
	rest -= int(unsafe.Sizeof(enigma))
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
