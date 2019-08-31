package voice

import (
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"

	"github.com/but80/go-smaf/v2/enums"
	"github.com/but80/go-smaf/v2/internal/util"
	pb "github.com/but80/go-smaf/v2/pb/smaf"
	"github.com/pkg/errors"
)

// VM35Voice は、MA-3/MA-5用音色データで、1つのプログラムチェンジに含まれる音色部を抽象化したインタフェースです。
type VM35Voice interface {
	fmt.Stringer
	Read(rdr io.Reader, rest *int) error
	ReadUnusedRest(rdr io.Reader, rest *int) error
}

// VM35VoicePC は、MA-3/MA-5用音色データで、1つのプログラムチェンジに相当します。
type VM35VoicePC struct {
	Version   VM35FMVoiceVersion `json:"is_vm5"`
	Name      string             `json:"name"`
	Flag      int                `json:"-"` // = 0x24
	BankMSB   int                `json:"bank_msb"`
	BankLSB   int                `json:"bank_lsb"`
	PC        int                `json:"pc"`
	DrumNote  enums.Note         `json:"drum_note"`
	Enigma1   int                `json:"-"`
	VoiceType enums.VoiceType    `json:"voice_type"`
	Voice     VM35Voice          `json:"voice"`
}

type vm3VoicePCHeaderRawData struct {
	Enigma1   uint16
	Flag      uint8
	BankMSB   uint8
	BankLSB   uint8
	PC        uint8
	DrumNote  uint8
	VoiceType uint8 // bit0: 0=Type=FM 1=Type=PCM
	Name      [16]byte
}

type vm5VoicePCHeaderRawData struct {
	Enigma1   uint16
	Name      [16]byte
	Flag      uint8
	BankMSB   uint8
	BankLSB   uint8
	PC        uint8
	DrumNote  uint8
	VoiceType uint8 // bit0: 0=Type=FM 1=Type=PCM
}

// ToPB は、この構造体の内容を Protocol Buffer 形式で出力可能な型に変換します。
func (p *VM35VoicePC) ToPB() *pb.VM35VoicePC {
	result := &pb.VM35VoicePC{
		Version:   pb.VM35FMVoiceVersion(p.Version),
		Name:      p.Name,
		BankMsb:   uint32(p.BankMSB),
		BankLsb:   uint32(p.BankLSB),
		Pc:        uint32(p.PC),
		DrumNote:  uint32(p.DrumNote),
		VoiceType: pb.VoiceType(p.VoiceType),
	}
	switch v := p.Voice.(type) {
	case *VM35FMVoice:
		result.FmVoice = v.ToPB()
	case *VM35PCMVoice:
		result.PcmVoice = v.ToPB()
	}
	return result
}

// Read は、バイト列を読み取ってパースした結果をこの構造体に格納します。
func (p *VM35VoicePC) Read(rdr io.Reader, rest *int) error {
	switch p.Version {
	case VM35FMVoiceVersionVM5:
		var data vm5VoicePCHeaderRawData
		err := binary.Read(rdr, binary.BigEndian, &data)
		if err != nil {
			return errors.WithStack(err)
		}
		*rest -= int(unsafe.Sizeof(data))
		p.Name = util.ZeroPadSliceToString(data.Name[:])
		p.Flag = int(data.Flag)
		p.BankMSB = int(data.BankMSB)
		p.BankLSB = int(data.BankLSB)
		p.PC = int(data.PC)
		p.DrumNote = enums.Note(data.DrumNote)
		p.Enigma1 = int(data.Enigma1)
		p.VoiceType = enums.VoiceType(data.VoiceType)
	case VM35FMVoiceVersionVM3Lib:
		var data vm3VoicePCHeaderRawData
		err := binary.Read(rdr, binary.BigEndian, &data)
		if err != nil {
			return errors.WithStack(err)
		}
		*rest -= int(unsafe.Sizeof(data))
		p.Name = util.ZeroPadSliceToString(data.Name[:])
		p.Flag = int(data.Flag)
		p.BankMSB = int(data.BankMSB)
		p.BankLSB = int(data.BankLSB)
		p.PC = int(data.PC)
		p.DrumNote = enums.Note(data.DrumNote)
		p.Enigma1 = int(data.Enigma1)
		p.VoiceType = enums.VoiceType(data.VoiceType)
	}
	switch p.VoiceType {
	case enums.VoiceTypeFM:
		p.Voice = &VM35FMVoice{}
		p.Voice.Read(rdr, rest)
		p.Voice.ReadUnusedRest(rdr, rest)
	case enums.VoiceTypePCM:
		p.Voice = &VM35PCMVoice{}
		p.Voice.Read(rdr, rest)
	//case enums.VoiceTypeAL:
	default:
		return fmt.Errorf("contains unsupported type of voice: %s", p.VoiceType.String())
	}
	return nil
}

func (p *VM35VoicePC) IsForDrum() bool {
	return p.DrumNote != 0
}

func (p *VM35VoicePC) String() string {
	// fmt.Sprintf("Flag: 0x%02X", v.Flag)
	// fmt.Sprintf("Enigma1: 0x%04X", v.Enigma1)
	// fmt.Sprintf("Enigma2: 0x%08X", v.Enigma2)
	s := fmt.Sprintf("Bank %d-%d @%d %s", p.BankMSB, p.BankLSB, p.PC, p.VoiceType.String())
	if p.IsForDrum() {
		s += fmt.Sprintf(" DrumNote=%s\n", p.DrumNote.String())
	}
	if p.Name != "" {
		s += fmt.Sprintf(": [%s]", p.Name)
	}
	s += "\n"
	return s + util.Indent(p.Voice.String(), "\t")
}
