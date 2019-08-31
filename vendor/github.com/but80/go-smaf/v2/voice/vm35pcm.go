package voice

import (
	"encoding/binary"
	"io"
	"unsafe"

	"github.com/but80/go-smaf/v2/internal/util"
	pb "github.com/but80/go-smaf/v2/pb/smaf"
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

// VM35PCMVoice は、MA-3/MA-5用音色データで、1つのプログラムチェンジに含まれるPCM音色部に相当します。
type VM35PCMVoice struct {
	RawData [19]byte `json:"raw_data"`
}

// ToPB は、この構造体の内容を Protocol Buffer 形式で出力可能な型に変換します。
func (v *VM35PCMVoice) ToPB() *pb.VM35PCMVoice {
	return &pb.VM35PCMVoice{
		RawData: v.RawData[:],
	}
}

// Read は、バイト列を読み取ってパースした結果をこの構造体に格納します。
func (v *VM35PCMVoice) Read(rdr io.Reader, rest *int) error {
	err := binary.Read(rdr, binary.BigEndian, &v.RawData)
	if err != nil {
		return errors.WithStack(err)
	}
	*rest -= int(unsafe.Sizeof(v.RawData))
	return nil
}

// ReadUnusedRest は、実際には使用しないバイト列をストリームの残りから読み取り、ヘッダの位置を合わせます。
func (v *VM35PCMVoice) ReadUnusedRest(rdr io.Reader, rest *int) error {
	return nil
}

func (v *VM35PCMVoice) String() string {
	return util.Hex(v.RawData[:])
}
