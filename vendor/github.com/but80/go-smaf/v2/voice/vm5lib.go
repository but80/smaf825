package voice

import (
	"encoding/binary"
	"io"
	"os"
	"strings"
	"unsafe"

	pb "github.com/but80/go-smaf/v2/pb/smaf"
	"github.com/pkg/errors"
)

// VM5VoiceLib は、MA-5用音色ライブラリです。
type VM5VoiceLib struct {
	Programs []*VM35VoicePC `json:"programs"`
}

// ToPB は、この構造体の内容を Protocol Buffer 形式で出力可能な型に変換します。
func (lib *VM5VoiceLib) ToPB() *pb.VM5VoiceLib {
	result := &pb.VM5VoiceLib{
		Programs: make([]*pb.VM35VoicePC, len(lib.Programs)),
	}
	for i, pc := range lib.Programs {
		result.Programs[i] = pc.ToPB()
	}
	return result
}

// Read は、バイト列を読み取ってパースした結果をこの構造体に格納します。
func (lib *VM5VoiceLib) Read(rdr io.Reader, rest *int) error {
	lib.Programs = []*VM35VoicePC{}
	for pc := 0; pc < 128 && 0 < *rest; pc++ {
		voice := &VM35VoicePC{Version: VM35FMVoiceVersionVM5}
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

// NewVM5VoiceLib は、指定したファイル内容をパースして新しい VM5VoiceLib を作成します。
func NewVM5VoiceLib(file string) (*VM5VoiceLib, error) {
	fh, err := os.Open(file)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer fh.Close()

	var hdr chunkHeader
	if err := binary.Read(fh, binary.BigEndian, &hdr); err != nil {
		return nil, errors.WithStack(err)
	}
	if hdr.Signature != 'V'<<24|'O'<<16|'M'<<8|'5' {
		return nil, errors.Errorf(`Header signature must be "VOM5"`)
	}

	total := int(hdr.Size) + int(unsafe.Sizeof(hdr))
	rest := int(hdr.Size)
	lib := &VM5VoiceLib{}
	if err := lib.Read(fh, &rest); err != nil {
		return nil, errors.Wrapf(err, "at 0x%X bytes", total-rest)
	}

	return lib, nil
}
