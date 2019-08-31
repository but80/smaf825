package event

// DurationEventPair は、次の演奏イベントと、そのイベントまでの待機時間のペアを格納する構造体です。
type DurationEventPair struct {
	Duration int   `json:"duration"`
	Event    Event `json:"event"`
}
