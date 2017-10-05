package enums

import (
	"fmt"
	"math"
)

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

type NoteFreq struct {
	Block, Fnum int
}

var freqTable [128]NoteFreq

const (
	Note_A3 = 9 + 12*3
)

func init() {
	for n := Note(0); n < 128; n++ {
		f := 440 * math.Pow(2.0, float64(n-Note_A3)/12.0)
		block := int(n) / 12
		if block < 0 {
			block = 0
		} else if 7 < block {
			block = 7
		}
		freqTable[n] = NoteFreq{
			Block: block,
			Fnum:  int(math.Floor(.5 + f*fnumK/math.Pow(2.0, float64(block)))),
		}
	}
}

func (n Note) String() string {
	return fmt.Sprintf("%s(%d)", n.Name(), int(n))
}

func (n Note) Name() string {
	i := int(n)
	return fmt.Sprintf("%s%d", noteName[i%12], i/12-1)
}

var fnumK = math.Pow(2.0, 19.0) / 48000.0 / 2.0

func (n Note) Freq(delta float64) NoteFreq {
	//return freqTable[n]
	f := 440 * math.Pow(2.0, (float64(n-Note_A3)+delta)/12.0)
	block := int(n) / 12
	if block < 0 {
		block = 0
	} else if 7 < block {
		block = 7
	}
	return NoteFreq{
		Block: block,
		Fnum:  int(math.Floor(.5 + f*fnumK/math.Pow(2.0, float64(block)))),
	}
}
