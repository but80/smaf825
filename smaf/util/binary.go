package util

import (
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

func ReadVariableInt(allow3bytes bool, rdr io.Reader, rest *int) (int, error) {
	result := int(0)
	i := 0
	for {
		var b uint8
		err := binary.Read(rdr, binary.BigEndian, &b)
		if err != nil {
			return 0, errors.WithStack(err)
		}
		*rest--
		if !allow3bytes && i == 1 {
			return (result + 0x80) | int(b), nil
		}
		result |= int(b) & 0x7F
		if (b & 0x80) == 0 {
			break
		}
		result <<= 7
		i++
	}
	return result, nil
}

func BoolToByte(b bool, v byte) byte {
	if b {
		return v
	}
	return 0
}

func BytesToInts(b []byte) []int {
	result := make([]int, len(b))
	for i, v := range b {
		result[i] = int(v)
	}
	return result
}
