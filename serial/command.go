package serial

type Command interface {
	Bytes() []byte
}

type SPICommand struct {
	Addr uint8
	Data []byte
}

func NewSPICommand(addr uint8, data []byte) *SPICommand {
	return &SPICommand{
		Addr: addr,
		Data: data,
	}
}

func NewSPICommand1(addr uint8, data byte) *SPICommand {
	return &SPICommand{
		Addr: addr,
		Data: []byte{data},
	}
}

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

type WaitCommand struct {
	Msec int
}

func (c *WaitCommand) Bytes() []byte {
	return []byte{0xFF, byte(c.Msec >> 8 & 255), byte(c.Msec & 255)}
}

type TerminateCommand struct {
}

func (c *TerminateCommand) Bytes() []byte {
	return []byte{0xFF, 0xFF, 0xFF}
}
