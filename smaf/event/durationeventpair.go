package event

type DurationEventPair struct {
	Duration int   `json:"duration"`
	Event    Event `json:"event"`
}
