package chunk

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"unsafe"

	"github.com/but80/go-smaf/v2/enums"
	"github.com/but80/go-smaf/v2/internal/util"
	"github.com/but80/go-smaf/v2/log"
	"github.com/but80/go-smaf/v2/subtypes"
	"github.com/pkg/errors"
)

type scoreTrackRawHeader struct {
	FormatType   uint8
	SequenceType uint8
	TimebaseD    uint8 // Sequence Data 内部で使用する Duration の基準時間を表現する。
	TimebaseG    uint8 // Sequence Data 内部で使用する GateTime の基準時間を表現する。
}

type ScoreTrackChunk struct {
	*ChunkHeader     `json:"chunk_header"`
	FormatType       enums.ScoreTrackFormatType                `json:"format_type"`
	SequenceType     enums.ScoreTrackSequenceType              `json:"sequence_type"`
	DurationTimeBase int                                       `json:"duration_time_base"`
	GateTimeBase     int                                       `json:"gate_time_base"`
	ChannelStatus    map[enums.Channel]*subtypes.ChannelStatus `json:"channel_status"`
	SubChunks        []Chunk                                   `json:"sub_chunks"`
}

func (c *ScoreTrackChunk) Traverse(fn func(Chunk)) {
	fn(c)
	for _, sub := range c.SubChunks {
		sub.Traverse(fn)
	}
}

func (c *ScoreTrackChunk) String() string {
	result := "ScoreTrackChunk: " + c.ChunkHeader.String()
	chst := []string{}
	for ch, st := range c.ChannelStatus {
		chst = append(chst, fmt.Sprintf("[%d] KeyControl=%s LED=%v Vibration=%v ChannelType=%s", ch, st.KeyControlStatus, st.LEDStatus, st.VibrationStatus, st.ChannelType))
	}
	sub := []string{
		fmt.Sprintf("FormatType: %s", c.FormatType.String()),
		fmt.Sprintf("SequenceType: %s", c.SequenceType.String()),
		fmt.Sprintf("DurationTimeBase: %d msec", c.DurationTimeBase),
		fmt.Sprintf("GateTimeBase: %d msec", c.GateTimeBase),
		"ChannelStatus:",
		util.Indent(strings.Join(chst, "\n"), "\t"),
	}
	for _, chunk := range c.SubChunks {
		sub = append(sub, chunk.String())
	}
	return result + "\n" + util.Indent(strings.Join(sub, "\n"), "\t")
}

func timeBase(b uint8) int {
	switch b {
	case 0x00:
		return 1
	case 0x01:
		return 2
	case 0x02:
		return 4
	case 0x03:
		return 5
	case 0x10:
		return 10
	case 0x11:
		return 20
	case 0x12:
		return 40
	case 0x13:
		return 50
	default:
		return 2
	}
}

// Read は、バイト列を読み取ってパースした結果をこの構造体に格納します。
func (c *ScoreTrackChunk) Read(rdr io.Reader) error {
	rest := int(c.ChunkHeader.Size)
	var rawHeader scoreTrackRawHeader
	err := binary.Read(rdr, binary.BigEndian, &rawHeader)
	if err != nil {
		return errors.WithStack(err)
	}
	rest -= int(unsafe.Sizeof(rawHeader))

	c.FormatType = enums.ScoreTrackFormatType(rawHeader.FormatType)
	c.SequenceType = enums.ScoreTrackSequenceType(rawHeader.SequenceType)
	c.DurationTimeBase = timeBase(rawHeader.TimebaseD)
	c.GateTimeBase = timeBase(rawHeader.TimebaseG)

	log.Debugf("FormatType %s", c.FormatType.String())
	if !c.FormatType.IsSupported() {
		return fmt.Errorf("Unsupported FormatType %d", int(c.FormatType))
	}

	c.ChannelStatus = map[enums.Channel]*subtypes.ChannelStatus{}
	switch c.FormatType {
	case enums.ScoreTrackFormatTypeHandyPhoneStandard:
		var b uint16
		err := binary.Read(rdr, binary.BigEndian, &b)
		rest -= 2
		if err != nil {
			return errors.WithStack(err)
		}
		for ch := enums.Channel(0); ch < 4; ch++ {
			c.ChannelStatus[ch] = &subtypes.ChannelStatus{
				KeyControlStatus: enums.KeyControlStatus((b >> 15 & 1) + 1),
				VibrationStatus:  b>>14&1 != 0,
				ChannelType:      enums.ChannelType(b >> 12 & 3),
			}
			b <<= 4
		}
	default:
		var b uint8
		for ch := enums.Channel(0); ch < 16; ch++ {
			err := binary.Read(rdr, binary.BigEndian, &b)
			rest--
			if err != nil {
				return errors.WithStack(err)
			}
			c.ChannelStatus[ch] = &subtypes.ChannelStatus{
				KeyControlStatus: enums.KeyControlStatus(b >> 6),
				VibrationStatus:  b>>5&1 != 0,
				LEDStatus:        b>>4&1 != 0,
				ChannelType:      enums.ChannelType(b & 3),
			}
		}
	}

	c.SubChunks = []Chunk{}
	for 8 <= rest {
		var hdr ChunkHeader
		err := hdr.Read(rdr, &rest)
		if err != nil {
			return errors.WithStack(err)
		}
		sub, err := hdr.CreateChunk(rdr, c.FormatType)
		if err != nil {
			return errors.WithStack(err)
		}
		c.SubChunks = append(c.SubChunks, sub)
	}
	return nil
}
