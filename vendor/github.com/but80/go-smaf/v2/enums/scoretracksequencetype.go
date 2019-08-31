package enums

import (
	"encoding/json"
	"fmt"
)

type ScoreTrackSequenceType int

const (
	// Sequence Data は1つの連続したシーケンス・データである。Seek Point や Phrase List はシーケンス中の意味のある位置を外部から参照する目的で利用する。
	ScoreTrackSequenceTypeStreamSequence ScoreTrackSequenceType = iota
	// Sequence Data は複数のフレーズデータを連続で表記したものである。Phrase List は外部から個別フレーズを認識する為に用いる。
	ScoreTrackSequenceTypeSubsequence
)

func (t ScoreTrackSequenceType) String() string {
	var s string
	switch t {
	case ScoreTrackSequenceTypeStreamSequence:
		s = "StreamSequence"
	case ScoreTrackSequenceTypeSubsequence:
		s = "Subsequence"
	default:
		s = fmt.Sprintf("undefined (0x%02X)", int(t))
	}
	return s
}

func (t ScoreTrackSequenceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}
