package subtypes

import "github.com/but80/go-smaf/v2/enums"

type ChannelStatus struct {
	// Key Control のリクエストを受けた時、該当 Channel に対し、Key Control を行うか否かの指定をする。
	// ON で Key Control を有効とする。無指定における解釈は実装依存とする。
	KeyControlStatus enums.KeyControlStatus `json:"key_control_status"`
	// Vibration Control のリクエストを受けた時、該当 Channel のデータに同期して Vibration を行うか否かの指定をする。
	// ON で Vibration を有効とする。
	VibrationStatus bool `json:"vibration_status"`
	// LED Control のリクエストを受けた時、該当 Channel のデータに同期して LED の制御を行う否かの指定をする。
	// ON で LED を有効とする。
	LEDStatus bool `json:"led_status"`
	// 該当 Channel に対し、Channel Type を指定する。
	ChannelType enums.ChannelType `json:"channel_type"`
}
