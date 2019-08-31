package voice

import (
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"

	"github.com/but80/go-smaf/v2/enums"
	"github.com/but80/go-smaf/v2/internal/util"
	"github.com/pkg/errors"
)

// VMAVoicePC は、MA-2用音色データで、1つのプログラムチェンジに相当します。
type VMAVoicePC struct {
	Name  string      `json:"name"`
	Bank  int         `json:"bank"`
	PC    int         `json:"pc"`
	Voice *VMAFMVoice `json:"voice"`
}

type vmaVoicePCHeaderRawData struct {
	//    | 7 | 6 | 5 | 4 | 3 | 2 | 1 | 0 |
	// +0 |              00?              |
	// +1 |             Bank              |
	// +2 |              PC               |
	Enigma uint8
	Bank   uint8
	PC     uint8
}

// Read は、バイト列を読み取ってパースした結果をこの構造体に格納します。
func (p *VMAVoicePC) Read(rdr io.Reader, rest *int) error {
	var data vmaVoicePCHeaderRawData
	err := binary.Read(rdr, binary.BigEndian, &data)
	if err != nil {
		return errors.WithStack(err)
	}
	*rest -= int(unsafe.Sizeof(data))
	p.Bank = int(data.Bank)
	p.PC = int(data.PC)
	p.Voice = &VMAFMVoice{}
	p.Voice.Read(rdr, rest)
	p.Voice.ReadUnusedRest(rdr, rest)
	//
	var enigma2 uint8
	err = binary.Read(rdr, binary.BigEndian, &enigma2)
	if err != nil {
		return errors.WithStack(err)
	}
	*rest--
	return nil
}

func (p *VMAVoicePC) String() string {
	s := fmt.Sprintf("Bank %d @%d", p.Bank, p.PC)
	if p.Name != "" {
		s += fmt.Sprintf(": [%s]", p.Name)
	}
	s += "\n"
	return s + util.Indent(p.Voice.String(), "\t")
}

// ToVM35 は、この構造体の内容をMA-3/MA-5用の音色データに変換します。
func (p *VMAVoicePC) ToVM35() *VM35VoicePC {
	return &VM35VoicePC{
		Name:      p.Name,
		Flag:      0x24,
		BankMSB:   0,
		BankLSB:   p.Bank,
		PC:        p.PC,
		DrumNote:  0,
		Enigma1:   0,
		VoiceType: enums.VoiceTypeFM,
		Voice:     p.Voice.ToVM35(),
	}
}
