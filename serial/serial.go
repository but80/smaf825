package serial

import (
	"math"

	"bufio"
	"io"

	"fmt"
	"strings"

	"strconv"
	"time"

	"sync"

	"github.com/but80/smaf825/smaf/enums"
	"github.com/but80/smaf825/smaf/log"
	"github.com/but80/smaf825/smaf/voice"
	"github.com/jacobsa/go-serial/serial"
	"github.com/pkg/errors"
	"github.com/xlab/closer"
)

const (
	SKETCH_VERSION_GTE  = 120
	SKETCH_VERSION_LT   = 140
	ARDUINO_BUFFER_SIZE = 60
)

var BaudRates = []int{300, 600, 1200, 2400, 4800, 9600, 14400, 19200, 28800, 38400, 57600, 115200}

func IsValidBaudRate(r int) bool {
	for _, v := range BaudRates {
		if v == r {
			return true
		}
	}
	return false
}

func BaudRateList() string {
	s := fmt.Sprint(BaudRates)
	return "(" + strings.Replace(s[1:len(s)-1], " ", "|", -1) + ")"
}

type SerialPort struct {
	deviceName    string
	ser           io.ReadWriteCloser
	closed        bool
	selectedCh    int
	sketchVersion int
	commands      []Command
	buffer        []byte
	sentTotal     int
	sendable      int
	bufferMutex   sync.Mutex
}

func NewSerialPort(deviceName string, baudRate int) (*SerialPort, error) {
	log.Infof("opening serial port")
	sp := &SerialPort{
		deviceName: deviceName,
		selectedCh: -1,
		commands:   []Command{},
		buffer:     []byte{},
		sendable:   ARDUINO_BUFFER_SIZE,
	}
	if sp.isNullDevice() {
		sp.closed = true
	} else {
		var err error
		sp.ser, err = serial.Open(serial.OpenOptions{
			PortName:              deviceName,
			BaudRate:              uint(baudRate),
			DataBits:              8,
			StopBits:              1,
			ParityMode:            serial.PARITY_EVEN,
			InterCharacterTimeout: 10000,
			MinimumReadSize:       0,
		})
		if err != nil {
			return nil, errors.WithStack(err)
		}
		closer.Bind(func() {
			sp.Close()
		})
		wait := make(chan error)
		go func() {
			reader := bufio.NewReaderSize(sp.ser, 2048)
			for !sp.closed {
				line, _, err := reader.ReadLine()
				if err == io.EOF {
					if wait != nil {
						wait <- err
					}
					return
				}
				if err != nil {
					log.Warnf("Serial port error: " + err.Error())
				}
				s := string(line)
				if s == "" {
					continue
				}
				if s[0] == '=' {
					readBytes, err := strconv.Atoi(s[1:])
					if err == nil {
						sp.sendable += readBytes
					}
					continue
				}
				log.Debugf("IN: %s", s)
				if wait == nil {
					continue
				}
				if s == "ready" {
					if !(SKETCH_VERSION_GTE <= sp.sketchVersion && sp.sketchVersion < SKETCH_VERSION_LT) {
						wait <- fmt.Errorf(
							`Sketch version mismatch (want %d <= version < %d, got %d). Please rewrite "bridge/bridge.ino" onto Arduino.`,
							SKETCH_VERSION_GTE, SKETCH_VERSION_LT, sp.sketchVersion,
						)
					}
					close(wait)
					wait = nil
				} else if 8 < len(s) && s[:8] == "version " {
					sp.sketchVersion, _ = strconv.Atoi(s[8:])
				}
			}
		}()
		err = <-wait
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return sp, nil
}

func (sp *SerialPort) Close() {
	sp.closed = true
	if sp.ser != nil {
		log.Infof("closing serial port")
		sp.ser.Close()
		log.Infof("done")
	}
	sp.ser = nil
}

func (sp *SerialPort) isNullDevice() bool {
	return sp.deviceName == "/dev/null" || sp.deviceName == "--"
}

func (sp *SerialPort) Flush() bool {
	sp.bufferMutex.Lock()
	defer sp.bufferMutex.Unlock()
	sp.flush()
	return len(sp.buffer) == 0
}

func (sp *SerialPort) flush() {
	if sp.closed {
		return
	}
	if 0 < len(sp.commands) {
		for _, c := range sp.commands {
			sp.buffer = append(sp.buffer, c.Bytes()...)
		}
		sp.commands = []Command{}
	}
	l := len(sp.buffer)
	if sp.sendable < l {
		l = sp.sendable
	}
	if l <= 0 {
		return
	}
	//log.Debugf("sending %d", l)
	n, err := sp.ser.Write(sp.buffer[:l])
	if err != nil {
		panic(errors.WithStack(err))
	}
	sp.buffer = sp.buffer[n:]
	sp.sentTotal += n
	sp.sendable -= n
	//log.Debugf("sent %d sendable=%d", n, sp.sendable)
}

var sendCommandOnce sync.Once

func (sp *SerialPort) sendCommand(c Command) {
	sendCommandOnce.Do(func() {
		ticker := time.NewTicker(8 * time.Millisecond)
		// @todo stop goroutine
		go func() {
			for range ticker.C {
				sp.Flush()
			}
		}()
	})
	sp.bufferMutex.Lock()
	defer sp.bufferMutex.Unlock()
	sp.commands = append(sp.commands, c)
}

func (sp *SerialPort) SendWait(msec int) {
	new := true
	sp.bufferMutex.Lock()
	defer func() {
		sp.bufferMutex.Unlock()
		if new {
			sp.sendCommand(&WaitCommand{Msec: msec})
		}
	}()
	if len(sp.commands) == 0 {
		return
	}
	last, ok := sp.commands[len(sp.commands)-1].(*WaitCommand)
	if !ok {
		return
	}
	last.Msec += msec
	new = false
}

func (sp *SerialPort) SendTerminate() {
	sp.sendCommand(&TerminateCommand{})
}

func (sp *SerialPort) sendData(addr uint8, data []byte) {
	sp.sendCommand(NewSPICommand(addr, data))
}

func (sp *SerialPort) send(addr uint8, data byte) {
	sp.sendCommand(NewSPICommand1(addr, data))
}

func (sp *SerialPort) sendChannelSelect(ch int) {
	// http://madscient.hatenablog.jp/entry/2017/08/13/013913
	// レジスタ#11の下位4ビットに操作したいチャンネル番号を0～15で書き込むことで、
	// レジスタ#12～20に対応するチャンネルのControl Registerが現れます。
	//
	// |I_ADR|W/R|D7 |D6 |D5 |D4 |D3       |D2       |D1       |D0       |Reset Value|
	// |#11  |W/R|"0"|"0"|"0"|"0"|CRGD_VNO3|CRGD_VNO2|CRGD_VNO1|CRGD_VNO0|00H        |
	//
	// CRGD_VNO
	//   The CRGD_VNO is used to specify a tone number.
	//   Reset Conditions
	//     1. When the power supplies are turned on (power-on reset).
	//     2. When the hardware reset is applied (RST_N="L").
	//     3. When the ALRST is set to "1".
	if sp.selectedCh != ch {
		sp.send(11, byte(ch&15))
	}
	sp.selectedCh = ch
}

func (sp *SerialPort) SendAllOff() {
	// https://github.com/yamaha-webmusic/ymf825board/blob/master/manual/fbd_spec2.md#sequencer-setting
	sp.send(8, 0xF6)
	sp.SendWait(1)
	sp.send(8, 0x00)
}

// 0<=v<64
func (sp *SerialPort) SendMasterVolume(v int) {
	sp.send(25, byte(v<<2))
}

// 0<=g<4
func (sp *SerialPort) SendAnalogGain(g int) {
	sp.send(3, byte(g))
}

// 0<=v<32
func (sp *SerialPort) SendSeqVol(v int) {
	sp.send(9, byte(v<<3))
}

func (sp *SerialPort) SendTones(data []*voice.VM35FMVoice) {
	// Contents Format
	// The contents format specifies tone parameters and the sequence of data that can be played back with this device consists of melody contents.
	// The contents are written into the register (I_ADR#7: CONTENTS_DATA_REG) via the CPU interface.
	// Data format
	//   Header: 1byte(80H + Maximum Tone Number)
	//   Tone Setting 30 to 480bytes(it depends on the number of the configured tones)
	//   Sequence Data(any size)
	//   End(80H,03H,81H,80H)
	// Tone Setting
	//   The tone parameters are set by the number of tones set to the Header. The parameter consists of 30 bytes of data for one tone.
	//   The data are transferred and assigned to the Tone parameter memory from Tone 0 in the order they are written;
	//   therefore, parameters of an intermediate Tone number cannot be written first. For details of the tone parameters, see "Tone Parameter"(fbd_spec3.md).

	log.Debugf("sending %d tones", len(data))
	b := []byte{0x80 + byte(len(data))}
	for _, voice := range data {
		b = append(b, voice.Bytes(true, true)...)
	}
	b = append(b, 0x80, 0x03, 0x81, 0x80)
	sp.sendData(7, b)
}

// 0<=vib<8
func (sp *SerialPort) SendVibrato(ch, vib int) {
	// |I_ADR|W/R|D7 |D6    |D5    |D4    |D3    |D2    |D1   |D0    |Reset Value|
	// |#17  |W  |"0"|"0"   |"0"   |"0"   |"0"   |XVB2  |XVB1 |XVB0  |00H        |
	//
	// XVB
	//   The XVB is used to set a vibrato modulation.
	//   This register is provided for each voice.
	//   A setting value relatively acts on a DVB setting value of the voice parameter, as shown below.
	//   When the calculation (add) result exceeds "3", "3"is used for the processing.
	//     "0": OFF (reset value)
	//     "1": 1 x (DVB value is used as is.)
	//     "2": 2 x (DVB += 1)
	//     "3": 2 x (DVB += 1)
	//     "4": 4 x (DVB += 2)
	//     "5": 4 x (DVB += 2)
	//     "6": 8 x (DVB += 3)
	//     "7": 8 x (DVB += 3)
	if ch < 0 {
		return
	}
	sp.sendChannelSelect(ch)
	sp.send(17, byte(vib&7))
}

// 0<=ChVol<32
func (sp *SerialPort) SendVolume(ch, ChVol int, DIR_CV bool) {
	// |I_ADR|W/R|D7 |D6    |D5    |D4    |D3    |D2    |D1 |D0    |Reset Value|
	// |#16  |W  |"0"|ChVol4|ChVol3|ChVol2|ChVol1|ChVol0|"0"|DIR_CV|60H        |
	//
	// ChVol
	//   This volume setting register is provided for each voice.
	//   The interpolation function is provided for this volume setting register.
	//   The relationship between setting values and volume gain values is the same as that of VoVol and SEQ_Vol. Reset Value is "18H" (-4.4 dB)
	//
	// DIR_CV
	//   The DIR_CV controls the interpolation of the SEQ_Vol and ChVol.
	//   This register is provided for each voice.
	//   DIR_CV="1":
	//     No interpolation in the SEQ_Vol and the ChVol# regardless of the DIR_SV and CHVOL_ITIME settings.
	//   DIR_CV# = "0" (reset value):
	//     The interpolation depends on the DIR_SV and CHVOL_ITIME settings.
	if ch < 0 {
		return
	}
	b := 0
	if DIR_CV {
		b = 1
	}
	sp.sendChannelSelect(ch)
	sp.send(16, byte(ChVol&31<<2|b))
}

// 0<=INT<4 0<=FRAC<512
func (sp *SerialPort) SendFineTune(ch, INT, FRAC int) {
	// |I_ADR|W/R|D7 |D6   |D5   |D4   |D3   |D2   |D1   |D0   |Reset Value|
	// |#18  |W  |"0"|"0"  |"0"  |INT1 |INT0 |FRAC8|FRAC7|FRAC6|08H        |
	// |#19  |W  |"0"|FRAC5|FRAC4|FRAC3|FRAC2|FRAC1|FRAC0|"0"  |00H        |
	//
	// INT, FRAC
	//   These registers specify a multiplier to the generated audio frequency. This number and frequency are proportional.
	//   The INT is an integer part and FRAC is a fraction part.
	//   These registers are provided for each voice.
	//   Reset Value
	//     INT : "01H"
	//     FRAC: "000H"
	if ch < 0 {
		return
	}
	sp.sendChannelSelect(ch)
	sp.send(18, byte(INT&3<<3|FRAC>>6&7))
	sp.send(19, byte(FRAC&63<<1))
}

func (sp *SerialPort) SendFineTuneByFloat(ch int, r float64) {
	if ch < 0 {
		return
	}
	r += .5 / 512.0
	INT := int(math.Floor(r))
	FRAC := int(math.Floor((r - float64(INT)) * 512))
	sp.SendFineTune(ch, INT, FRAC)
}

func (sp *SerialPort) SendKeyOn(ch int, note enums.Note, delta float64, VoVol, ToneNum int) {
	// https://github.com/yamaha-webmusic/ymf825board/blob/master/manual/fbd_spec2.md#control-register-write-registers
	//
	// |I_ADR |W/R|D7 |D6    |D5    |D4    |D3       |D2       |D1       |D0       |Reset Value|
	// |#12   |W  |"0"|VoVol4|VoVol3|VoVol2|VoVol1   |VoVol0   |"0"      |"0"      |60H        |
	// |#13   |W  |"0"|"0"   |FNUM9 |FNUM8 |FNUM7    |BLOCK2   |BLOCK1   |BLOCK0   |00H        |
	// |#14   |W  |"0"|FNUM6 |FNUM5 |FNUM4 |FNUM3    |FNUM2    |FNUM1    |FNUM0    |00H        |
	// |#15   |W  |"0"|KeyOn |Mute  |EG_RST|ToneNum3 |ToneNum2 |ToneNum1 |ToneNum0 |00H        |
	if ch < 0 {
		return
	}
	f := note.Freq(delta)
	sp.sendChannelSelect(ch)
	sp.send(12, byte(VoVol&31<<2))
	sp.send(13, byte((f.Fnum>>7&7)<<3|f.Block&7))
	sp.send(14, byte(f.Fnum&127))
	sp.send(15, 0x40|byte(ToneNum&15))
}

func (sp *SerialPort) SendPitch(ch int, note enums.Note, delta float64) {
	if ch < 0 {
		return
	}
	f := note.Freq(delta)
	sp.sendChannelSelect(ch)
	sp.send(13, byte((f.Fnum>>7&7)<<3|f.Block&7))
	sp.send(14, byte(f.Fnum&127))
}

func (sp *SerialPort) SendKeyOff(ch, ToneNum int) {
	if ch < 0 {
		return
	}
	sp.sendChannelSelect(ch)
	sp.send(15, byte(ToneNum&15))
}

func (sp *SerialPort) SendMuteAndEGReset(ch int) {
	if ch < 0 {
		return
	}
	sp.sendChannelSelect(ch)
	sp.send(15, 0x30)
}
