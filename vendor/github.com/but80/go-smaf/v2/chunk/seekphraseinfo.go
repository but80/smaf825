package chunk

import (
	"fmt"
	"io"
	"strings"

	"github.com/but80/go-smaf/v2/internal/util"
	"github.com/pkg/errors"
)

type SeekPhraseInfoChunk struct {
	*ChunkHeader
	Stream []uint8 `json:"stream"`
}

func (c *SeekPhraseInfoChunk) Traverse(fn func(Chunk)) {
	fn(c)
}

func (c *SeekPhraseInfoChunk) String() string {
	result := "SeekPhraseInfoChunk: " + c.ChunkHeader.String()
	sub := []string{
		fmt.Sprintf("Stream: %s", util.Escape(c.Stream)),
	}
	return result + "\n" + util.Indent(strings.Join(sub, "\n"), "\t")
}

// Read は、バイト列を読み取ってパースした結果をこの構造体に格納します。
func (c *SeekPhraseInfoChunk) Read(rdr io.Reader) error {
	c.Stream = make([]uint8, c.ChunkHeader.Size)
	n, err := rdr.Read(c.Stream)
	if err != nil {
		return errors.WithStack(err)
	}
	if n < len(c.Stream) {
		return errors.Errorf("Cannot read enough byte length specified in chunk header")
	}
	return nil
}
