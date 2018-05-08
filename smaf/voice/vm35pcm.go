package voice

import (
	"encoding/binary"
	"io"
	"unsafe"

	"github.com/but80/smaf825/pb"

	"github.com/but80/smaf825/smaf/util"
	"github.com/pkg/errors"
)

//     | 7 | 6 | 5 | 4 | 3 | 2 | 1 | 0 |
// + 0 |             Fs(H)             |
// + 1 |             Fs(L)             |
// + 2 |      PANPOT       |   ?   |P E|
// + 3 |  LFO  |           ?           |
// + 4 |      S R      |XOF|   |SUS|   |
// + 5 |      R R      |      D R      |
// + 6 |      A R      |      S L      |
// + 7 |          T L          |   ?   |
// + 8 | ? |  DAM  |EAM| ? |  DVB  |EVB|
// + 9 |               ?               |
// +10 |               ?               |
// +11 |             LP(H)             |
// +12 |             LP(L)             |
// +13 |             EP(H)             |
// +14 |             EP(L)             |
// +15 |R M|         ...WaveID         |
// +16 |               ?               |
// +17 |               ?               |
// +18 |               ?               |

type VM35PCMVoice struct {
	RawData [19]byte `json:"raw_data"`
}

func (v *VM35PCMVoice) ToPB() *pb.VM35PCMVoice {
	return &pb.VM35PCMVoice{
		RawData: v.RawData[:],
	}
}

func (v *VM35PCMVoice) Read(rdr io.Reader, rest *int) error {
	err := binary.Read(rdr, binary.BigEndian, &v.RawData)
	if err != nil {
		return errors.WithStack(err)
	}
	*rest -= int(unsafe.Sizeof(v.RawData))
	return nil
}

func (v *VM35PCMVoice) ReadUnusedRest(rdr io.Reader, rest *int) error {
	return nil
}

func (v *VM35PCMVoice) String() string {
	return util.Hex(v.RawData[:])
}
