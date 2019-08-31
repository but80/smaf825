package enums

import "fmt"

type CC int

const (
	CCBankSelectMSB = 0
	CCModulation    = 1
	CCDataEntry     = 6
	CCMainVolume    = 7
	CCPanpot        = 10
	CCExpression    = 11
	CCBankSelectLSB = 32
	CCRPNLSB        = 100
	CCRPNMSB        = 101
	CCAllSoundOff   = 120
	CCMonoOn        = 126
	CCPolyOn        = 127
)

func (cc CC) String() string {
	s := "unknown"
	switch cc {
	case CCBankSelectMSB:
		s = "BankSelectMSB"
	case CCModulation:
		s = "Modulation"
	case CCDataEntry:
		s = "DataEntry"
	case CCMainVolume:
		s = "MainVolume"
	case CCPanpot:
		s = "Panpot"
	case CCExpression:
		s = "Expression"
	case CCBankSelectLSB:
		s = "BankSelectLSB"
	case CCRPNLSB:
		s = "RPNLSB"
	case CCRPNMSB:
		s = "RPNMSB"
	case CCAllSoundOff:
		s = "AllSoundOff"
	case CCMonoOn:
		s = "MonoOn"
	case CCPolyOn:
		s = "PolyOn"
	}
	return fmt.Sprintf("%s(%d)", s, int(cc))
}
