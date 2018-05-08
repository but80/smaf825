package voice

import (
	"encoding/binary"
	"io"
	"os"
	"strings"
	"unsafe"

	"github.com/but80/smaf825/smaf/util"
	"github.com/pkg/errors"
)

type VMAVoiceLib struct {
	Programs []*VMAVoicePC `json:"programs"`
}

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

func NewVMAVoiceLib(file string) (*VMAVoiceLib, error) {
	fh, err := os.Open(file)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer fh.Close()

	var hdr chunkHeader
	err = binary.Read(fh, binary.BigEndian, &hdr)
	if hdr.Signature != 'F'<<24|'M'<<16|' '<<8|' ' {
		return nil, errors.Errorf(`Header signature must be "FM  "`)
	}

	total := int(hdr.Size) + int(unsafe.Sizeof(hdr))
	rest := int(hdr.Size)
	lib := &VMAVoiceLib{}
	err = lib.Read(fh, &rest)
	if err != nil {
		return nil, errors.Wrapf(err, "at 0x%X bytes", total-rest)
	}

	return lib, nil
}
