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
	"github.com/but80/go-smaf/v2/log"
	pb "github.com/but80/go-smaf/v2/pb/smaf"
	"github.com/pkg/errors"
)

type VM35FMVoiceVersion int

const (
	VM35FMVoiceVersionVM3Lib VM35FMVoiceVersion = iota
	VM35FMVoiceVersionVM3Exclusive
	VM35FMVoiceVersionVM5
)

type VM35FMOperator struct {
	Num     int                `json:"-"` // Operator number
	Version VM35FMVoiceVersion `json:"-"`
	MULTI   enums.Multiplier   `json:"multi"` // Multiplier
	DT      int                `json:"dt"`    // Detune
	AR      int                `json:"ar"`    // Attack Rate
	DR      int                `json:"dr"`    // Decay Rate
	SR      int                `json:"sr"`    // Sustain Rate
	RR      int                `json:"rr"`    // Release Rate
	SL      int                `json:"sl"`    // Sustain Level
	TL      int                `json:"tl"`    // Total Level
	KSL     int                `json:"ksl"`   // Key Scaling Level
	DAM     int                `json:"dam"`   // Depth of AM
	DVB     int                `json:"dvb"`   // Depth of Vibrato
	FB      int                `json:"fb"`    // Feedback
	WS      int                `json:"ws"`    // Wave Shape
	XOF     bool               `json:"xof"`   // Ignore KeyOff
	SUS     bool               `json:"sus"`   // Keep sustain rate after KeyOff (unused in YMF825)
	KSR     bool               `json:"ksr"`   // Key Scaling Rate
	EAM     bool               `json:"eam"`   // Enable AM
	EVB     bool               `json:"evb"`   // Enable Vibrato
}

// ToPB は、この構造体の内容を Protocol Buffer 形式で出力可能な型に変換します。
func (op *VM35FMOperator) ToPB() *pb.VM35FMOperator {
	return &pb.VM35FMOperator{
		Multi: uint32(op.MULTI),
		Dt:    uint32(op.DT),
		Ar:    uint32(op.AR),
		Dr:    uint32(op.DR),
		Sr:    uint32(op.SR),
		Rr:    uint32(op.RR),
		Sl:    uint32(op.SL),
		Tl:    uint32(op.TL),
		Ksl:   uint32(op.KSL),
		Dam:   uint32(op.DAM),
		Dvb:   uint32(op.DVB),
		Fb:    uint32(op.FB),
		Ws:    uint32(op.WS),
		Xof:   op.XOF,
		Sus:   op.SUS,
		Ksr:   op.KSR,
		Eam:   op.EAM,
		Evb:   op.EVB,
	}
}

// Read は、バイト列を読み取ってパースした結果をこの構造体に格納します。
func (op *VM35FMOperator) Read(rdr io.Reader, rest *int) error {
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

func (op *VM35FMOperator) Bytes(forYMF825 bool) []byte {
	sus := op.SUS
	ws := op.WS & 31
	if forYMF825 {
		sus = false
		if ws < 0 || ws == 15 || ws == 23 || 31 <= ws {
			log.Warnf("Invalid wave shape %d", ws)
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

func (op *VM35FMOperator) String() string {
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

// VM35FMVoice は、MA-3/MA-5用音色データで、1つのプログラムチェンジに含まれるFM音色部に相当します。
type VM35FMVoice struct {
	Version   VM35FMVoiceVersion `json:"-"`
	DrumKey   enums.Note         `json:"drum_key"`
	PANPOT    enums.Panpot       `json:"panpot"` // Panpot (unused in YMF825)
	BO        enums.BasicOctave  `json:"bo"`
	LFO       int                `json:"lfo"`
	PE        bool               `json:"pe"` // Panpot Enable (unused in YMF825)
	ALG       enums.Algorithm    `json:"alg"`
	Operators [4]*VM35FMOperator `json:"operators"`
}

// NewVM35FMVoice は、指定したバイト列をパースして新しい VM35FMVoice を作成します。
func NewVM35FMVoice(data []byte, version VM35FMVoiceVersion) (*VM35FMVoice, error) {
	voice := &VM35FMVoice{Version: version}
	rest := len(data)
	rdr := bytes.NewReader(data)
	err := voice.Read(rdr, &rest)
	if err != nil {
		return nil, errors.Wrapf(err, "NewVM35FMVoice invalid data: %s (want %d, got %d bytes)", util.Hex(data), len(data)-rest, len(data))
	}
	switch version {
	case VM35FMVoiceVersionVM3Lib, VM35FMVoiceVersionVM3Exclusive:
		voice.ReadUnusedRest(rdr, &rest)
	}
	if rest != 0 {
		return nil, fmt.Errorf("Wrong size of VM3/VM5 voice data (want %d, got %d bytes): %s", len(data)-rest, len(data), util.Hex(data))
	}
	return voice, nil
}

// ToPB は、この構造体の内容を Protocol Buffer 形式で出力可能な型に変換します。
func (v *VM35FMVoice) ToPB() *pb.VM35FMVoice {
	result := &pb.VM35FMVoice{
		DrumKey:   uint32(v.DrumKey),
		Panpot:    uint32(v.PANPOT),
		Bo:        uint32(v.BO),
		Lfo:       uint32(v.LFO),
		Pe:        v.PE,
		Alg:       uint32(v.ALG),
		Operators: make([]*pb.VM35FMOperator, 4),
	}
	for i, op := range v.Operators {
		result.Operators[i] = op.ToPB()
	}
	return result
}

// Read は、バイト列を読み取ってパースした結果をこの構造体に格納します。
func (v *VM35FMVoice) Read(rdr io.Reader, rest *int) error {
	switch v.Version {
	case VM35FMVoiceVersionVM3Exclusive:
		//    | 7 | 6 | 5 | 4 | 3 | 2 | 1 | 0 |
		// ------------------------------------ Global
		// +0 |       |PN4|LF1|SR3|RR3|AR3|TL5|  // bit0-3は1つ次のOpに作用
		// +1 |                               |  // Drumkey?
		// +2 | - |   PAN0123     |       | ? |
		// +3 | - |LF0|P E|       |    ALG    |
		// ------------------------------------ Op0
		// +4 | - |   SR012   |XOF| - |SUS|KSR|
		// +5 | - |   RR012   |      D R      |
		// +6 | - |   AR012   |      S L      |
		// +7 | - |      TL01234      |  KSL  |
		// +8 |   -   |ML3|WS4|SR3|RR3|AR3|TL5|  // bit0-3は1つ次のOpに作用
		// +9 | - |  DAM  |EAM| - |  DVB  |EVB|
		// +A | - |  MUL012   | - |    DT     |
		// +B | - |     WS0123    |    FB     |
		// ------------------------------------ Op1
		// ...
		raw := make([]byte, 4+8*4)
		n, err := rdr.Read(raw)
		if err != nil {
			return errors.WithStack(err)
		}
		*rest -= n
		raw[2] |= raw[0] << 2 & 0x80
		raw[3] |= raw[0] << 3 & 0x80
		for op := 0; op < 4; op++ {
			raw[4+op*8] |= raw[op*8] << 4 & 0x80
			raw[5+op*8] |= raw[op*8] << 5 & 0x80
			raw[6+op*8] |= raw[op*8] << 6 & 0x80
			raw[7+op*8] |= raw[op*8] << 7 & 0x80
			raw[10+op*8] |= raw[8+op*8] << 2 & 0x80
			raw[11+op*8] |= raw[8+op*8] << 3 & 0x80
		}
		//    | 7 | 6 | 5 | 4 | 3 | 2 | 1 | 0 |
		// ------------------------------------ Global
		// +0 |                               |
		// +1 |                               |  // Drumkey?
		// +2 |      PANPOT       |       | ? |
		// +3 |  LFO  |P E|       |    ALG    |
		// ------------------------------------ Op0
		// +4 |      S R      |XOF| - |SUS|KSR|
		// +5 |      R R      |      D R      |
		// +6 |      A R      |      S L      |
		// +7 |         T L           |  KSL  |
		// +8 |                               |
		// +9 | - |  DAM  |EAM| - |  DVB  |EVB|
		// +A |      MUL      | - |    DT     |
		// +B |        W S        |    FB     |
		// ------------------------------------ Op1
		// ...
		fixed := raw[1:4]
		for op := 0; op < 4; op++ {
			fixed = append(fixed, raw[4+op*8:8+op*8]...)
			fixed = append(fixed, raw[9+op*8:12+op*8]...)
		}
		rdr = bytes.NewReader(fixed)
		restVal := len(fixed)
		rest = &restVal
	}

	//          | 7 | 6 | 5 | 4 | 3 | 2 | 1 | 0 |
	// Global+0 |            DrumKey            |
	// Global+1 |       PANPOT      | - |  B O  |
	// Global+2 |  LFO  |PE |   -   |    ALG    |

	var global [3]uint8
	err := binary.Read(rdr, binary.BigEndian, &global)
	if err != nil {
		return errors.Wrapf(err, "VM35FMVoice read failed")
	}
	*rest -= int(unsafe.Sizeof(global))
	v.DrumKey = enums.Note(global[0])
	v.PANPOT = enums.Panpot(global[1] >> 3)
	v.BO = enums.BasicOctave(global[1] & 3)
	v.LFO = int(global[2] >> 6 & 3)
	v.PE = global[2]&0x20 != 0
	v.ALG = enums.Algorithm(global[2] & 7)
	v.Operators = [4]*VM35FMOperator{}
	n := v.ALG.OperatorCount()
	for op := 0; op < 4; op++ {
		v.Operators[op] = &VM35FMOperator{Version: v.Version, Num: op}
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
func (v *VM35FMVoice) ReadUnusedRest(rdr io.Reader, rest *int) error {
	// 2オペレータ音色の第3・第4オペレータ部の読み取り
	n := v.ALG.OperatorCount()
	for op := n; op < 4; op++ {
		v.Operators[op] = &VM35FMOperator{Num: op}
		err := v.Operators[op].Read(rdr, rest)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func (v *VM35FMVoice) Bytes(staticLen bool, forYMF825 bool) []byte {
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

type vm35FMVoiceMarshaler VM35FMVoice

func (v VM35FMVoice) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		vm35FMVoiceMarshaler
		YMF825Data []int `json:"ymf825_data"`
	}{
		vm35FMVoiceMarshaler: vm35FMVoiceMarshaler(v),
		YMF825Data:           util.BytesToInts(v.Bytes(true, true)),
	})
}

func (v *VM35FMVoice) String() string {
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

// NewDemoVM35FMVoice は、デモ音色として初期化された新しい VM35FMVoice を作成します。
func NewDemoVM35FMVoice() *VM35FMVoice {
	v := &VM35FMVoice{
		Version:   VM35FMVoiceVersionVM5,
		DrumKey:   enums.NoteA3,
		PANPOT:    enums.PanpotCenter,
		BO:        enums.BasicOctaveNormal,
		ALG:       enums.Algorithm(0),
		Operators: [4]*VM35FMOperator{},
	}
	for i := 0; i < 4; i++ {
		v.Operators[i] = &VM35FMOperator{}
	}
	op1 := &VM35FMOperator{
		MULTI: 1,
		AR:    14,
		DR:    2,
		SR:    8,
		SL:    8,
		RR:    8,
	}
	v.Operators[1] = op1
	return v
}
