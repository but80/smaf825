package chunk

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"unsafe"

	"github.com/but80/go-smaf/v2/enums"
	"github.com/but80/go-smaf/v2/log"
	"github.com/but80/go-smaf/v2/subtypes"
	"github.com/pkg/errors"
)

type Signature uint32

func (s Signature) String() string {
	return fmt.Sprintf("%c%c%c%c=0x%08X", s>>24, s>>16&255, s>>8&255, s&255, uint32(s))
}

func (s Signature) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%c%c%c%c", s>>24, s>>16&255, s>>8&255, s&255))
}

type ExclusiveContainer interface {
	GetExclusives() []*subtypes.Exclusive
}

type Chunk interface {
	fmt.Stringer
	Read(io.Reader) error
	Traverse(func(Chunk))
}

type ChunkHeader struct {
	Signature Signature `json:"signature"`
	Size      uint32    `json:"-"`
}

func (hdr *ChunkHeader) String() string {
	if hdr == nil {
		return "<nil *ChunkHeader>"
	}
	s := hdr.Signature
	ss := fmt.Sprintf("%c%c%c", s>>24, s>>16&255, s>>8&255)
	switch ss {
	case "MTR", "ATR", "GTR", "Dch":
		ss += fmt.Sprintf("*(0x%02X)", uint32(s)&255)
	default:
		ss += fmt.Sprintf("%c", s&255)
	}
	return fmt.Sprintf("%s, %d bytes", ss, hdr.Size)
}

// Read は、バイト列を読み取ってパースした結果をこの構造体に格納します。
func (hdr *ChunkHeader) Read(rdr io.Reader, rest *int) error {
	err := binary.Read(rdr, binary.BigEndian, hdr)
	if err != nil {
		return errors.WithStack(err)
	}
	*rest -= int(unsafe.Sizeof(hdr)) + int(hdr.Size)
	return nil
}

func (hdr *ChunkHeader) CreateChunk(rdr io.Reader, formatType enums.ScoreTrackFormatType) (Chunk, error) {
	var chunk Chunk
	switch hdr.Signature {
	case 'C'<<24 | 'N'<<16 | 'T'<<8 | 'I': // CNTI
		log.Debugf("Creating ContentsInfoChunk")
		chunk = &ContentsInfoChunk{ChunkHeader: hdr}
	case 'O'<<24 | 'P'<<16 | 'D'<<8 | 'A': // OPDA
		log.Debugf("Creating OptionalDataChunk")
		chunk = &OptionalDataChunk{ChunkHeader: hdr}
	case 'M'<<24 | 'M'<<16 | 'M'<<8 | 'G': // MMMG
		log.Debugf("Creating MMMGChunk")
		chunk = &MMMGChunk{ChunkHeader: hdr}
	case 'M'<<24 | 's'<<16 | 'p'<<8 | 'I': // MspI
		log.Debugf("Creating SeekPhraseInfoChunk")
		chunk = &SeekPhraseInfoChunk{ChunkHeader: hdr}
	case 'M'<<24 | 't'<<16 | 's'<<8 | 'u': // Mtsu
		log.Debugf("Creating ScoreTrackSetupDataChunk")
		chunk = &ScoreTrackSetupDataChunk{ChunkHeader: hdr}
	case 'M'<<24 | 't'<<16 | 's'<<8 | 'q': // Mtsq
		log.Debugf("Creating ScoreTrackSequenceDataChunk")
		chunk = &ScoreTrackSequenceDataChunk{
			ChunkHeader: hdr,
			FormatType:  formatType,
		}
	case 'S'<<24 | 'E'<<16 | 'Q'<<8 | 'U': // SEQU
		log.Debugf("Creating ScoreTrackSequenceDataChunk")
		chunk = &ScoreTrackSequenceDataChunk{
			ChunkHeader: hdr,
			FormatType:  enums.ScoreTrackFormatTypeSEQU,
		}
	case 'V'<<24 | 'O'<<16 | 'I'<<8 | 'C': // VOIC
		log.Debugf("Creating MMMGVoiceChunk")
		chunk = &MMMGVoiceChunk{ChunkHeader: hdr}
	case 'E'<<24 | 'X'<<16 | 'V'<<8 | 'O': // EXVO
		log.Debugf("Creating MMMGEXVOChunk")
		chunk = &MMMGEXVOChunk{ChunkHeader: hdr}
	default:
		switch hdr.Signature & 0xFFFFFF00 {
		case 'M'<<24 | 'T'<<16 | 'R'<<8: // MTR*
			log.Debugf("Creating ScoreTrackChunk")
			chunk = &ScoreTrackChunk{ChunkHeader: hdr}
		case 'D'<<24 | 'c'<<16 | 'h'<<8: // Dch*
			log.Debugf("Creating DataChunk")
			chunk = &DataChunk{ChunkHeader: hdr}
		default: // unknown sub chunk
			log.Debugf("Creating UnknownChunk")
			chunk = &UnknownChunk{ChunkHeader: hdr}
		}
	}
	log.Enter()
	defer log.Leave()
	err := chunk.Read(rdr)
	if err != nil {
		return nil, errors.Wrapf(err, "Creating chunk %s", hdr.Signature.String())
	}
	return chunk, nil
}
