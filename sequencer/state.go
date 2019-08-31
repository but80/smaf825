package sequencer

import (
	"fmt"
	"sort"

	"github.com/ahmetalpbalkan/go-cursor"
	"github.com/but80/go-smaf/v2/enums"
	"github.com/but80/go-smaf/v2/log"
	"github.com/but80/go-smaf/v2/voice"
)

// ChannelState は、チャンネルの状態です。
type ChannelState struct {
	// KeyControlStatus は、現在のキーのコントロール状態です。
	KeyControlStatus enums.KeyControlStatus
	// Velocity は、現在のベロシティです。
	Velocity int
	// GateTimeRest は、各ノートの現在のゲートタイム残時間 [tick] です。
	GateTimeRest map[enums.Note]int
	// BankMSB は、現在のバンクMSBです。
	BankMSB int
	// BankLSB は、現在のバンクLSBです。
	BankLSB int
	// PC は、現在のプログラムチェンジです。
	PC int
	// ToneID は、現在の音色IDです。
	ToneID int
	// PitchBend は、現在のピッチベンド値です。
	PitchBend int
	// PitchBendRange は、現在のピッチベンド幅です。
	PitchBendRange int
	// Modulation は、現在のモジュレーション量です。
	Modulation int
	// Volume は、現在のボリュームです。
	Volume int
	// Panpot は、現在のパンポット値です。
	Panpot int
	// Expression は、現在のエクスプレッション値です。
	Expression int
	// OctaveShift は、現在のオクターブシフト値です。
	OctaveShift int
	// Mono は、現在モノモードのときtrueとします。
	Mono bool
	// RPNMSB は、現在のRPN MSBです。
	RPNMSB int
	// RPNLSB は、現在のRPN LSBです。
	RPNLSB int
}

// Tick は、現在の状態を 1 tick 進め、ゲートタイム残時間を更新します。
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

// AllOff は、全ノートのゲートタイム残時間を 0 にリセットします。
// 現時点で発音中だったノートの一覧を返します。
func (cs *ChannelState) AllOff() []enums.Note {
	notes := []enums.Note{}
	for note := range cs.GateTimeRest {
		notes = append(notes, note)
	}
	cs.GateTimeRest = map[enums.Note]int{}
	return notes
}

// HasRest は、まだ発音中のノートがあるときtrueを返します。
func (cs *ChannelState) HasRest() bool {
	return 0 < len(cs.GateTimeRest)
}

// NoteOn は、指定のノートがキーオンされたものとしてゲートタイム残時間を更新します。
func (cs *ChannelState) NoteOn(note enums.Note, gateTime int) {
	if cs.KeyControlStatus != enums.KeyControlStatusOff {
		cs.GateTimeRest = map[enums.Note]int{}
	}
	cs.GateTimeRest[note] = gateTime
}

// Print は、現在の状態を表示します。
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
		"%2d %3d-%-3d %3d %-8s %5d %3d %3d %3d %-3s %-4s\n",
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

// Tones は、音色データです。
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

// StateStatic は、シーケンサの状態です。
type StateStatic struct {
	// Channels は、各チャンネルの状態です。
	Channels [16]*ChannelState
	// Tones は、音色データです。
	Tones Tones
	// IsMA5 は、MA-5用のシーケンスを再生中はtrueとします。
	IsMA5 bool
}

// AddTone は、音色データを追加します。
func (ss *StateStatic) AddTone(pc *voice.VM35VoicePC) {
	ss.Tones = append(ss.Tones, pc)
	if pc.Version == voice.VM35FMVoiceVersionVM5 {
		ss.IsMA5 = true
	}
}

// ToneData は、音色データ中のFM音色定義部を一覧で返します。
func (ss *StateStatic) ToneData() []*voice.VM35FMVoice {
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

// GetToneIDByPC は、プログラムチェンジ番号から音色データIDを取得します。
func (ss *StateStatic) GetToneIDByPC(bankMSB, bankLSB, PC int) int {
	for i, pc := range ss.Tones {
		if pc.BankMSB == bankMSB && pc.BankLSB == bankLSB && pc.PC == PC /*&& !pc.IsForDrum()*/ { // @todo uncomment
			return i
		}
	}
	return -1
}

// GetToneIDByPCAndDrumNote は、プログラムチェンジ番号とノートナンバーから音色データIDを取得します。
func (ss *StateStatic) GetToneIDByPCAndDrumNote(bankMSB, bankLSB, PC int, note enums.Note) int {
	for i, pc := range ss.Tones {
		if pc.BankMSB == bankMSB && pc.BankLSB == bankLSB && pc.PC == PC && pc.DrumNote == note {
			return i
		}
	}
	return -1
}

// Tick は、現在の状態を 1 tick 進め、全チャンネルのゲートタイム残時間を更新します。
func (ss *StateStatic) Tick(fn func(int, []enums.Note)) {
	for ch := 0; ch < 16; ch++ {
		notes := ss.Channels[ch].Tick()
		if 0 < len(notes) {
			fn(ch, notes)
		}
	}
}

// HasRest は、いずれかのチャンネルにまだ発音中のノートがあるときtrueを返します。
func (ss *StateStatic) HasRest() bool {
	for ch := 0; ch < 16; ch++ {
		if ss.Channels[ch].HasRest() {
			return true
		}
	}
	return false
}

// Print は、現在の状態を表示します。
func (ss *StateStatic) Print() {
	fmt.Print(cursor.ClearEntireScreen())
	fmt.Print(cursor.MoveTo(0, 0))
	fmt.Println("Ch Bank     PC Note      Bend Mod Vol Exp Pan Mono")
	for i, cs := range ss.Channels {
		cs.Print(i)
	}
}

// State は、シーケンサの状態を保持する唯一のインスタンスです。
var State = StateStatic{Channels: [16]*ChannelState{}}

func init() {
	State.Tones = []*voice.VM35VoicePC{}
	for i := 0; i < 16; i++ {
		State.Channels[i] = &ChannelState{
			KeyControlStatus: enums.KeyControlStatusOn,
			GateTimeRest:     map[enums.Note]int{},
			ToneID:           0,
			Panpot:           64,
			Volume:           100,
			Expression:       127,
			PitchBendRange:   2,
		}
	}
}
