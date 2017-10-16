package chunk

import (
	"fmt"
	"io"
	"strings"

	"encoding/binary"

	"github.com/but80/smaf825/smaf/enums"
	"github.com/but80/smaf825/smaf/event"
	"github.com/but80/smaf825/smaf/huffman"
	"github.com/but80/smaf825/smaf/util"
	"github.com/pkg/errors"
)

type ScoreTrackSequenceDataChunk struct {
	*ChunkHeader      `json:"chunk_header"`
	FormatType        enums.ScoreTrackFormatType           `json:"format_type"`
	Events            []event.DurationEventPair            `json:"events"`
	IsChannelUsed     map[enums.Channel]bool               `json:"-"`
	UsedChannelCount  int                                  `json:"-"`
	UsedNoteCount     map[enums.Channel]int                `json:"-"`
	NoteToChannel     map[enums.Channel]map[enums.Note]int `json:"-"`
	ChannelToChannels map[enums.Channel][]int              `json:"-"`
	IgnoredPC         map[uint32]bool                      `json:"-"`
}

func (c *ScoreTrackSequenceDataChunk) Traverse(fn func(Chunk)) {
	fn(c)
}

func (c *ScoreTrackSequenceDataChunk) String() string {
	result := "SequenceDataChunk: " + c.ChunkHeader.String()
	sub := []string{}
	for _, pair := range c.Events {
		sub = append(sub, pair.Event.String())
		if 0 < pair.Duration {
			sub = append(sub, fmt.Sprintf("      ..%d steps..", pair.Duration))
		}
	}
	return result + "\n" + util.Indent(strings.Join(sub, "\n"), "\t")
}

func (c *ScoreTrackSequenceDataChunk) Read(rdr io.Reader) error {
	var err error
	rest := int(c.Size)
	if c.FormatType == enums.ScoreTrackFormatType_MobileStandardCompressed {
		hrdr := huffman.NewHuffmanReader(rdr)
		rdr = hrdr
		rest, err = hrdr.Rest()
		if err != nil {
			return errors.WithStack(err)
		}
	}
	c.Events = []event.DurationEventPair{}
	ctx := event.NewSequenceBuilderContext()
	for 1 <= rest {
		if 4 == rest {
			var eos uint32
			err := binary.Read(rdr, binary.BigEndian, &eos)
			if err != nil {
				return errors.WithStack(err)
			}
			if eos == 0 {
				break
			}
			return errors.Errorf("Invalid event: 0x%08X at last", eos)
		}
		var pair event.DurationEventPair
		switch c.FormatType {
		case enums.ScoreTrackFormatType_HandyPhoneStandard:
			pair.Duration, err = util.ReadVariableInt(false, rdr, &rest)
			if err == nil {
				pair.Event, err = event.CreateEventHPS(rdr, &rest, ctx)
			}
		case enums.ScoreTrackFormatType_SEQU:
			pair.Duration, err = util.ReadVariableInt(false, rdr, &rest)
			if err == nil {
				pair.Event, err = event.CreateEventSEQU(rdr, &rest, ctx)
			}
		case enums.ScoreTrackFormatType_MobileStandardNonCompressed, enums.ScoreTrackFormatType_MobileStandardCompressed:
			pair.Duration, err = util.ReadVariableInt(true, rdr, &rest)
			if err == nil {
				pair.Event, err = event.CreateEvent(rdr, &rest, ctx)
			}
		default:
			return fmt.Errorf("Unsupported FormatType")
		}
		if err != nil {
			return errors.Wrapf(err, "at 0x%X in Mtsq", int(c.Size)-rest)
		}
		if pair.Event == nil {
			break
		}
		c.Events = append(c.Events, pair)
	}
	return nil
}

func (c *ScoreTrackSequenceDataChunk) AggregateUsage(channelsToSplit []enums.Channel) {
	c.IsChannelUsed = map[enums.Channel]bool{}
	usedNotes := map[enums.Channel]map[enums.Note]bool{}
	pc := map[enums.Channel]uint32{}
	for _, e := range c.Events {
		ch := e.Event.GetChannel()
		switch evt := e.Event.(type) {
		case *event.ControlChangeEvent:
			switch evt.CC {
			case enums.CC_BankSelectMSB:
				pc[ch] = pc[ch]&0x00FFFFFF | uint32(evt.Value)<<24
			case enums.CC_BankSelectLSB:
				pc[ch] = pc[ch]&0xFF00FFFF | uint32(evt.Value)<<16
			}
		case *event.ProgramChangeEvent:
			// @todo multiple PC in one channel
			pc[ch] = pc[ch]&0xFFFF00FF | uint32(evt.PC)<<8
		case *event.NoteEvent:
			c.IsChannelUsed[ch] = true
			if usedNotes[ch] == nil {
				usedNotes[ch] = map[enums.Note]bool{}
			}
			usedNotes[ch][evt.Note] = true
		}
	}
	// @todo Check available channel count
	unusedChannels := []enums.Channel{}
	for ch := enums.Channel(0); ch < 16; ch++ {
		if !c.IsChannelUsed[ch] {
			unusedChannels = append(unusedChannels, ch)
		}
	}
	//
	c.UsedChannelCount = 0
	for range c.IsChannelUsed {
		c.UsedChannelCount++
	}
	c.UsedNoteCount = map[enums.Channel]int{}
	for ch, n := range usedNotes {
		for range n {
			c.UsedNoteCount[ch]++
		}
	}
	//
	c.NoteToChannel = map[enums.Channel]map[enums.Note]int{}
	c.ChannelToChannels = map[enums.Channel][]int{}
	c.IgnoredPC = map[uint32]bool{}
	for _, ch := range channelsToSplit {
		if !c.IsChannelUsed[ch] {
			continue
		}
		c.NoteToChannel[ch] = map[enums.Note]int{}
		c.ChannelToChannels[ch] = []int{int(ch)}
		for note := range usedNotes[ch] {
			c.NoteToChannel[ch][note] = int(ch)
		}
		first := true
		for note := range usedNotes[ch] {
			if first {
				first = false
				continue
			}
			if len(unusedChannels) == 0 {
				fmt.Printf("Too many drum notes (%d in Ch.%d). %s is ignored\n", len(usedNotes[ch]), ch, note)
				c.NoteToChannel[ch][note] = -1
				c.IgnoredPC[pc[ch]|uint32(note)] = true
			} else {
				chTo := int(unusedChannels[0])
				unusedChannels = unusedChannels[1:]
				c.NoteToChannel[ch][note] = chTo
				c.ChannelToChannels[ch] = append(c.ChannelToChannels[ch], chTo)
			}
		}
	}
	//fmt.Printf("%+v\n", c.IgnoredPC)
	//fmt.Printf("%+v\n", c.IsChannelUsed)
	//fmt.Printf("%+v\n", c.NoteToChannel)
	//fmt.Printf("%+v\n", c.ChannelToChannels)
}

func (c *ScoreTrackSequenceDataChunk) IsIgnoredPC(bankMSB, bankLSB, PC int, drumNote enums.Note) bool {
	return c.IgnoredPC[uint32(bankMSB)<<24|uint32(bankLSB)<<16|uint32(PC)<<8|uint32(drumNote)]
}

func (c *ScoreTrackSequenceDataChunk) ChannelTo(orgCh enums.Channel, note enums.Note) int {
	if c.NoteToChannel[orgCh] == nil {
		return int(orgCh)
	}
	return c.NoteToChannel[orgCh][note]
}

func (c *ScoreTrackSequenceDataChunk) ChannelsTo(orgCh enums.Channel) []int {
	if c.ChannelToChannels[orgCh] == nil {
		return []int{int(orgCh)}
	}
	return c.ChannelToChannels[orgCh]
}
