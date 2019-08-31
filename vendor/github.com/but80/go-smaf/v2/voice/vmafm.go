package voice

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unsafe"

	"github.com/but80/go-smaf/v2/enums"
	"github.com/but80/go-smaf/v2/internal/util"
	"github.com/pkg/errors"
)

// VMAFMOperator は、MA-2用音色データに含まれるオペレータ部分です。
type VMAFMOperator struct {
	Num  int              `json:"-"`    // Operator number
	MULT enums.Multiplier `json:"mult"` // Multiplier
	KSL  int              `json:"ksl"`  // Key Scaling Level
	TL   int              `json:"tl"`   // Total Level
	AR   int              `json:"ar"`   // Attack Rate
	DR   int              `json:"dr"`   // Decay Rate
	SL   int              `json:"sl"`   // Sustain Level
	RR   int              `json:"rr"`   // Release Rate
	WS   int              `json:"ws"`   // Wave Shape
	DVB  int              `json:"dvb"`  // Depth of Vibrato
	DAM  int              `json:"dam"`  // Depth of AM
	VIB  bool             `json:"vib"`  // Vibrato
	EGT  bool             `json:"egt"`  //
	SUS  bool             `json:"sus"`  // Keep sustain rate after KeyOff (unused in YMF825)
	KSR  bool             `json:"ksr"`  // Key Scaling Rate
	AM   bool             `json:"am"`   // AM
}

// Read は、バイト列を読み取ってパースした結果をこの構造体に格納します。
func (op *VMAFMOperator) Read(rdr io.Reader, rest *int) error {
	//    | 7 | 6 | 5 | 4 | 3 | 2 | 1 | 0 |
	// +0 |     MULT      |VIB|EGT|SUS|KSR|
	// +1 |      R R      |      D R      |
	// +2 |      A R      |      S L      |
	// +3 |          T L          |  KSL  |
	// +4 |  DVB  |  DAM  |A M|    W S    |

	data := [5]uint8{}
	err := binary.Read(rdr, binary.BigEndian, &data)
	if err != nil {
		return errors.WithStack(err)
	}
	*rest -= int(unsafe.Sizeof(data))

	op.MULT = enums.Multiplier(data[0] >> 4)
	op.VIB = data[0]&0x08 != 0
	op.EGT = data[0]&0x04 != 0
	op.SUS = data[0]&0x02 != 0
	op.KSR = data[0]&0x01 != 0
	op.RR = int(data[1] >> 4)
	op.DR = int(data[1] & 15)
	op.AR = int(data[2] >> 4)
	op.SL = int(data[2] & 15)
	op.TL = int(data[3] >> 2)
	op.KSL = int(data[3] & 3)
	op.DVB = int(data[4] >> 6 & 3)
	op.DAM = int(data[4] >> 4 & 3)
	op.AM = data[4]&0x08 != 0
	op.WS = int(data[4] & 7)
	return nil
}

func (op *VMAFMOperator) Bytes() []byte {
	return []byte{
		byte(op.MULT&15)<<4 | util.BoolToByte(op.VIB, 0x08) | util.BoolToByte(op.EGT, 0x04) | util.BoolToByte(op.SUS, 0x02) | util.BoolToByte(op.KSR, 0x01),
		byte(op.RR&15)<<4 | byte(op.DR&15),
		byte(op.AR&15)<<4 | byte(op.SL&15),
		byte(op.TL&63)<<2 | byte(op.KSL&3),
		byte(op.DVB&3)<<6 | byte(op.DAM&3)<<4 | util.BoolToByte(op.AM, 0x08) | byte(op.WS&7),
	}
}

func (op *VMAFMOperator) String() string {
	t := []string{
		fmt.Sprintf("ADR=%d,%d,%d", op.AR, op.DR, op.RR),
		fmt.Sprintf("SL=%d", op.SL),
		fmt.Sprintf("TL=%d", op.TL),
		fmt.Sprintf("KSL=%d", op.KSL),
		fmt.Sprintf("WS=%d", op.WS),
	}
	if op.AM {
		t = append(t, fmt.Sprintf("AM=%d", op.DAM))
	}
	if op.VIB {
		t = append(t, fmt.Sprintf("VB=%d", op.DVB))
	}
	if op.EGT {
		t = append(t, "EGT")
	}
	if op.SUS {
		t = append(t, "SUS")
	}
	if op.KSR {
		t = append(t, "KSR")
	}
	s := strings.Join(t, " ")
	return fmt.Sprintf("Op #%d: MULT=%s\n", op.Num+1, op.MULT) + util.Indent(s, "\t")
}

// ToVM35 は、この構造体の内容をMA-3/MA-5用の音色データに変換します。
func (op *VMAFMOperator) ToVM35(fb int) *VM35FMOperator {
	sr := op.RR
	if op.EGT {
		sr = 0
	}
	return &VM35FMOperator{
		Num:   op.Num,
		MULTI: op.MULT,
		DT:    0,
		AR:    op.AR,
		DR:    op.DR,
		SR:    sr,
		RR:    op.RR,
		SL:    op.SL,
		TL:    op.TL,
		KSL:   op.KSL,
		DAM:   op.DAM,
		DVB:   op.DVB,
		FB:    fb,
		WS:    op.WS,
		XOF:   false,
		SUS:   op.SUS,
		KSR:   op.KSR,
		EAM:   op.AM,
		EVB:   op.VIB,
	}
}

// VMAFMVoice は、MA-2用音色データで、1つのプログラムチェンジに含まれる音色部に相当します。
type VMAFMVoice struct {
	LFO       int               `json:"lfo"`
	FB        int               `json:"fb"`
	ALG       enums.Algorithm   `json:"alg"`
	Operators [4]*VMAFMOperator `json:"operators"`
}

// NewVMAFMVoice は、指定したバイト列をパースして新しい VMAFMVoice を作成します。
func NewVMAFMVoice(data []byte) (*VMAFMVoice, error) {
	voice := &VMAFMVoice{}
	rest := len(data)
	rdr := bytes.NewReader(data)
	err := voice.Read(rdr, &rest)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if 0 < rest {
		err = voice.ReadUnusedRest(rdr, &rest)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	if rest != 0 {
		return nil, fmt.Errorf("Wrong size of VMA voice data (want %d, got %d)", len(data)+rest, len(data))
	}
	return voice, nil
}

// Read は、バイト列を読み取ってパースした結果をこの構造体に格納します。
func (v *VMAFMVoice) Read(rdr io.Reader, rest *int) error {
	//    | 7 | 6 | 5 | 4 | 3 | 2 | 1 | 0 |
	// +0 |  LFO  |    F B    |    ALG    |
	// +1 |              01?              |

	var global [2]uint8
	err := binary.Read(rdr, binary.BigEndian, &global)
	if err != nil {
		return errors.WithStack(err)
	}
	*rest -= int(unsafe.Sizeof(global))
	v.LFO = int(global[0] >> 6 & 3)
	v.FB = int(global[0] >> 3 & 7)
	v.ALG = enums.Algorithm(global[0] & 7)
	v.Operators = [4]*VMAFMOperator{}
	n := v.ALG.OperatorCount()
	for op := 0; op < 4; op++ {
		v.Operators[op] = &VMAFMOperator{Num: op}
	}
	for op := 0; op < n; op++ {
		err := v.Operators[op].Read(rdr, rest)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// ReadUnusedRest は、実際には使用しないバイト列をストリームの残りから読み取り、ヘッダの位置を合わせます。
func (v *VMAFMVoice) ReadUnusedRest(rdr io.Reader, rest *int) error {
	// 2オペレータ音色の第3・第4オペレータ部の読み取り
	n := v.ALG.OperatorCount()
	for op := n; op < 4; op++ {
		v.Operators[op] = &VMAFMOperator{Num: op}
		err := v.Operators[op].Read(rdr, rest)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (v *VMAFMVoice) Bytes(staticLen bool) []byte {
	b := []byte{
		byte(v.LFO&3)<<6 | byte(v.FB&7)<<3 | byte(v.ALG&7),
		1,
	}
	n := 4
	if !staticLen {
		n = v.ALG.OperatorCount()
	}
	for op := 0; op < n; op++ {
		b = append(b, v.Operators[op].Bytes()...)
	}
	return b
}

type vmaFMVoiceMarshaler VMAFMVoice

func (v VMAFMVoice) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		vmaFMVoiceMarshaler
		YMF825Data []int `json:"ymf825_data"`
	}{
		vmaFMVoiceMarshaler: vmaFMVoiceMarshaler(v),
		YMF825Data:          util.BytesToInts(v.ToVM35().Bytes(true, true)),
	})
}

func (v *VMAFMVoice) String() string {
	s := []string{}
	s = append(s, fmt.Sprintf("LFO=%d FB=%d ALG=%s", v.LFO, v.FB, v.ALG))
	for op := 0; op < v.ALG.OperatorCount(); op++ {
		s = append(s, v.Operators[op].String())
	}
	s = append(s, "Raw="+util.Hex(v.Bytes(false)))
	return strings.Join(s, "\n")
}

// ToVM35 は、この構造体の内容をMA-3/MA-5用の音色データに変換します。
func (v *VMAFMVoice) ToVM35() *VM35FMVoice {
	result := &VM35FMVoice{
		DrumKey:   enums.Note(0),
		PANPOT:    enums.PanpotCenter,
		BO:        enums.BasicOctaveNormal,
		LFO:       v.LFO,
		PE:        false,
		ALG:       v.ALG,
		Operators: [4]*VM35FMOperator{},
	}
	fb := v.FB
	for op := 0; op < 4; op++ {
		result.Operators[op] = v.Operators[op].ToVM35(fb)
		fb = 0
	}
	return result
}
