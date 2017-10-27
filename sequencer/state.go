package sequencer

import (
	"fmt"

	"sort"

	"github.com/ahmetalpbalkan/go-cursor"
	"github.com/but80/smaf825/smaf/enums"
	"github.com/but80/smaf825/smaf/log"
	"github.com/but80/smaf825/smaf/voice"
)

type ChannelState struct {
	KeyControlStatus enums.KeyControlStatus
	Velocity         int
	GateTimeRest     map[enums.Note]int
	BankMSB          int
	BankLSB          int
	PC               int
	ToneID           int
	PitchBend        int
	PitchBendRange   int
	Modulation       int
	Volume           int
	Panpot           int
	Expression       int
	OctaveShift      int
	Mono             bool
	RPNMSB           int
	RPNLSB           int
}

func (cs *ChannelState) Tick() []enums.Note {
	notes := []enums.Note{}
	for note, t := range cs.GateTimeRest {
		if t <= 0 {
			delete(cs.GateTimeRest, note)
			continue
		}
		t--
		cs.GateTimeRest[note] = t
		if t == 0 {
			notes = append(notes, note)
			delete(cs.GateTimeRest, note)
		}
	}
	return notes
}

func (cs *ChannelState) AllOff() []enums.Note {
	notes := []enums.Note{}
	for note := range cs.GateTimeRest {
		notes = append(notes, note)
	}
	cs.GateTimeRest = map[enums.Note]int{}
	return notes
}

func (cs *ChannelState) HasRest() bool {
	return 0 < len(cs.GateTimeRest)
}

func (cs *ChannelState) NoteOn(note enums.Note, gateTime int) {
	if cs.KeyControlStatus != enums.KeyControlStatus_Off {
		cs.GateTimeRest = map[enums.Note]int{}
	}
	cs.GateTimeRest[note] = gateTime
}

func (cs *ChannelState) Print(num int) {
	mono := "Off"
	if cs.Mono {
		mono = "On"
	}
	pan := "C"
	if cs.Panpot < 64 {
		pan = fmt.Sprintf("L%d", 64-cs.Panpot)
	} else if 64 < cs.Panpot {
		pan = fmt.Sprintf("R%d", cs.Panpot-64)
	}
	note := "-"
	if 0 < len(cs.GateTimeRest) {
		for n := range cs.GateTimeRest {
			note = n.String()
			break
		}
	}
	fmt.Printf(
		"%2d %3d-%-3d %3d %-4s %5d %3d %3d %3d %-3s %-4s\n",
		num+1,
		cs.BankMSB,
		cs.BankLSB,
		cs.PC,
		note,
		cs.PitchBend,
		cs.Modulation,
		cs.Volume,
		cs.Expression,
		pan,
		mono,
	)
}

type Tones []*voice.VM35VoicePC

func (t Tones) Len() int {
	return len(t)
}

func (t Tones) Less(i, j int) bool {
	return t.sortKey(i) < t.sortKey(j)
}

func (t Tones) sortKey(i int) string {
	v := t[i]
	return fmt.Sprintf("%02X%02X%02X%02X", v.DrumNote, v.PC, v.BankLSB, v.BankMSB)
}

func (t Tones) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

type SequencerState struct {
	Channels [16]*ChannelState
	Tones    Tones
}

func (ss *SequencerState) AddTone(pc *voice.VM35VoicePC) {
	ss.Tones = append(ss.Tones, pc)
}

func (ss *SequencerState) ToneData() []*voice.VM35FMVoice {
	sort.Sort(ss.Tones)
	tones := []*voice.VM35FMVoice{}
	for _, t := range ss.Tones {
		tones = append(tones, t.Voice.(*voice.VM35FMVoice))
	}
	if 16 < len(tones) {
		log.Warnf("Too many tones (got %d tones, want <=16)", len(tones))
		tones = tones[:16]
	}
	return tones
}

func (ss *SequencerState) GetToneIDByPC(bankMSB, bankLSB, PC int) int {
	for i, pc := range ss.Tones {
		if pc.BankMSB == bankMSB && pc.BankLSB == bankLSB && pc.PC == PC /*&& !pc.IsForDrum()*/ { // @todo uncomment
			return i
		}
	}
	return -1
}

func (ss *SequencerState) GetToneIDByPCAndDrumNote(bankMSB, bankLSB, PC int, note enums.Note) int {
	for i, pc := range ss.Tones {
		if pc.BankMSB == bankMSB && pc.BankLSB == bankLSB && pc.PC == PC && pc.DrumNote == note {
			return i
		}
	}
	return -1
}

func (ss *SequencerState) Tick(fn func(int, []enums.Note)) {
	for ch := 0; ch < 16; ch++ {
		notes := ss.Channels[ch].Tick()
		if 0 < len(notes) {
			fn(ch, notes)
		}
	}
}

func (ss *SequencerState) HasRest() bool {
	for ch := 0; ch < 16; ch++ {
		if ss.Channels[ch].HasRest() {
			return true
		}
	}
	return false
}

func (ss *SequencerState) Print() {
	fmt.Print(cursor.ClearEntireScreen())
	fmt.Print(cursor.MoveTo(0, 0))
	fmt.Println("Ch Bank     PC Note  Bend Mod Vol Exp Pan Mono")
	for i, cs := range ss.Channels {
		cs.Print(i)
	}
}

var State = SequencerState{Channels: [16]*ChannelState{}}

func init() {
	State.Tones = []*voice.VM35VoicePC{}
	for i := 0; i < 16; i++ {
		State.Channels[i] = &ChannelState{
			GateTimeRest:   map[enums.Note]int{},
			ToneID:         0,
			Panpot:         64,
			Volume:         100,
			Expression:     127,
			PitchBendRange: 2,
		}
	}
}
