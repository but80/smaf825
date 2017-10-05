package chunk

import (
	"io"
	"strings"

	"encoding/binary"
	"fmt"

	"github.com/mersenne-sister/smaf825/smaf/subtypes"
	"github.com/mersenne-sister/smaf825/smaf/util"
	"github.com/pkg/errors"
)

type ScoreTrackSetupDataChunk struct {
	*ChunkHeader
	Exclusives    []*subtypes.Exclusive
	UnknownStream []uint8
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

func (c *ScoreTrackSetupDataChunk) Read(rdr io.Reader) error {
	rest := int(c.Size)
	for 2 <= rest {
		var sig uint8
		err := binary.Read(rdr, binary.BigEndian, &sig)
		if err != nil {
			return errors.WithStack(err)
		}
		rest--
		if sig == 0xff {
			err = binary.Read(rdr, binary.BigEndian, &sig)
			if err != nil {
				return errors.WithStack(err)
			}
			rest--
		}
		switch sig {
		case 0xF0:
			ex := subtypes.NewExclusive(false)
			ex.Read(rdr, &rest)
			c.Exclusives = append(c.Exclusives, ex)
		default:
			c.UnknownStream = make([]uint8, rest)
			n, err := rdr.Read(c.UnknownStream)
			if err != nil {
				return errors.WithStack(err)
			}
			if n < len(c.UnknownStream) {
				return errors.Errorf("Cannot read enough byte length specified in chunk header")
			}
			rest -= len(c.UnknownStream)
			c.UnknownStream = append([]uint8{sig}, c.UnknownStream...)
		}
	}
	return nil
}
