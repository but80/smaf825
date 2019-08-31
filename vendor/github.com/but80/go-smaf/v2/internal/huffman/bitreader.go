package huffman

import (
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

// BitReader は、バイト列からビット単位で値を読み取れる Reader ストリームです。
type BitReader struct {
	reader io.Reader
	buf    uint8
	rest   int
}

// NewBitReader は、新しい BitReader を作成します。
func NewBitReader(rdr io.Reader) *BitReader {
	return &BitReader{reader: rdr}
}

// ReadBit は、ストリームから1ビット読み取ります。
func (r *BitReader) ReadBit() (bool, error) {
	if r.rest == 0 {
		err := binary.Read(r.reader, binary.LittleEndian, &r.buf)
		if err != nil {
			return false, errors.WithStack(err)
		}
		r.rest = 8
	}
	/*
	   buf 01234567  rest=8
	   buf 1234567-  rest=7
	   buf 234567--  rest=6
	   ..
	   buf 7-------  rest=1
	   buf --------  rest=0
	*/
	result := r.buf&0x80 != 0
	r.buf <<= 1
	r.rest--
	return result, nil
}

// ReadUint8 は、ストリームから8ビット読み取り、符号なし整数として返します。
func (r *BitReader) ReadUint8() (uint8, error) {
	var buf2 uint8
	err := binary.Read(r.reader, binary.LittleEndian, &buf2)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	result := r.buf | buf2>>uint(r.rest)
	r.buf = buf2 << uint(8-r.rest)
	return result, nil
}
