package enums

import (
	"fmt"
	"math"

	"github.com/but80/go-smaf/v2/log"
)

// Note は、ノートナンバーを表す整数型です。
type Note int

var noteName = []string{
	"C",
	"C#",
	"D",
	"D#",
	"E",
	"F",
	"F#",
	"G",
	"G#",
	"A",
	"A#",
	"B",
}

//	|C3		|261.6	|4	|357|
//	|C#3	|277.2	|4	|378|
//	|D3		|293.7	|4	|401|
//	|D#3	|311.1	|4	|425|
//	|E3		|329.6	|4	|450|
//	|F3		|349.2	|4	|477|
//	|F#3	|370	|4	|505|
//	|G3		|392	|4	|535|
//	|G#3	|415.3	|4	|567|
//	|A3		|440	|4	|601|
//	|A#3	|466.2	|4	|637|
//	|B3		|493.9	|4	|674|

// NoteFreq は、BLOCK と FNUM で表現される周波数です。
type NoteFreq struct {
	Block, Fnum int
}

const (
	// NoteA3 は、A3のノートナンバーです。
	NoteA3 = 9 + 12*3
)

func (n Note) String() string {
	return fmt.Sprintf("%s(%d)", n.Name(), int(n))
}

// Name は、このノートナンバーにあたる音階の名前を返します。
func (n Note) Name() string {
	i := int(n)
	return fmt.Sprintf("%s%d", noteName[i%12], i/12-1)
}

var fnumK = math.Pow(2.0, 19.0) / 48000.0 / 2.0

// Freq は、このノートナンバーにあたる音階の周波数を返します。
func (n Note) Freq(delta float64) NoteFreq {
	f := 440 * math.Pow(2.0, (float64(n-NoteA3)+delta)/12.0)
	block := int(n) / 12
	if block < 0 {
		block = 0
	} else if 7 < block {
		block = 7
	}
	fnum := 0
	for {
		fnum = int(math.Floor(.5 + f*fnumK/math.Pow(2.0, float64(block))))
		if fnum < 0 {
			if 0 < block {
				block--
				continue
			}
			log.Warnf("Too low fnum: %s", n)
			fnum = 0
		} else if 1024 <= fnum {
			if block < 7 {
				block++
				continue
			}
			log.Warnf("Too high fnum: %s", n)
			fnum = 1023
		}
		break
	}
	return NoteFreq{
		Block: block,
		Fnum:  fnum,
	}
}
