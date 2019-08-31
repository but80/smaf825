package chunk

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/but80/go-smaf/v2/internal/util"
	"github.com/but80/go-smaf/v2/log"
	"github.com/but80/go-smaf/v2/subtypes"
	"github.com/pkg/errors"
)

type ScoreTrackSetupDataChunk struct {
	*ChunkHeader  `json:"chunk_header"`
	Exclusives    []*subtypes.Exclusive `json:"exclusives"`
	UnknownStream []uint8               `json:"unknown_stream,omitempty"`
}

func (c *ScoreTrackSetupDataChunk) GetExclusives() []*subtypes.Exclusive {
	return c.Exclusives
}

func (c *ScoreTrackSetupDataChunk) Traverse(fn func(Chunk)) {
	fn(c)
}

func (c *ScoreTrackSetupDataChunk) String() string {
	result := "SetupDataChunk: " + c.ChunkHeader.String()
	sub := []string{}
	for _, ex := range c.Exclusives {
		sub = append(sub, ex.String())
	}
	if 0 < len(c.UnknownStream) {
		sub = append(sub, fmt.Sprintf("UnknownStream: %s", util.Hex(c.UnknownStream)))
	}
	return result + "\n" + util.Indent(strings.Join(sub, "\n"), "\t")
}

// Read は、バイト列を読み取ってパースした結果をこの構造体に格納します。
func (c *ScoreTrackSetupDataChunk) Read(rdr io.Reader) error {
	rest := int(c.Size)
	for 1 <= rest {
		var sig uint8
		err := binary.Read(rdr, binary.BigEndian, &sig)
		if err != nil {
			log.Debugf("Read failed")
			return errors.WithStack(err)
		}
		rest--
		if sig == 0xff {
			log.Debugf("Read 0xff")
			err = binary.Read(rdr, binary.BigEndian, &sig)
			if err != nil {
				return errors.WithStack(err)
			}
			rest--
		}
		if rest == 0 {
			log.Debugf("Unexpected EOF")
			break
		}
		switch sig {
		case 0xF0:
			log.Debugf("Creating Exclusive")
			ex := subtypes.NewExclusive(false)
			log.Enter()
			err := ex.Read(rdr, &rest)
			log.Leave()
			if err != nil {
				return errors.WithStack(err)
			}
			c.Exclusives = append(c.Exclusives, ex)
		default:
			log.Debugf("Creating UnknownStream")
			c.UnknownStream = make([]uint8, rest)
			n, err := rdr.Read(c.UnknownStream)
			if err != nil {
				return errors.WithStack(err)
			}
			if n < len(c.UnknownStream) {
				return errors.Errorf("Cannot read enough byte length specified in chunk header (%d < %d)", n, len(c.UnknownStream))
			}
			rest -= len(c.UnknownStream)
			c.UnknownStream = append([]uint8{sig}, c.UnknownStream...)
		}
	}
	if rest != 0 {
		return errors.Errorf("Chunk size mismatch (%d != %d)", int(c.Size)-rest, c.Size)
	}
	return nil
}
