package huffman

// 参考 http://oku.edu.mie-u.ac.jp/~okumura/algo/

import (
	"fmt"
	"io"

	"encoding/binary"

	"github.com/pkg/errors"
	//"github.com/but80/smaf825/smaf/log"
)

const (
	n = 256
)

type HuffmanDecoder struct {
	reader      *BitReader
	avail       int
	left, right [2*n - 1]int
}

func (d *HuffmanDecoder) readtree() (int, error) {
	bit, err := d.reader.ReadBit()
	//log.Debugf("%v", bit)
	if err != nil {
		return -1, errors.WithStack(err)
	}
	if bit {
		i := d.avail
		d.avail++
		if 2*n-1 <= i {
			return -1, fmt.Errorf("Invalid huffman table")
		}
		d.left[i], err = d.readtree() // read left branch
		if err != nil {
			return -1, errors.WithStack(err)
		}
		d.right[i], err = d.readtree() // read right branch
		if err != nil {
			return -1, errors.WithStack(err)
		}
		return i, nil // return node
	} else {
		value, err := d.reader.ReadUint8()
		if err != nil {
			return -1, errors.WithStack(err)
		}
		return int(value), nil // return leaf
	}
}

func (d *HuffmanDecoder) Read(p []byte) (int, error) {
	d.avail = 256
	root, err := d.readtree()
	if err != nil {
		return 0, errors.WithStack(err)
	}
	size := len(p)
	//log.Debugf("left: %v", d.left)
	//log.Debugf("right: %v", d.right)
	//log.Debugf("size: %d", size)
	for k := 0; k < size; k++ {
		j := root
		for n <= j {
			b, err := d.reader.ReadBit()
			if err != nil {
				return k, errors.WithStack(err)
			}
			if b {
				j = d.right[j]
			} else {
				j = d.left[j]
			}
		}
		p[k] = byte(j)
	}
	return size, nil
}

func NewHuffmanDecoder(rdr io.Reader) *HuffmanDecoder {
	return &HuffmanDecoder{
		reader: NewBitReader(rdr),
	}
}

type HuffmanReader struct {
	reader  io.Reader
	decoder *HuffmanDecoder
	buf     []byte
}

func NewHuffmanReader(rdr io.Reader) *HuffmanReader {
	return &HuffmanReader{
		reader:  rdr,
		decoder: NewHuffmanDecoder(rdr),
	}
}

func (r *HuffmanReader) cache() error {
	if r.buf != nil {
		return nil
	}
	var size uint32
	err := binary.Read(r.reader, binary.BigEndian, &size)
	if err != nil {
		return errors.WithStack(err)
	}
	r.buf = make([]byte, size)
	_, err = r.decoder.Read(r.buf)
	if err != nil {
		return errors.WithStack(err)
	}
	//log.Debugf("%s", util.Hex(r.buf))
	return nil
}

func (r *HuffmanReader) Rest() (int, error) {
	err := r.cache()
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return len(r.buf), nil
}

func (r *HuffmanReader) Read(p []byte) (int, error) {
	err := r.cache()
	if err != nil {
		return 0, errors.WithStack(err)
	}
	size := len(p)
	var eof error
	if len(r.buf) < size {
		size = len(r.buf)
		eof = io.EOF
	}
	copy(p, r.buf[:size])
	r.buf = r.buf[size:]
	return size, eof
}
