package voice

import (
	"encoding/binary"
	"io"
	"os"
	"strings"

	"unsafe"

	"github.com/pkg/errors"
)

type VM5VoiceLib struct {
	Programs []*VM35VoicePC `json:"programs"`
}

func (lib *VM5VoiceLib) Read(rdr io.Reader, rest *int) error {
	lib.Programs = []*VM35VoicePC{}
	for pc := 0; pc < 128 && 0 < *rest; pc++ {
		voice := &VM35VoicePC{Version: VM35FMVoiceVersion_VM5}
		err := voice.Read(rdr, rest)
		if err != nil {
			return errors.WithStack(err)
		}
		lib.Programs = append(lib.Programs, voice)
	}
	return nil
}

func (lib *VM5VoiceLib) String() string {
	s := []string{}
	for _, v := range lib.Programs {
		s = append(s, v.String())
	}
	return strings.Join(s, "\n")
}

func NewVM5VoiceLib(file string) (*VM5VoiceLib, error) {
	fh, err := os.Open(file)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer fh.Close()

	var hdr chunkHeader
	err = binary.Read(fh, binary.BigEndian, &hdr)
	if hdr.Signature != 'V'<<24|'O'<<16|'M'<<8|'5' {
		return nil, errors.Errorf(`Header signature must be "VOM5"`)
	}

	total := int(hdr.Size) + int(unsafe.Sizeof(hdr))
	rest := int(hdr.Size)
	lib := &VM5VoiceLib{}
	err = lib.Read(fh, &rest)
	if err != nil {
		return nil, errors.Wrapf(err, "at 0x%X bytes", total-rest)
	}

	return lib, nil
}
