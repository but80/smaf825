package enums

import "fmt"

type CC int

const (
	CC_BankSelectMSB = 0
	CC_Modulation    = 1
	CC_DataEntry     = 6
	CC_MainVolume    = 7
	CC_Panpot        = 10
	CC_Expression    = 11
	CC_BankSelectLSB = 32
	CC_RPNLSB        = 100
	CC_RPNMSB        = 101
	CC_AllSoundOff   = 120
	CC_MonoOn        = 126
	CC_PolyOn        = 127
)

func (cc CC) String() string {
	s := "unknown"
	switch cc {
	case CC_BankSelectMSB:
		s = "BankSelectMSB"
	case CC_Modulation:
		s = "Modulation"
	case CC_DataEntry:
		s = "DataEntry"
	case CC_MainVolume:
		s = "MainVolume"
	case CC_Panpot:
		s = "Panpot"
	case CC_Expression:
		s = "Expression"
	case CC_BankSelectLSB:
		s = "BankSelectLSB"
	case CC_RPNLSB:
		s = "RPNLSB"
	case CC_RPNMSB:
		s = "RPNMSB"
	case CC_AllSoundOff:
		s = "AllSoundOff"
	case CC_MonoOn:
		s = "MonoOn"
	case CC_PolyOn:
		s = "PolyOn"
	}
	return fmt.Sprintf("%s(%d)", s, int(cc))
}
