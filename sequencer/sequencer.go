package sequencer

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/but80/smaf825/serial"
	"github.com/pkg/errors"
	"github.com/xlab/closer"
	"gopkg.in/but80/go-smaf.v1/chunk"
	"gopkg.in/but80/go-smaf.v1/enums"
	"gopkg.in/but80/go-smaf.v1/event"
	"gopkg.in/but80/go-smaf.v1/log"
	"gopkg.in/but80/go-smaf.v1/util"
	"gopkg.in/but80/go-smaf.v1/voice"
)

/*

ダミーのシリアルデバイスを作ってテスト

$ socat -d -d pty,raw pty,raw &
socat[31853] N PTY is /dev/ttys006  <- serial.Config の Name に指定
socat[31853] N PTY is /dev/ttys007  <- これを cat して確認
socat[31853] N starting data transfer loop with FDs [5,5] and [7,7]

$ hexdump -C < /dev/ttys007

*/

// Options は、シーケンサのオプション設定です。
type Options struct {
	Loop     int
	Volume   int
	Gain     int
	SeqVol   int
	BaudRate int
}

// Sequencer は、シーケンサです。
type Sequencer struct {
	// DeviceName は、出力先シリアルポートのデバイス名です。
	DeviceName string
	// ShowState は、trueのときシーケンサの状態を画面表示します。
	ShowState bool

	controller *serial.Controller
}

var debugFlags = struct {
	Tone       bool
	Volume     bool
	Octave     bool
	KeyControl bool
}{
	Tone:       false,
	Volume:     false,
	Octave:     false,
	KeyControl: false,
}

// Play は、シーケンサの演奏を開始します。
func (q *Sequencer) Play(mmf *chunk.FileChunk, opts *Options) error {
	var err error
	var info *chunk.ContentsInfoChunk
	var data *chunk.DataChunk
	var setup chunk.ExclusiveContainer
	var score *chunk.ScoreTrackChunk
	sequences := []*chunk.ScoreTrackSequenceDataChunk{}
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
		case chunk.ExclusiveContainer:
			setup = ck
		case *chunk.ScoreTrackChunk:
			score = ck
		case *chunk.ScoreTrackSequenceDataChunk:
			sequences = append(sequences, ck)
		}
	})
	switch setup.(type) {
	case nil:
		return fmt.Errorf("Score track setup chunk not found")
	}
	if len(sequences) == 0 {
		return fmt.Errorf("Sequence data chunk not found")
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
		log.Infof("")
		log.Infof("=============== playing %s", strings.Join(contentsInfo, " - "))
		log.Infof("")
	}
	//
	channelsToSplit := []enums.Channel{}
	if score != nil {
		allOff := true
		for _, st := range score.ChannelStatus {
			if st.KeyControlStatus != enums.KeyControlStatus_Off {
				allOff = false
				break
			}
		}
		for ch, st := range score.ChannelStatus {
			if allOff || debugFlags.KeyControl {
				st.KeyControlStatus = enums.KeyControlStatus_On
			}
			State.Channels[ch].KeyControlStatus = st.KeyControlStatus
			if st.KeyControlStatus == enums.KeyControlStatus_Off {
				channelsToSplit = append(channelsToSplit, ch)
			}
		}
	}
	sequence := chunk.MergeSequenceDataChunks(sequences)
	sequence.AggregateUsage(channelsToSplit)
	//
	log.Debugf("collecting voices")
	for _, x := range setup.GetExclusives() {
		switch x.Type {
		case enums.ExclusiveType_VM35Voice:
			v := x.VM35VoicePC
			if v != nil && v.VoiceType == enums.VoiceType_FM && !sequence.IsIgnoredPC(v.BankMSB, v.BankLSB, v.PC, v.DrumNote) {
				State.AddTone(v)
			}
		case enums.ExclusiveType_VMAVoice:
			v := x.VMAVoicePC
			if v != nil && !sequence.IsIgnoredPC(0, v.Bank, v.PC, 0) {
				State.AddTone(v.ToVM35())
			}
		}
	}
	//
	if q.controller == nil {
		q.controller, err = serial.NewSerialPort(q.DeviceName, opts.BaudRate)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	q.controller.SendMasterVolume(opts.Volume)
	q.controller.SendAnalogGain(opts.Gain)
	q.controller.SendSeqVol(opts.SeqVol)
	//
	log.Debugf("sending voices")
	q.controller.SendAllOff() // トーン設定時は発音をすべて停止
	if debugFlags.Tone {
		q.controller.SendTones([]*voice.VM35FMVoice{
			voice.NewDemoVM35FMVoice(),
		})
	} else {
		q.controller.SendTones(State.ToneData())
	}
	//
	var timeBase, durationTickCycle, gateTickCycle int
	if score == nil {
		timeBase = 20
		durationTickCycle = 1
		gateTickCycle = 1
	} else {
		timeBase = util.GCD(score.DurationTimeBase, score.GateTimeBase)
		durationTickCycle = score.DurationTimeBase / timeBase
		gateTickCycle = score.GateTimeBase / timeBase
	}
	log.Debugf("common time base = %d msec", timeBase)
	log.Debugf("durationTickCycle = %d", durationTickCycle)
	log.Debugf("gateTickCycle = %d", gateTickCycle)
	ticker := time.NewTicker(time.Duration(timeBase)*time.Millisecond - time.Millisecond)
	end := make(chan bool)
	stopped := false
	closer.Bind(func() {
		stopped = true
		q.controller.SendAllOff()
	})
	go func() {
		loop := opts.Loop
		iEvent := 0
		durationRest := 0
		q.controller.SendWait(1000)
		var pendingEvent event.Event
		for !stopped && (iEvent < len(sequence.Events) || State.HasRest()) {
			q.controller.SendWait(timeBase)
			select {
			case <-ticker.C:
				keyOffFound := false
				State.Tick(func(ch int, notes []enums.Note) {
					cs := State.Channels[ch]
					for _, note := range notes {
						chTo := sequence.ChannelTo(enums.Channel(ch), note)
						toneID := cs.ToneID
						if cs.KeyControlStatus == enums.KeyControlStatus_Off {
							toneID = State.GetToneIDByPCAndDrumNote(cs.BankMSB, cs.BankLSB, cs.PC, note)
						}
						q.controller.SendKeyOff(chTo, toneID)
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
						//	log.Debugf("dur %d", pair.Duration)
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
	q.controller.SendAllOff()
	for !q.controller.Flush() {
		time.Sleep(time.Millisecond)
	}
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
		vel := float64(cs.Velocity) / 127.0
		exp := float64(cs.Expression) / 127.0
		var vol float64
		if State.IsMA5 {
			vol = vel + exp - 1.0
			if vol < .0 {
				vol = .0
			}
		} else {
			vol = (vel + exp) * .5
			if vel == .0 || exp == .0 {
				vol = .0
			}
		}
		if debugFlags.Volume {
			vol = 1.0
		}
		delta := float64(cs.PitchBend) * float64(cs.PitchBendRange) / 8192.0
		note := evt.Note
		toneID := cs.ToneID
		chTo := sequence.ChannelTo(ch, note)
		if cs.KeyControlStatus == enums.KeyControlStatus_Off {
			toneID = State.GetToneIDByPCAndDrumNote(cs.BankMSB, cs.BankLSB, cs.PC, note)
			if 0 <= toneID {
				note = State.Tones[toneID].Voice.(*voice.VM35FMVoice).DrumKey
			}
		}
		if debugFlags.Tone {
			toneID = 0
		}
		if 0 <= toneID {
			q.controller.SendKeyOn(chTo, note+enums.Note(cs.OctaveShift*12), delta, int(math.Floor(.5+31.0*vol)), toneID)
		}

	case *event.PitchBendEvent:
		cs.PitchBend = evt.Value
		delta := float64(cs.PitchBend) * float64(cs.PitchBendRange) / 8192.0
		for note := range cs.GateTimeRest {
			chTo := sequence.ChannelTo(ch, note)
			q.controller.SendPitch(chTo, note, delta)
		}

	case *event.ControlChangeEvent:
		q.sendCC(sequence, evt)

	case *event.ProgramChangeEvent:
		cs.PC = evt.PC
		toneID := State.GetToneIDByPC(cs.BankMSB, cs.BankLSB, cs.PC)
		if 0 <= toneID {
			cs.ToneID = toneID
		} else {
			log.Warnf("Undefined or unsupported PC %d-%d-@%d", cs.BankMSB, cs.BankLSB, cs.PC)
		}

	case *event.OctaveShiftEvent:
		if debugFlags.Octave {
			cs.OctaveShift = 0
		} else {
			cs.OctaveShift = evt.Value
		}

	case *event.ExclusiveEvent:
		// @todo process ExclusiveEvent

	case *event.NopEvent:
		// nop
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
			q.controller.SendVibrato(chTo, scale127(evt.Value, 7, 1.0))
		}
	case enums.CC_MainVolume:
		vol := evt.Value
		if debugFlags.Volume {
			vol = 127
		}
		cs.Volume = vol
		for _, chTo := range chsTo {
			q.controller.SendVolume(chTo, scale127(vol, 31, 1.0), true)
		}
	case enums.CC_Panpot:
		cs.Panpot = evt.Value
	case enums.CC_Expression:
		cs.Expression = evt.Value
	case enums.CC_BankSelectLSB:
		cs.BankLSB = evt.Value
	case enums.CC_MonoOn:
		cs.Mono = true
	case enums.CC_PolyOn:
		cs.Mono = false
	case enums.CC_RPNLSB:
		cs.RPNLSB = evt.Value
	case enums.CC_RPNMSB:
		cs.RPNMSB = evt.Value
	case enums.CC_AllSoundOff:
		notes := cs.AllOff()
		for _, note := range notes {
			chTo := sequence.ChannelTo(enums.Channel(ch), note)
			toneID := cs.ToneID
			if cs.KeyControlStatus == enums.KeyControlStatus_Off {
				toneID = State.GetToneIDByPCAndDrumNote(cs.BankMSB, cs.BankLSB, cs.PC, note)
			}
			q.controller.SendKeyOff(chTo, toneID)
		}
	case enums.CC_DataEntry:
		switch cs.RPNMSB {
		case 0:
			switch cs.RPNLSB {
			case 0: // Pitch bend sensitivity
				cs.PitchBendRange = evt.Value
				delta := float64(cs.PitchBend) * float64(cs.PitchBendRange) / 8192.0
				for note := range cs.GateTimeRest {
					chTo := sequence.ChannelTo(ch, note)
					q.controller.SendPitch(chTo, note, delta)
				}
			//case 1: // Master fine tuning
			//case 2: // Master coarse tuning
			default:
				log.Warnf("Unsupported RPN %d-%d = %d", cs.RPNMSB, cs.RPNLSB, evt.Value)
			}
		default:
			log.Warnf("Unsupported RPN %d-%d = %d", cs.RPNMSB, cs.RPNLSB, evt.Value)
		}
	default:
		log.Warnf("Unsupported CC %s = %d", evt.CC.String(), evt.Value)
	}
}

func debugSendTestTone(deviceName string) error {
	sp, err := serial.NewSerialPort(deviceName, 76800)
	if err != nil {
		return errors.WithStack(err)
	}
	defer sp.Close()

	// 初期化処理について
	// http://madscient.hatenablog.jp/entry/2017/08/13/013913
	// https://github.com/yamaha-webmusic/ymf825board/blob/master/manual/fbd_spec1.md#initialization-procedure

	v, _ := voice.NewVM35FMVoice([]byte{
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
	}, voice.VM35FMVoiceVersion_VM5)

	sp.SendAllOff() // トーン設定時は発音をすべて停止
	sp.SendTones([]*voice.VM35FMVoice{v})

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
