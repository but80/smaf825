package subtypes

import "github.com/mersenne-sister/smaf825/smaf/enums"

type ChannelStatus struct {
	// Key Control のリクエストを受けた時、該当 Channel に対し、Key Control を行うか否かの指定をする。
	// ON で Key Control を有効とする。無指定における解釈は実装依存とする。
	KeyControlStatus enums.KeyControlStatus
	// Vibration Control のリクエストを受けた時、該当 Channel のデータに同期して Vibration を行うか否かの指定をする。
	// ON で Vibration を有効とする。
	VibrationStatus bool
	// LED Control のリクエストを受けた時、該当 Channel のデータに同期して LED の制御を行う否かの指定をする。
	// ON で LED を有効とする。
	LEDStatus bool
	// 該当 Channel に対し、Channel Type を指定する。
	ChannelType enums.ChannelType
}
