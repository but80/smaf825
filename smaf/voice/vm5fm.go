package voice

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"unsafe"

	"bytes"

	"github.com/but80/smaf825/smaf/enums"
	"github.com/but80/smaf825/smaf/util"
	"github.com/pkg/errors"
)

type VM5FMOperator struct {
	Num   int              `json:"-"`     // Operator number
	MULTI enums.Multiplier `json:"multi"` // Multiplier
	DT    int              `json:"dt"`    // Detune
	AR    int              `json:"ar"`    // Attack Rate
	DR    int              `json:"dr"`    // Decay Rate
	SR    int              `json:"sr"`    // Sustain Rate
	RR    int              `json:"rr"`    // Release Rate
	SL    int              `json:"sl"`    // Sustain Level
	TL    int              `json:"tl"`    // Total Level
	KSL   int              `json:"ksl"`   // Key Scaling Level
	DAM   int              `json:"dam"`   // Depth of AM
	DVB   int              `json:"dvb"`   // Depth of Vibrato
	FB    int              `json:"fb"`    // Feedback
	WS    int              `json:"ws"`    // Wave Shape
	XOF   bool             `json:"xof"`   // Ignore KeyOff
	SUS   bool             `json:"sus"`   // Keep sustain rate after KeyOff (unused in YMF825)
	KSR   bool             `json:"ksr"`   // Key Scaling Rate
	EAM   bool             `json:"eam"`   // Enable AM
	EVB   bool             `json:"evb"`   // Enable Vibrato
}

func (op *VM5FMOperator) Read(rdr io.Reader, rest *int) error {
	//    | 7 | 6 | 5 | 4 | 3 | 2 | 1 | 0 |
	// +0 |      S R      |XOF| - |SUS|KSR|
	// +1 |      R R      |      D R      |
	// +2 |      A R      |      S L      |
	// +3 |          T L          |  KSL  |
	// +4 | - |  DAM  |EAM| - |  DVB  |EVB|
	// +5 |     MULTI     | - |    D T    |
	// +6 |        W S        |    F B    |

	data := [7]uint8{}
	err := binary.Read(rdr, binary.BigEndian, &data)
	if err != nil {
		return errors.WithStack(err)
	}
	*rest -= int(unsafe.Sizeof(data))

	op.SR = int(data[0] >> 4)
	op.XOF = data[0]&0x08 != 0
	op.SUS = data[0]&0x02 != 0
	op.KSR = data[0]&0x01 != 0
	op.RR = int(data[1] >> 4)
	op.DR = int(data[1] & 15)
	op.AR = int(data[2] >> 4)
	op.SL = int(data[2] & 15)
	op.TL = int(data[3] >> 2)
	op.KSL = int(data[3] & 3)
	op.DAM = int(data[4] >> 5 & 3)
	op.EAM = data[4]&0x10 != 0
	op.DVB = int(data[4] >> 1 & 3)
	op.EVB = data[4]&0x01 != 0
	op.MULTI = enums.Multiplier(data[5] >> 4)
	op.DT = int(data[5] & 7)
	op.WS = int(data[6] >> 3)
	op.FB = int(data[6] & 7)
	return nil
}

func (op *VM5FMOperator) Bytes(forYMF825 bool) []byte {
	sus := op.SUS
	ws := op.WS & 31
	if forYMF825 {
		sus = false
		if ws < 0 || ws == 15 || ws == 23 || 31 <= ws {
			fmt.Printf("Invalid wave shape %d\n", ws)
		}
	}
	return []byte{
		byte(op.SR&15)<<4 | util.BoolToByte(op.XOF, 0x08) | util.BoolToByte(sus, 0x02) | util.BoolToByte(op.KSR, 0x01),
		byte(op.RR&15)<<4 | byte(op.DR&15),
		byte(op.AR&15)<<4 | byte(op.SL&15),
		byte(op.TL&63)<<2 | byte(op.KSL&3),
		byte(op.DAM&3)<<5 | util.BoolToByte(op.EAM, 0x10) | byte(op.DVB&3)<<1 | util.BoolToByte(op.EVB, 0x01),
		byte(op.MULTI&15)<<4 | byte(op.DT&7),
		byte(ws)<<3 | byte(op.FB&7),
	}
}

func (op *VM5FMOperator) String() string {
	t := []string{
		fmt.Sprintf("ADSR=%d,%d,%d,%d", op.AR, op.DR, op.SR, op.RR),
		fmt.Sprintf("SL=%d", op.SL),
		fmt.Sprintf("TL=%d", op.TL),
		fmt.Sprintf("KSL=%d", op.KSL),
		fmt.Sprintf("FB=%d", op.FB),
		fmt.Sprintf("WS=%d", op.WS),
	}
	if op.EAM {
		t = append(t, fmt.Sprintf("AM=%d", op.DAM))
	}
	if op.EVB {
		t = append(t, fmt.Sprintf("VB=%d", op.DVB))
	}
	if op.XOF {
		t = append(t, "XOF")
	}
	if op.SUS {
		t = append(t, "SUS")
	}
	if op.KSR {
		t = append(t, "KSR")
	}
	s := strings.Join(t, " ")
	return fmt.Sprintf("Op #%d: MULTI=%s DT=%d\n", op.Num+1, op.MULTI, op.DT) + util.Indent(s, "\t")
}

type vm5FMVoiceRawData struct {
	DrumKey   uint8
	Enigma    uint8
	Global    uint16
	Operators [4][7]uint8
}

type VM5FMVoice struct {
	DrumKey   enums.Note        `json:"drum_key"`
	PANPOT    enums.Panpot      `json:"panpot"` // Panpot (unused in YMF825)
	BO        enums.BasicOctave `json:"bo"`
	LFO       int               `json:"lfo"`
	PE        bool              `json:"pe"` // Panpot Enable (unused in YMF825)
	ALG       enums.Algorithm   `json:"alg"`
	Operators [4]*VM5FMOperator `json:"operators"`
}

func NewVM5FMVoice(data []byte) (*VM5FMVoice, error) {
	voice := &VM5FMVoice{}
	rest := len(data)
	err := voice.Read(bytes.NewReader(data), &rest)
	if err != nil {
		return nil, errors.Wrapf(err, "NewVM5FMVoice invalid data: %s", util.Hex(data))
	}
	if rest != 0 {
		return nil, fmt.Errorf("Wrong size of VM5 voice data (want %d, got %d bytes)", len(data)-rest, len(data))
	}
	return voice, nil
}

func (v *VM5FMVoice) Read(rdr io.Reader, rest *int) error {
	//          | 7 | 6 | 5 | 4 | 3 | 2 | 1 | 0 |
	// Global+0 |            DrumKey            |
	// Global+1 |       PANPOT      | - |  B O  |
	// Global+2 |  LFO  |PE |   -   |    ALG    |

	var global [3]uint8
	err := binary.Read(rdr, binary.BigEndian, &global)
	if err != nil {
		return errors.Wrapf(err, "VM5FMVoice read failed")
	}
	*rest -= int(unsafe.Sizeof(global))
	v.DrumKey = enums.Note(global[0])
	v.PANPOT = enums.Panpot(global[1] >> 3)
	v.BO = enums.BasicOctave(global[1] & 3)
	v.LFO = int(global[2] >> 6 & 3)
	v.PE = global[2]&0x20 != 0
	v.ALG = enums.Algorithm(global[2] & 7)
	v.Operators = [4]*VM5FMOperator{}
	n := v.ALG.OperatorCount()
	for op := 0; op < 4; op++ {
		v.Operators[op] = &VM5FMOperator{Num: op}
	}
	for op := 0; op < n; op++ {
		err := v.Operators[op].Read(rdr, rest)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (v *VM5FMVoice) ReadUnusedRest(rdr io.Reader, rest *int) error {
	n := v.ALG.OperatorCount()
	for op := n; op < 4; op++ {
		v.Operators[op] = &VM5FMOperator{Num: op}
		err := v.Operators[op].Read(rdr, rest)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (v *VM5FMVoice) Bytes(staticLen bool, forYMF825 bool) []byte {
	pan := v.PANPOT
	pe := v.PE
	if forYMF825 {
		pan = 0
		pe = false
	}
	b := []byte{
		byte(pan&31)<<3 | byte(v.BO&3),
		byte(v.LFO&3)<<6 | util.BoolToByte(pe, 0x20) | byte(v.ALG&7),
	}
	n := 4
	if !staticLen {
		n = v.ALG.OperatorCount()
	}
	for op := 0; op < n; op++ {
		b = append(b, v.Operators[op].Bytes(forYMF825)...)
	}
	return b
}

func (v *VM5FMVoice) String() string {
	s := []string{}
	//s = append(s, fmt.Sprintf("Flag: 0x%02X", v.Flag))
	s = append(s, fmt.Sprintf("DrumKey=%s PANPOT=%s LFO=%d PE=%v ALG=%s", v.DrumKey.String(), v.PANPOT, v.LFO, v.PE, v.ALG))
	//s = append(s, fmt.Sprintf("Enigma1: 0x%04X", v.Enigma1))
	//s = append(s, fmt.Sprintf("Enigma2: 0x%08X", v.Enigma2))
	for op := 0; op < v.ALG.OperatorCount(); op++ {
		s = append(s, v.Operators[op].String())
	}
	s = append(s, "Raw="+util.Hex(v.Bytes(false, false)))
	return strings.Join(s, "\n")
}
