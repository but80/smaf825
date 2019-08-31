package event

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/but80/go-smaf/v2/enums"
	"github.com/but80/go-smaf/v2/internal/util"
	"github.com/but80/go-smaf/v2/subtypes"
	"github.com/pkg/errors"
)

type SequenceBuilderContext struct {
	lastVelocity [16]int
}

// NewSequenceBuilderContext は、新しい SequenceBuilderContext を作成します。
func NewSequenceBuilderContext() *SequenceBuilderContext {
	ctx := &SequenceBuilderContext{lastVelocity: [16]int{}}
	ctx.reset()
	return ctx
}

func (ctx *SequenceBuilderContext) reset() {
	for i := 0; i < 16; i++ {
		ctx.lastVelocity[i] = 64
	}
}

type Event interface {
	fmt.Stringer
	GetChannel() enums.Channel
	ShiftChannel(n int)
}

type NoteEvent struct {
	Channel  enums.Channel `json:"channel"`
	Note     enums.Note    `json:"note"`
	Velocity int           `json:"velocity"`
	GateTime int           `json:"gate_time"`
}

func (e *NoteEvent) GetChannel() enums.Channel {
	return e.Channel
}
func (e *NoteEvent) ShiftChannel(n int) {
	e.Channel += enums.Channel(n)
}

func (e *NoteEvent) String() string {
	return fmt.Sprintf("Tr.%02d Note %s Vel=%d", e.Channel, e.Note.String(), e.Velocity)
}

type ControlChangeEvent struct {
	Channel enums.Channel `json:"channel"`
	CC      enums.CC      `json:"cc"`
	Value   int           `json:"value"`
}

func (e *ControlChangeEvent) GetChannel() enums.Channel {
	return e.Channel
}
func (e *ControlChangeEvent) ShiftChannel(n int) {
	e.Channel += enums.Channel(n)
}

func (e *ControlChangeEvent) String() string {
	return fmt.Sprintf("Tr.%02d CC %s Value=%d", e.Channel, e.CC.String(), e.Value)
}

type ProgramChangeEvent struct {
	Channel enums.Channel `json:"channel"`
	PC      int           `json:"pc"`
}

func (e *ProgramChangeEvent) GetChannel() enums.Channel {
	return e.Channel
}
func (e *ProgramChangeEvent) ShiftChannel(n int) {
	e.Channel += enums.Channel(n)
}

func (e *ProgramChangeEvent) String() string {
	return fmt.Sprintf("Tr.%02d PC @%d", e.Channel, e.PC)
}

type PitchBendEvent struct {
	Channel enums.Channel `json:"channel"`
	Value   int           `json:"value"`
}

func (e *PitchBendEvent) GetChannel() enums.Channel {
	return e.Channel
}
func (e *PitchBendEvent) ShiftChannel(n int) {
	e.Channel += enums.Channel(n)
}

func (e *PitchBendEvent) String() string {
	return fmt.Sprintf("Tr.%02d PitchBend %d", e.Channel, e.Value)
}

type OctaveShiftEvent struct {
	Channel enums.Channel `json:"channel"`
	Value   int           `json:"value"`
}

func (e *OctaveShiftEvent) GetChannel() enums.Channel {
	return e.Channel
}
func (e *OctaveShiftEvent) ShiftChannel(n int) {
	e.Channel += enums.Channel(n)
}

func (e *OctaveShiftEvent) String() string {
	return fmt.Sprintf("Tr.%02d OctaveShift %d", e.Channel, e.Value)
}

type FineTuneEvent struct {
	Channel enums.Channel `json:"channel"`
	Value   int           `json:"value"`
}

func (e *FineTuneEvent) GetChannel() enums.Channel {
	return e.Channel
}
func (e *FineTuneEvent) ShiftChannel(n int) {
	e.Channel += enums.Channel(n)
}

func (e *FineTuneEvent) String() string {
	return fmt.Sprintf("Tr.%02d Fine %d", e.Channel, e.Value)
}

type ExclusiveEvent struct {
	Exclusive *subtypes.Exclusive `json:"exclusive"`
}

func (e *ExclusiveEvent) GetChannel() enums.Channel {
	return 0
}
func (e *ExclusiveEvent) ShiftChannel(n int) {
}

func (e *ExclusiveEvent) String() string {
	return fmt.Sprintf("Tr.-- %s", e.Exclusive.String())
}

type NopEvent struct {
}

func (e *NopEvent) GetChannel() enums.Channel {
	return 0
}
func (e *NopEvent) ShiftChannel(n int) {
}

func (e *NopEvent) String() string {
	return "Tr.-- NOP"
}

var shortModTable = []int{
	0x00, 0x00, 0x08, 0x10, 0x18, 0x20, 0x28, 0x30,
	0x38, 0x40, 0x48, 0x50, 0x60, 0x70, 0x7F, 0x7F,
}

var shortExpTable = []int{
	0x00, 0x00, 0x1F, 0x27, 0x2F, 0x37, 0x3F, 0x47,
	0x4F, 0x57, 0x5F, 0x67, 0x6F, 0x77, 0x7F, 0x7F,
}

func CreateEventSEQU(rdr io.Reader, rest *int, ctx *SequenceBuilderContext) (Event, error) {
	var sig uint8
	err := binary.Read(rdr, binary.BigEndian, &sig)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	*rest--

	if sig == 0x00 {
		err := binary.Read(rdr, binary.BigEndian, &sig)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		*rest--

		ch := enums.Channel(sig >> 6)
		msg := int(sig & 0x3f)
		if msg == 0x00 {
			var fine uint8
			err := binary.Read(rdr, binary.BigEndian, &fine)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			*rest--
			return &FineTuneEvent{Channel: ch, Value: int(fine)}, nil
		} else if 0x01 <= msg && msg <= 0x0e {
			return &ControlChangeEvent{Channel: ch, CC: enums.CCExpression, Value: shortExpTable[msg]}, nil
		} else if 0x11 <= msg && msg <= 0x1e {
			return &PitchBendEvent{Channel: ch, Value: int(msg-0x10) * 16384 / 16}, nil
		} else if 0x21 <= msg && msg <= 0x2e {
			return &ControlChangeEvent{Channel: ch, CC: enums.CCModulation, Value: shortModTable[msg-0x20]}, nil
		} else if msg == 0x30 {
			var value uint8
			err := binary.Read(rdr, binary.BigEndian, &value)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			*rest--
			return &ProgramChangeEvent{Channel: ch, PC: int(value)}, nil
		} else if msg == 0x31 {
			var value uint8
			err := binary.Read(rdr, binary.BigEndian, &value)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			*rest--
			return &ControlChangeEvent{Channel: ch, CC: enums.CCBankSelectLSB, Value: int(value)}, nil
		} else if msg == 0x32 {
			var value uint8
			err := binary.Read(rdr, binary.BigEndian, &value)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			*rest--
			v := int(value)
			if 0x80 <= v {
				v = 0x80 - v
			}
			return &OctaveShiftEvent{Channel: ch, Value: v}, nil
		} else if msg == 0x33 {
			var value uint8
			err := binary.Read(rdr, binary.BigEndian, &value)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			*rest--
			return &ControlChangeEvent{Channel: ch, CC: enums.CCModulation, Value: int(value)}, nil
		} else if msg == 0x34 {
			var value uint8
			err := binary.Read(rdr, binary.BigEndian, &value)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			*rest--
			return &PitchBendEvent{Channel: ch, Value: int(value) * 16384 / 256}, nil
		} else if msg == 0x36 {
			var value uint8
			err := binary.Read(rdr, binary.BigEndian, &value)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			*rest--
			return &ControlChangeEvent{Channel: ch, CC: enums.CCExpression, Value: int(value)}, nil
		} else if msg == 0x37 {
			var value uint8
			err := binary.Read(rdr, binary.BigEndian, &value)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			*rest--
			return &ControlChangeEvent{Channel: ch, CC: enums.CCMainVolume, Value: int(value)}, nil
		} else if msg == 0x3a {
			var value uint8
			err := binary.Read(rdr, binary.BigEndian, &value)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			*rest--
			return &ControlChangeEvent{Channel: ch, CC: enums.CCPanpot, Value: int(value)}, nil
		} else if msg == 0x3b {
			var value uint8
			err := binary.Read(rdr, binary.BigEndian, &value)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			*rest--
			return &ControlChangeEvent{Channel: ch, CC: enums.CCExpression, Value: int(value)}, nil
		} else {
			return nil, errors.Errorf("Invalid event: 0x%02X", sig)
		}
	} else if sig == 0xff {
		var sig2 uint8
		err := binary.Read(rdr, binary.BigEndian, &sig2)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		*rest--
		switch sig2 {
		case 0x00:
			return &NopEvent{}, nil
		case 0xF0:
			ex := subtypes.NewExclusive(false)
			err = ex.Read(rdr, rest)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			return &ExclusiveEvent{Exclusive: ex}, nil
		default:
			return nil, errors.Errorf("Invalid event: 0x%02X%02X", sig, sig2)
		}
	} else {
		ch := enums.Channel(sig >> 6)
		note := enums.Note(sig&15) + (enums.Note(sig>>4&3)+3)*12
		gate, err := util.ReadVariableInt(false, rdr, rest)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return &NoteEvent{Channel: ch, Note: note, Velocity: 127, GateTime: gate}, nil
	}
}

func CreateEventHPS(rdr io.Reader, rest *int, ctx *SequenceBuilderContext) (Event, error) {
	var sig uint8
	err := binary.Read(rdr, binary.BigEndian, &sig)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	*rest--

	if sig == 0xff {
		err := binary.Read(rdr, binary.BigEndian, &sig)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		*rest--

		if sig == 0x00 {
			return &NopEvent{}, nil
		}
		if sig == 0xf0 {
			ex := subtypes.NewExclusive(true)
			err = ex.Read(rdr, rest)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			return &ExclusiveEvent{Exclusive: ex}, nil
		}

		return nil, errors.Errorf("Invalid event: 0xFF%02X", sig)
	}

	if sig != 0 {
		oct := int(sig >> 4 & 3)
		notenum := int(sig & 15)
		gatetime, err := util.ReadVariableInt(false, rdr, rest)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return &NoteEvent{
			Channel:  enums.Channel(sig >> 6),
			Note:     enums.Note((oct+3)*12 + notenum),
			Velocity: 127,
			GateTime: gatetime,
		}, nil
	}

	err = binary.Read(rdr, binary.BigEndian, &sig)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	*rest--

	ch := enums.Channel(sig >> 6)
	switch sig >> 4 & 3 {

	case 3:
		var val uint8
		err = binary.Read(rdr, binary.BigEndian, &val)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		*rest--
		value := int(val)

		switch sig & 15 {
		case 0: // program change
			return &ProgramChangeEvent{Channel: ch, PC: value}, nil
		case 1: // bank select
			return &ControlChangeEvent{Channel: ch, CC: enums.CCBankSelectLSB, Value: value}, nil
		case 2: // octave shift
			if 0x80 <= value {
				value = -(value - 0x80)
			}
			return &OctaveShiftEvent{Channel: ch, Value: value}, nil
		case 3: // modulation
			return &ControlChangeEvent{Channel: ch, CC: enums.CCModulation, Value: value}, nil
		case 4: // pitch bend
			return &PitchBendEvent{Channel: ch, Value: (value - 64) * (8192 / 64)}, nil
		case 7: // volume Gain[dB] = 20*log(Data^2/127^2)
			return &ControlChangeEvent{Channel: ch, CC: enums.CCMainVolume, Value: value}, nil
		case 10: // pan
			return &ControlChangeEvent{Channel: ch, CC: enums.CCPanpot, Value: value}, nil
		case 11: // expression
			return &ControlChangeEvent{Channel: ch, CC: enums.CCExpression, Value: value}, nil
		default:
			return nil, errors.Errorf("Invalid event: 0x%02X", sig)
		}

	case 2:
		// modulation
		value := int(sig & 15)
		return &ControlChangeEvent{Channel: ch, CC: enums.CCModulation, Value: shortModTable[value]}, nil

	case 1:
		// pitch bend
		value := int(sig & 15)
		return &PitchBendEvent{Channel: ch, Value: (value*8 - 64) * (8192 / 64)}, nil

	case 0:
		// expression
		value := int(sig & 15)
		return &ControlChangeEvent{Channel: ch, CC: enums.CCExpression, Value: shortExpTable[value]}, nil
	}

	return nil, errors.Errorf("Invalid event: 0x00%02X", sig)
}

func CreateEvent(rdr io.Reader, rest *int, ctx *SequenceBuilderContext) (Event, error) {
	var sig uint8
	err := binary.Read(rdr, binary.BigEndian, &sig)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	*rest--

	ch := enums.Channel(sig & 0x0F)
	switch sig & 0xF0 {

	case 0x80, 0x90:
		var noteNum uint8
		err = binary.Read(rdr, binary.BigEndian, &noteNum)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		*rest--
		var vel int
		if sig&0xF0 == 0x90 {
			var v uint8
			err = binary.Read(rdr, binary.BigEndian, &v)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			vel = int(v)
			ctx.lastVelocity[ch] = vel
		} else {
			vel = ctx.lastVelocity[ch]
		}
		dur, err := util.ReadVariableInt(true, rdr, rest)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return &NoteEvent{Channel: ch, Note: enums.Note(noteNum), Velocity: vel, GateTime: dur}, nil

	case 0xB0:
		var cc uint8
		err = binary.Read(rdr, binary.BigEndian, &cc)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		*rest--
		var value uint8
		err = binary.Read(rdr, binary.BigEndian, &value)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		*rest--
		return &ControlChangeEvent{Channel: ch, CC: enums.CC(cc), Value: int(value)}, nil

	case 0xC0:
		var pc uint8
		err = binary.Read(rdr, binary.BigEndian, &pc)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		*rest--
		return &ProgramChangeEvent{Channel: ch, PC: int(pc)}, nil

	case 0xE0:
		var v uint16
		err = binary.Read(rdr, binary.LittleEndian, &v)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		*rest -= 2
		return &PitchBendEvent{Channel: ch, Value: int(v&0x7F|v&0x7F00>>1) - 8192}, nil

	case 0xF0:
		switch sig {
		case 0xF0:
			ex := subtypes.NewExclusive(true)
			err = ex.Read(rdr, rest)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			return &ExclusiveEvent{Exclusive: ex}, nil
		case 0xFF:
			var s uint8
			err = binary.Read(rdr, binary.BigEndian, &s)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			switch s {
			case 0x00:
				return &NopEvent{}, nil
			case 0x2F:
				err = binary.Read(rdr, binary.BigEndian, &s)
				if err != nil {
					return nil, errors.WithStack(err)
				}
				switch s {
				case 0x00:
					return nil, nil
				}
			}
		}
	}

	return nil, errors.Errorf("Invalid event: 0x%02X", sig)
}
