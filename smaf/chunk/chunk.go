package chunk

import (
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"

	"github.com/mersenne-sister/smaf825/smaf/enums"
	"github.com/pkg/errors"
)

type Chunk interface {
	fmt.Stringer
	Read(io.Reader) error
	Traverse(func(Chunk))
}

type ChunkHeader struct {
	Signature uint32
	Size      uint32
}

func (hdr *ChunkHeader) String() string {
	s := hdr.Signature
	ss := fmt.Sprintf("%c%c%c", s>>24, s>>16&255, s>>8&255)
	switch ss {
	case "MTR", "ATR", "GTR":
		ss += fmt.Sprintf("*(0x%02X)", s&255)
	default:
		ss += fmt.Sprintf("%c", s&255)
	}
	return fmt.Sprintf("%s, %d bytes", ss, hdr.Size)
}

func (hdr *ChunkHeader) Read(rdr io.Reader, rest *int) error {
	err := binary.Read(rdr, binary.BigEndian, hdr)
	if err != nil {
		return err
	}
	*rest -= int(unsafe.Sizeof(hdr)) + int(hdr.Size)
	return nil
}

func (hdr *ChunkHeader) CreateChunk(rdr io.Reader, formatType enums.ScoreTrackFormatType) (Chunk, error) {
	var chunk Chunk
	switch hdr.Signature {
	case 'C'<<24 | 'N'<<16 | 'T'<<8 | 'I': // CNTI
		chunk = &ContentsInfoChunk{ChunkHeader: hdr}
	case 'O'<<24 | 'P'<<16 | 'D'<<8 | 'A': // OPDA
		chunk = &OptionalDataChunk{ChunkHeader: hdr}
	case 'M'<<24 | 'M'<<16 | 'M'<<8 | 'G': // MMMG
		chunk = &MMMGChunk{ChunkHeader: hdr}
	case 'M'<<24 | 's'<<16 | 'p'<<8 | 'I': // MspI
		chunk = &SeekPhraseInfoChunk{ChunkHeader: hdr}
	case 'M'<<24 | 't'<<16 | 's'<<8 | 'u': // Mtsu
		chunk = &ScoreTrackSetupDataChunk{ChunkHeader: hdr}
	case 'M'<<24 | 't'<<16 | 's'<<8 | 'q': // Mtsq
		chunk = &ScoreTrackSequenceDataChunk{
			ChunkHeader: hdr,
			FormatType:  formatType,
		}
	case 'V'<<24 | 'O'<<16 | 'I'<<8 | 'C': // VOIC
		chunk = &MMMGVoiceChunk{ChunkHeader: hdr}
	case 'E'<<24 | 'X'<<16 | 'V'<<8 | 'O': // EXVO
		chunk = &MMMGEXVOChunk{ChunkHeader: hdr}
	default:
		switch hdr.Signature & 0xFFFFFF00 {
		case 'M'<<24 | 'T'<<16 | 'R'<<8: // MTRx
			chunk = &ScoreTrackChunk{ChunkHeader: hdr}
		case 'D'<<24 | 'c'<<16 | 'h'<<8: // Dchx
			chunk = &DataChunk{ChunkHeader: hdr}
		default: // unknown sub chunk
			chunk = &UnknownChunk{ChunkHeader: hdr}
		}
	}
	err := chunk.Read(rdr)
	if err != nil {
		return nil, errors.Wrapf(err, "Creating %T (0x%08X)", chunk, hdr.Signature)
	}
	return chunk, nil
}
