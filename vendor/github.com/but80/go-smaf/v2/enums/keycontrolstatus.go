package enums

import (
	"encoding/json"
	"fmt"
)

// KeyControlStatus は、キーのオン・オフ等のコントロール状態を表す列挙子型です。
type KeyControlStatus int

const (
	// KeyControlStatusNonSpecified は、キーオン・オフのいずれでもない状態です。
	KeyControlStatusNonSpecified = iota
	// KeyControlStatusOff は、キーオフ状態です。
	KeyControlStatusOff
	// KeyControlStatusOn は、キーオン状態です。
	KeyControlStatusOn
)

func (t KeyControlStatus) String() string {
	var s string
	switch t {
	case KeyControlStatusNonSpecified:
		s = "NonSpecified"
	case KeyControlStatusOff:
		s = "Off"
	case KeyControlStatusOn:
		s = "On"
	default:
		s = fmt.Sprintf("undefined (0x%02X)", int(t))
	}
	return s
}

func (t KeyControlStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}
