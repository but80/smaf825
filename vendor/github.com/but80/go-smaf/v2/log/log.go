package log

import (
	"fmt"
	"os"
	"strings"
)

// Level は、ログレベルを表す列挙子です。
type Level int

const (
	// LevelNone は、ログを記録しないログレベルです。
	LevelNone = iota
	// LevelWarn は、警告のみを記録するログレベルです。
	LevelWarn
	// LevelInfo は、参考情報までを記録するログレベルです。
	LevelInfo
	// LevelDebug は、デバッグ情報までを記録するログレベルです。
	LevelDebug
)

// currentLevel は、現在のログレベルです。
var currentLevel Level = LevelInfo

// SetLevel は、ログレベルを設定します。
func SetLevel(level Level) Level {
	old := currentLevel
	currentLevel = level
	return old
}

// Warnf は、Warnレベルのログをフォーマットして記録します。
func Warnf(f string, args ...interface{}) {
	if LevelWarn <= currentLevel {
		fmt.Fprintf(os.Stderr, "[WARNING] "+f+"\n", args...)
	}
}

// Infof は、Infoレベルのログをフォーマットして記録します。
func Infof(f string, args ...interface{}) {
	if LevelInfo <= currentLevel {
		fmt.Fprintf(os.Stderr, f+"\n", args...)
	}
}

var indent = 0

// Debugf は、Debugレベルのログをフォーマットして記録します。
func Debugf(f string, args ...interface{}) {
	if LevelDebug <= currentLevel {
		fmt.Fprintf(os.Stderr, strings.Repeat("  ", indent)+f+"\n", args...)
	}
}

// Enter は、この後に記録するログのインデントレベルを一段深くします。
func Enter() {
	indent++
}

// Leave は、この後に記録するログのインデントレベルを一段浅くします。
func Leave() {
	indent--
}
