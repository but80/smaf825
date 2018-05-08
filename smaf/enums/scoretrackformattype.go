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
	ScoreTrackFormatType_SEQU    = -1
	ScoreTrackFormatType_Default = ScoreTrackFormatType_HandyPhoneStandard
)

func (t ScoreTrackFormatType) IsSupported() bool {
	switch t {
	case ScoreTrackFormatType_HandyPhoneStandard, ScoreTrackFormatType_MobileStandardCompressed, ScoreTrackFormatType_MobileStandardNonCompressed:
		return true
	}
	return false
}

func (t ScoreTrackFormatType) String() string {
	var s string
	switch t {
	case ScoreTrackFormatType_HandyPhoneStandard:
		s = "HandyPhoneStandard"
	case ScoreTrackFormatType_MobileStandardCompressed:
		s = "MobileStandardCompressed"
	case ScoreTrackFormatType_MobileStandardNonCompressed:
		s = "MobileStandardNonCompressed"
	default:
		s = fmt.Sprintf("undefined (0x%02X)", int(t))
	}
	return s
}

func (t ScoreTrackFormatType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}
