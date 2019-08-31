package voice

import (
	"encoding/binary"
	"io"
	"os"
	"strings"
	"unsafe"

	"github.com/but80/go-smaf/v2/internal/util"
	"github.com/pkg/errors"
)

// VMAVoiceLib は、MA-2用音色ライブラリです。
type VMAVoiceLib struct {
	Programs []*VMAVoicePC `json:"programs"`
}

// Read は、バイト列を読み取ってパースした結果をこの構造体に格納します。
func (lib *VMAVoiceLib) Read(rdr io.Reader, rest *int) error {
	lib.Programs = []*VMAVoicePC{}
	for pc := 0; pc < 128 && 0 < *rest; pc++ {
		voice := &VMAVoicePC{}
		name := [16]byte{}
		err := binary.Read(rdr, binary.BigEndian, &name)
		*rest -= int(unsafe.Sizeof(name))
		if err != nil {
			return errors.WithStack(err)
		}
		voice.Name = util.ZeroPadSliceToString(name[:])
		lib.Programs = append(lib.Programs, voice)
	}
	for pc := 0; pc < 128 && 0 < *rest; pc++ {
		voice := lib.Programs[pc]
		err := voice.Read(rdr, rest)
		if err != nil {
			return errors.WithStack(err)
		}
		lib.Programs = append(lib.Programs, voice)
	}
	return nil
}

func (lib *VMAVoiceLib) String() string {
	s := []string{}
	for _, v := range lib.Programs {
		s = append(s, v.String())
	}
	return strings.Join(s, "\n")
}

// NewVMAVoiceLib は、指定したファイル内容をパースして新しい VMAVoiceLib を作成します。
func NewVMAVoiceLib(file string) (*VMAVoiceLib, error) {
	fh, err := os.Open(file)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer fh.Close()

	var hdr chunkHeader
	if err := binary.Read(fh, binary.BigEndian, &hdr); err != nil {
		return nil, errors.WithStack(err)
	}
	if hdr.Signature != 'F'<<24|'M'<<16|' '<<8|' ' {
		return nil, errors.Errorf(`Header signature must be "FM  "`)
	}

	total := int(hdr.Size) + int(unsafe.Sizeof(hdr))
	rest := int(hdr.Size)
	lib := &VMAVoiceLib{}
	if err := lib.Read(fh, &rest); err != nil {
		return nil, errors.Wrapf(err, "at 0x%X bytes", total-rest)
	}

	return lib, nil
}
