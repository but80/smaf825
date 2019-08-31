package chunk

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
	"unsafe"

	"github.com/but80/go-smaf/v2/enums"
	"github.com/but80/go-smaf/v2/internal/util"
	"github.com/but80/go-smaf/v2/subtypes"
	"github.com/but80/go-smaf/v2/voice"
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
	if c == nil {
		return "<nil *FileChunk>"
	}
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

// Read は、バイト列を読み取ってパースした結果をこの構造体に格納します。
func (c *FileChunk) Read(rdr io.Reader) error {
	c.SubChunks = []Chunk{}

	c.ChunkHeader = &ChunkHeader{}
	err := binary.Read(rdr, binary.BigEndian, c.ChunkHeader)
	if err != nil {
		return errors.WithStack(err)
	}

	rest := int(c.ChunkHeader.Size)

	for 10 <= rest {
		var hdr ChunkHeader
		from := int(c.ChunkHeader.Size) - rest + int(unsafe.Sizeof(c.ChunkHeader))
		err := hdr.Read(rdr, &rest)
		to := int(c.ChunkHeader.Size) - rest + int(unsafe.Sizeof(c.ChunkHeader))
		if err != nil {
			return errors.Wrapf(err, "at 0x%X -> 0x%X", from, to)
		}
		sub, err := hdr.CreateChunk(rdr, enums.ScoreTrackFormatTypeDefault)
		if err != nil {
			return errors.Wrapf(err, "at 0x%X -> 0x%X", from, to)
		}
		c.SubChunks = append(c.SubChunks, sub)
	}

	rest -= 2

	if rest != 0 {
		return fmt.Errorf("Size mismatch (want %d, got %d)", int(c.ChunkHeader.Size), int(c.ChunkHeader.Size)-rest)
	}

	return nil
}

type CollectedVoices struct {
	Voices []*voice.VM35FMVoice `json:"voices"`
}

func (v *CollectedVoices) String() string {
	result := []string{}
	for _, v := range v.Voices {
		result = append(result, v.String())
	}
	return strings.Join(result, "\n\n") + "\n"
}

type CollectedExclusives struct {
	Exclusives []*subtypes.Exclusive `json:"voices"`
}

func (v *CollectedExclusives) String() string {
	result := []string{}
	for _, v := range v.Exclusives {
		result = append(result, v.String())
	}
	return strings.Join(result, "\n\n") + "\n"
}

func (v *CollectedExclusives) Voices() *CollectedVoices {
	result := &CollectedVoices{Voices: []*voice.VM35FMVoice{}}
	for _, v := range v.Exclusives {
		if v.VM35VoicePC != nil {
			switch vv := v.VM35VoicePC.Voice.(type) {
			case *voice.VM35FMVoice:
				result.Voices = append(result.Voices, vv)
			}
		}
		if v.VMAVoicePC != nil {
			result.Voices = append(result.Voices, v.VMAVoicePC.Voice.ToVM35())
		}
	}
	return result
}

func (c *FileChunk) CollectExclusives() *CollectedExclusives {
	voices := &CollectedExclusives{}
	c.Traverse(func(c Chunk) {
		switch sc := c.(type) {
		case *ScoreTrackSetupDataChunk:
			voices.Exclusives = append(voices.Exclusives, sc.Exclusives...)
		case *MMMGEXVOChunk:
			voices.Exclusives = append(voices.Exclusives, sc.Exclusive)
		}
	})
	return voices
}

// NewFileChunk は、新しい FileChunk を作成します。
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
	c.CRCGot, err = calcCRC(fh, int(unsafe.Sizeof(c.ChunkHeader))+int(c.Size)-2)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = binary.Read(fh, binary.BigEndian, &c.CRCWant)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return c, nil
}
