package serial

// Command は、シリアルポート経由でArduinoに送られるコマンドが備えるインタフェースです。
type Command interface {
	// Bytes は、シリアルポートに送信するバイト列を生成します。
	Bytes() []byte
}

// SPICommand は、SPI通信を行うコマンドです。
type SPICommand struct {
	Addr uint8
	Data []byte
}

// NewSPICommand は、新しい SPICommand を作成します。
func NewSPICommand(addr uint8, data []byte) *SPICommand {
	return &SPICommand{
		Addr: addr,
		Data: data,
	}
}

// NewSPICommand1 は、新しい SPICommand を作成します。
func NewSPICommand1(addr uint8, data byte) *SPICommand {
	return &SPICommand{
		Addr: addr,
		Data: []byte{data},
	}
}

// Bytes は、シリアルポートに送信するバイト列を生成します。
func (c *SPICommand) Bytes() []byte {
	n := len(c.Data)
	var hdr []byte
	if 1 == n {
		hdr = []byte{c.Addr}
	} else {
		hdr = []byte{c.Addr | 0x80, byte(n >> 8 & 255), byte(n & 255)}
	}
	return append(hdr, c.Data...)
}

// WaitCommand は、一定時間待機させるコマンドです。
type WaitCommand struct {
	// Msec は、待機時間 [ミリ秒] です。
	Msec int
}

// Bytes は、シリアルポートに送信するバイト列を生成します。
func (c *WaitCommand) Bytes() []byte {
	return []byte{0xFF, byte(c.Msec >> 8 & 255), byte(c.Msec & 255)}
}

// TerminateCommand は、シリアル通信を終了するコマンドです。
type TerminateCommand struct {
}

// Bytes は、シリアルポートに送信するバイト列を生成します。
func (c *TerminateCommand) Bytes() []byte {
	return []byte{0xFF, 0xFF, 0xFF}
}
