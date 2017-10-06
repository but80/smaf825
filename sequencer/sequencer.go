package sequencer

import (
	"time"

	"fmt"

	"math"

	"strings"

	"github.com/mersenne-sister/smaf825/serial"
	"github.com/mersenne-sister/smaf825/smaf/chunk"
	"github.com/mersenne-sister/smaf825/smaf/enums"
	"github.com/mersenne-sister/smaf825/smaf/event"
	"github.com/mersenne-sister/smaf825/smaf/util"
	"github.com/mersenne-sister/smaf825/smaf/voice"
	"github.com/pkg/errors"
	"github.com/xlab/closer"
)

/*

ダミーのシリアルデバイスを作ってテスト

$ socat -d -d pty,raw pty,raw &
socat[31853] N PTY is /dev/ttys006  <- serial.Config の Name に指定
socat[31853] N PTY is /dev/ttys007  <- これを cat して確認
socat[31853] N starting data transfer loop with FDs [5,5] and [7,7]

$ hexdump -C < /dev/ttys007

*/

type Sequencer struct {
	DeviceName string
	ShowState  bool
	port       *serial.SerialPort
}

func (q *Sequencer) Play(mmf *chunk.FileChunk, loop, volume, gain, seqvol int) error {
	var err error
	n := 0
	var info *chunk.ContentsInfoChunk
	var data *chunk.DataChunk
	var setup *chunk.ScoreTrackSetupDataChunk
	var score *chunk.ScoreTrackChunk
	var sequence *chunk.ScoreTrackSequenceDataChunk
	mmf.Traverse(func(c chunk.Chunk) {
		switch ck := c.(type) {
		case *chunk.ContentsInfoChunk:
			if ck.HasOptions {
				info = ck
			}
		case *chunk.DataChunk:
			if ck.HasOptions {
				data = ck
			}
		case *chunk.ScoreTrackSetupDataChunk:
			n++
			setup = ck
		case *chunk.ScoreTrackChunk:
			n++
			score = ck
		case *chunk.ScoreTrackSequenceDataChunk:
			n++
			sequence = ck
		}
	})
	if setup == nil {
		return fmt.Errorf("Score track setup chunk not found")
	}
	if score == nil {
		return fmt.Errorf("Score track chunk not found")
	}
	if sequence == nil {
		return fmt.Errorf("Sequence data chunk not found")
	}
	if 3 < n {
		return fmt.Errorf("Too many score/sequence chunks")
	}
	//
	contentsInfo := []string{}
	if info != nil {
		if info.Options.Artist != "" {
			contentsInfo = append(contentsInfo, info.Options.Artist)
		}
		if info.Options.Title != "" {
			contentsInfo = append(contentsInfo, info.Options.Title)
		}
	}
	if data != nil {
		if data.Options.Artist != "" {
			contentsInfo = append(contentsInfo, data.Options.Artist)
		}
		if data.Options.Title != "" {
			contentsInfo = append(contentsInfo, data.Options.Title)
		}
	}
	if 0 < len(contentsInfo) {
		fmt.Printf("\n=============== playing %s\n\n", strings.Join(contentsInfo, " - "))
	}
	//
	channelsToSplit := []int{}
	for ch, st := range score.ChannelStatus {
		State.Channels[ch].KeyControlStatus = st.KeyControlStatus
		if st.KeyControlStatus == enums.KeyControlStatus_Off {
			channelsToSplit = append(channelsToSplit, ch)
		}
	}
	sequence.AggregateUsage(channelsToSplit)
	//
	//fmt.Println("collecting voices")
	for _, x := range setup.Exclusives {
		switch x.Type {
		case enums.ExclusiveType_VM5Voice:
			v := x.VM5VoicePC
			if v != nil && !sequence.IsIgnoredPC(v.BankMSB, v.BankLSB, v.PC, v.DrumNote) {
				State.AddTone(v)
			}
		case enums.ExclusiveType_VMAVoice:
			v := x.VMAVoicePC
			if v != nil {
				State.AddTone(v.ToVM5())
			}
		}
	}
	//
	if q.port == nil {
		q.port, err = serial.NewSerialPort(q.DeviceName)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	q.port.SendMasterVolume(volume)
	q.port.SendAnalogGain(gain)
	q.port.SendSeqVol(seqvol)
	//
	fmt.Println("sending voices")
	q.port.SendAllOff() // トーン設定時は発音をすべて停止
	q.port.SendTones(State.ToneData())
	//
	timeBase := util.GCD(score.DurationTimeBase, score.GateTimeBase)
	durationTickCycle := score.DurationTimeBase / timeBase
	gateTickCycle := score.GateTimeBase / timeBase
	fmt.Printf("common time base = %d msec\n", timeBase)
	fmt.Printf("durationTickCycle = %d\n", durationTickCycle)
	fmt.Printf("gateTickCycle = %d\n", gateTickCycle)
	ticker := time.NewTicker(time.Duration(timeBase) * time.Millisecond)
	end := make(chan bool)
	stopped := false
	closer.Bind(func() {
		stopped = true
	})
	go func() {
		iEvent := 0
		durationRest := 0
		var pendingEvent event.Event
		for !stopped && (iEvent < len(sequence.Events) || State.HasRest()) {
			select {
			case <-ticker.C:
				keyOffFound := false
				State.Tick(func(ch int, notes []enums.Note) {
					cs := State.Channels[ch]
					for _, note := range notes {
						chTo := sequence.ChannelTo(ch, note)
						toneID := cs.ToneID
						if cs.KeyControlStatus == enums.KeyControlStatus_Off {
							toneID = State.GetToneIDByPCAndDrumNote(cs.BankMSB, cs.BankLSB, cs.PC, note)
						}
						q.port.SendKeyOff(chTo, toneID)
					}
					keyOffFound = true
				})
				durationRest--
				if 0 < durationRest {
					if keyOffFound && q.ShowState {
						State.Print()
					}
					continue
				}
				if pendingEvent != nil {
					q.processEvent(sequence, gateTickCycle, pendingEvent)
					pendingEvent = nil
				}
				for iEvent < len(sequence.Events) {
					pair := sequence.Events[iEvent]
					iEvent++
					if len(sequence.Events) <= iEvent && loop != 1 {
						loop--
						iEvent = 0
					}
					if 0 < pair.Duration {
						if q.ShowState {
							State.Print()
						}
						//if 128 <= pair.Duration {
						//	fmt.Printf("dur %d\n", pair.Duration)
						//}
						durationRest = pair.Duration * durationTickCycle
						pendingEvent = pair.Event
						break
					}
					q.processEvent(sequence, gateTickCycle, pair.Event)
				}
			}
		}
		end <- true
	}()
	<-end
	ticker.Stop()
	q.port.SendAllOff()
	return nil
}

func scale127(v, max int, curve float64) int {
	r := float64(v) / 127.0
	r = math.Pow(r, curve)
	return int(math.Floor(.5 + float64(max)*r))
}

func (q *Sequencer) processEvent(sequence *chunk.ScoreTrackSequenceDataChunk, gateTickCycle int, e event.Event) {
	ch := e.GetChannel()
	cs := State.Channels[ch]
	switch evt := e.(type) {

	case *event.NoteEvent:
		cs.Velocity = evt.Velocity
		cs.NoteOn(evt.Note, evt.GateTime*gateTickCycle) // @todo Add "+1" for tie/slur only
		vol := float64(cs.Velocity) / 127.0 * float64(cs.Expression) / 127.0
		delta := float64(cs.PitchBend) / 4096.0 // @todo consider bend range
		note := evt.Note
		toneID := cs.ToneID
		// @todo Fix note and select tone ID for tracks with KeyControlStatus_Off
		chTo := sequence.ChannelTo(ch, note)
		if cs.KeyControlStatus == enums.KeyControlStatus_Off {
			toneID = State.GetToneIDByPCAndDrumNote(cs.BankMSB, cs.BankLSB, cs.PC, note)
			if 0 <= toneID {
				note = State.Tones[toneID].Voice.(*voice.VM5FMVoice).DrumKey
			}
		}
		if 0 <= toneID {
			q.port.SendKeyOn(chTo, note+enums.Note(cs.OctaveShift*12), delta, int(math.Floor(.5+31.0*vol)), toneID)
		}

	case *event.PitchBendEvent:
		cs.PitchBend = evt.Value
		delta := float64(cs.PitchBend) / 4096.0 // @todo consider bend range
		for note := range cs.GateTimeRest {
			chTo := sequence.ChannelTo(ch, note)
			q.port.SendPitch(chTo, note, delta)
		}

	case *event.ControlChangeEvent:
		q.sendCC(sequence, evt)

	case *event.ProgramChangeEvent:
		cs.PC = evt.PC
		toneID := State.GetToneIDByPC(cs.BankMSB, cs.BankLSB, cs.PC)
		if 0 <= toneID {
			cs.ToneID = toneID
		}

	case *event.OctaveShiftEvent:
		cs.OctaveShift = evt.Value

	case *event.ExclusiveEvent:
		// @todo process ExclusiveEvent

	case *event.NopEvent:
		// nop
		//fmt.Println("nop")
	default:
	}
}

func (q *Sequencer) sendCC(sequence *chunk.ScoreTrackSequenceDataChunk, evt *event.ControlChangeEvent) {
	ch := evt.GetChannel()
	cs := State.Channels[ch]
	chsTo := sequence.ChannelsTo(ch)
	switch evt.CC {
	case enums.CC_BankSelectMSB:
		cs.BankMSB = evt.Value
	case enums.CC_Modulation:
		cs.Modulation = evt.Value
		for _, chTo := range chsTo {
			q.port.SendVibrato(chTo, scale127(evt.Value, 7, 1.0))
		}
	case enums.CC_MainVolume:
		cs.Volume = evt.Value
		for _, chTo := range chsTo {
			q.port.SendVolume(chTo, scale127(evt.Value, 31, 1.0), true)
		}
	case enums.CC_Panpot:
		cs.Panpot = evt.Value
	case enums.CC_Expression:
		cs.Expression = evt.Value
	case enums.CC_BankSelectLSB:
		cs.BankLSB = evt.Value
	case enums.CC_MonoOn:
		cs.Mono = true
	}
}

func Test(deviceName string) error {
	sp, err := serial.NewSerialPort(deviceName)
	if err != nil {
		return errors.WithStack(err)
	}
	defer sp.Close()

	// 初期化処理について
	// http://madscient.hatenablog.jp/entry/2017/08/13/013913
	// https://github.com/yamaha-webmusic/ymf825board/blob/master/manual/fbd_spec1.md#initialization-procedure

	v, _ := voice.NewVM5FMVoice([]byte{
		//0x01, 0x85,
		//0x00, 0x7F, 0xF4, 0xBB, 0x00, 0x10, 0x40,
		//0x00, 0xAF, 0xA0, 0x0E, 0x03, 0x10, 0x40,
		//0x00, 0x2F, 0xF3, 0x9B, 0x00, 0x20, 0x41,
		//0x00, 0xAF, 0xA0, 0x0E, 0x01, 0x10, 0x40,

		// Slap bass
		0x00, 0x43,
		0x23, 0x37, 0xF2, 0x3A, 0x44, 0x10, 0x03,
		0x63, 0x66, 0xF4, 0x54, 0x44, 0x90, 0x00,
		0x23, 0x69, 0xC2, 0x62, 0x44, 0x10, 0x00,
		0xF3, 0x82, 0xFF, 0x0C, 0x44, 0x10, 0x00,
	})

	sp.SendAllOff() // トーン設定時は発音をすべて停止
	sp.SendTones([]*voice.VM5FMVoice{v})

	ch := 0

	sp.SendMuteAndEGReset(ch)
	sp.SendVolume(ch, 28, true)
	sp.SendVibrato(ch, 0)
	sp.SendFineTune(ch, 1, 0)

	for o := 1; o < 4; o++ {
		for _, i := range []int{0, 2, 4, 5, 7, 9, 11} {
			n := enums.Note(o*12 + i)
			sp.SendKeyOn(ch, n, .0, 21, 0)
			time.Sleep(200 * time.Millisecond)
			sp.SendKeyOff(ch, 0)
			time.Sleep(100 * time.Millisecond)
		}
	}

	sp.SendTerminate()

	return nil
}
