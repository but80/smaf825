package util

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"golang.org/x/text/encoding/japanese"
)

var indentRe = regexp.MustCompile("(?m)^")

func Indent(text string, indent string) string {
	if text == "" {
		return text
	}
	return indentRe.ReplaceAllString(text, indent)
}

func Hex(stream []uint8) string {
	if len(stream) == 0 {
		return "[]"
	}
	s := ""
	for _, b := range stream {
		s += fmt.Sprintf(" %02X", b)
	}
	return "[" + s[1:] + "]"
}

func Escape(stream []uint8) string {
	j, err := json.Marshal(string(stream))
	if err != nil {
		return fmt.Sprintf("%+v", stream)
	}
	return string(j)
}

func ZeroPadSliceToString(s []byte) string {
	i := len(s)
	for 0 < i && s[i-1] == 0 {
		i--
	}
	return string(s[:i])
}

func DecodeShiftJIS(s []uint8) string {
	decoder := japanese.ShiftJIS.NewDecoder()
	reader := bufio.NewReader(decoder.Reader(bytes.NewReader([]byte(s))))
	result := []string{}
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			//panic(err)
			break
		}
		result = append(result, string(line))
	}
	return strings.Join(result, "\n")
}

var splitOptionalDataRe1 = regexp.MustCompile(`([^\\,]|\\.)+`)
var splitOptionalDataRe2 = regexp.MustCompile(`\\.`)

func SplitOptionalData(s string) map[string]string {
	result := map[string]string{}
	pairs := splitOptionalDataRe1.FindAllString(s, -1)
	for _, pair := range pairs {
		p := strings.SplitN(pair, ":", 2)
		result[p[0]] = splitOptionalDataRe2.ReplaceAllStringFunc(p[1], func(s string) string {
			return s[1:]
		})
	}
	return result
}
