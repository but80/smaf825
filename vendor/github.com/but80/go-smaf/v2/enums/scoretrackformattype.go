package enums

import (
	"encoding/json"
	"fmt"
)

type ScoreTrackFormatType int

const (
	ScoreTrackFormatTypeHandyPhoneStandard ScoreTrackFormatType = iota
	ScoreTrackFormatTypeMobileStandardCompressed
	ScoreTrackFormatTypeMobileStandardNonCompressed
	ScoreTrackFormatTypeSEQU    = -1
	ScoreTrackFormatTypeDefault = ScoreTrackFormatTypeHandyPhoneStandard
)

func (t ScoreTrackFormatType) IsSupported() bool {
	switch t {
	case ScoreTrackFormatTypeHandyPhoneStandard, ScoreTrackFormatTypeMobileStandardCompressed, ScoreTrackFormatTypeMobileStandardNonCompressed:
		return true
	}
	return false
}

func (t ScoreTrackFormatType) String() string {
	var s string
	switch t {
	case ScoreTrackFormatTypeHandyPhoneStandard:
		s = "HandyPhoneStandard"
	case ScoreTrackFormatTypeMobileStandardCompressed:
		s = "MobileStandardCompressed"
	case ScoreTrackFormatTypeMobileStandardNonCompressed:
		s = "MobileStandardNonCompressed"
	default:
		s = fmt.Sprintf("undefined (0x%02X)", int(t))
	}
	return s
}

func (t ScoreTrackFormatType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}
