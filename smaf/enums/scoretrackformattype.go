package enums

import (
	"encoding/json"
	"fmt"
)

type ScoreTrackFormatType int

const (
	ScoreTrackFormatType_HandyPhoneStandard ScoreTrackFormatType = iota
	ScoreTrackFormatType_MobileStandardCompressed
	ScoreTrackFormatType_MobileStandardNonCompressed
	ScoreTrackFormatType_Default = ScoreTrackFormatType_HandyPhoneStandard
)

func (t ScoreTrackFormatType) String() string {
	s := "undefined"
	switch t {
	case ScoreTrackFormatType_HandyPhoneStandard:
		s = "HandyPhoneStandard"
	case ScoreTrackFormatType_MobileStandardCompressed:
		s = "MobileStandardCompressed"
	case ScoreTrackFormatType_MobileStandardNonCompressed:
		s = "MobileStandardNonCompressed"
	}
	return fmt.Sprintf("%s(0x%02X)", s, int(t))
}

func (t ScoreTrackFormatType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}
