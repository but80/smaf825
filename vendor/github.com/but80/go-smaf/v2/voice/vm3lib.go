package voice

import (
	"encoding/binary"
	"io"
	"os"
	"strings"
	"unsafe"

	"github.com/pkg/errors"
)

// VM3VoiceLib は、MA-3用音色ライブラリです。
type VM3VoiceLib struct {
	Programs []*VM35VoicePC `json:"programs"`
}

// Read は、バイト列を読み取ってパースした結果をこの構造体に格納します。
func (lib *VM3VoiceLib) Read(rdr io.Reader, rest *int) error {
	lib.Programs = []*VM35VoicePC{}
	for pc := 0; pc < 128 && 0 < *rest; pc++ {
		voice := &VM35VoicePC{Version: VM35FMVoiceVersionVM3Lib}
		if err := voice.Read(rdr, rest); err != nil {
			return errors.WithStack(err)
		}
		lib.Programs = append(lib.Programs, voice)
	}
	return nil
}

func (lib *VM3VoiceLib) String() string {
	s := []string{}
	for _, v := range lib.Programs {
		s = append(s, v.String())
	}
	return strings.Join(s, "\n")
}

// NewVM3VoiceLib は、指定したファイル内容をパースして新しい VM3VoiceLib を作成します。
func NewVM3VoiceLib(file string) (*VM3VoiceLib, error) {
	fh, err := os.Open(file)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer fh.Close()

	var hdr chunkHeader
	if err := binary.Read(fh, binary.BigEndian, &hdr); err != nil {
		return nil, errors.WithStack(err)
	}
	if hdr.Signature != 'F'<<24|'M'<<16|'M'<<8|'3' {
		return nil, errors.Errorf(`Header signature must be "FMM3"`)
	}

	total := int(hdr.Size) + int(unsafe.Sizeof(hdr))
	rest := int(hdr.Size)
	lib := &VM3VoiceLib{}
	if err := lib.Read(fh, &rest); err != nil {
		return lib, errors.Wrapf(err, "at 0x%X bytes", total-rest)
	}

	return lib, nil
}
