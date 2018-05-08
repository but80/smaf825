package voice

import (
	"encoding/binary"
	"io"
	"os"
	"strings"
	"unsafe"

	"github.com/pkg/errors"
)

type VM3VoiceLib struct {
	Programs []*VM35VoicePC `json:"programs"`
}

func (lib *VM3VoiceLib) Read(rdr io.Reader, rest *int) error {
	lib.Programs = []*VM35VoicePC{}
	for pc := 0; pc < 128 && 0 < *rest; pc++ {
		voice := &VM35VoicePC{Version: VM35FMVoiceVersion_VM3Lib}
		err := voice.Read(rdr, rest)
		if err != nil {
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

func NewVM3VoiceLib(file string) (*VM3VoiceLib, error) {
	fh, err := os.Open(file)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer fh.Close()

	var hdr chunkHeader
	err = binary.Read(fh, binary.BigEndian, &hdr)
	if hdr.Signature != 'F'<<24|'M'<<16|'M'<<8|'3' {
		return nil, errors.Errorf(`Header signature must be "FMM3"`)
	}

	total := int(hdr.Size) + int(unsafe.Sizeof(hdr))
	rest := int(hdr.Size)
	lib := &VM3VoiceLib{}
	err = lib.Read(fh, &rest)
	if err != nil {
		return lib, errors.Wrapf(err, "at 0x%X bytes", total-rest)
	}

	return lib, nil
}
