package chunk

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"

	"unsafe"

	"github.com/but80/smaf825/smaf/enums"
	"github.com/but80/smaf825/smaf/util"
	"github.com/pkg/errors"
)

type FileChunk struct {
	*ChunkHeader
	SubChunks []Chunk `json:"sub_chunks"`
	CRCGot    uint16  `json:"-"`
	CRCWant   uint16  `json:"-"`
}

func (c *FileChunk) Traverse(fn func(Chunk)) {
	fn(c)
	for _, sub := range c.SubChunks {
		sub.Traverse(fn)
	}
}

func (c *FileChunk) String() string {
	result := "MMF File Chunk: " + c.ChunkHeader.String()
	sub := []string{}
	for _, chunk := range c.SubChunks {
		sub = append(sub, chunk.String())
	}
	crc := fmt.Sprintf("CRC = want: 0x%04X, got: 0x%04X", c.CRCWant, c.CRCGot)
	if c.CRCWant == c.CRCGot {
		crc += " (valid)"
	} else {
		crc += " (invalid)"
	}
	sub = append(sub, crc)
	return result + "\n" + util.Indent(strings.Join(sub, "\n"), "\t")
}

func calcCRC(rdr io.Reader, len int) (uint16, error) {
	crctable := [0x100]uint16{}
	var r uint16
	for i := uint16(0); i < 0x100; i++ {
		r = i << 8
		for j := 0; j < 8; j++ {
			if (r >> 15) != 0 {
				r = (r << 1) ^ 0x1021
			} else {
				r <<= 1
			}
		}
		crctable[i] = r
	}

	var crc uint16 = 0xFFFF
	for i := 0; i < len; i++ {
		var b uint8
		err := binary.Read(rdr, binary.BigEndian, &b)
		if err != nil {
			return 0, errors.WithStack(err)
		}
		crc = (crc << 8) ^ crctable[uint8(crc>>8)^b]
	}
	return crc ^ 0xFFFF, nil
}

func (c *FileChunk) Read(rdr io.Reader) error {
	c.SubChunks = []Chunk{}

	c.ChunkHeader = &ChunkHeader{}
	err := binary.Read(rdr, binary.BigEndian, c.ChunkHeader)
	if err != nil {
		return errors.WithStack(err)
	}

	rest := int(c.ChunkHeader.Size)

	for 8 <= rest {
		var hdr ChunkHeader
		from := int(c.ChunkHeader.Size) - rest + int(unsafe.Sizeof(c.ChunkHeader))
		err := hdr.Read(rdr, &rest)
		to := int(c.ChunkHeader.Size) - rest + int(unsafe.Sizeof(c.ChunkHeader))
		if err != nil {
			return errors.Wrapf(err, "at 0x%X -> 0x%X", from, to)
		}
		sub, err := hdr.CreateChunk(rdr, enums.ScoreTrackFormatType_Default)
		if err != nil {
			return errors.Wrapf(err, "at 0x%X -> 0x%X", from, to)
		}
		c.SubChunks = append(c.SubChunks, sub)
	}

	return nil
}

func NewFileChunk(file string) (*FileChunk, error) {
	fh, err := os.Open(file)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer fh.Close()

	c := &FileChunk{}
	err = c.Read(fh)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	_, err = fh.Seek(0, 0)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	fi, err := fh.Stat()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	c.CRCGot, err = calcCRC(fh, int(fi.Size())-2)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = binary.Read(fh, binary.BigEndian, &c.CRCWant)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return c, nil
}
