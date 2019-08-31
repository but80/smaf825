package chunk

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"unsafe"

	"github.com/but80/go-smaf/v2/internal/util"
	"github.com/pkg/errors"
)

type ContentsInfoChunk struct {
	*ChunkHeader
	Header struct {
		ContentsClass    uint8 `json:"contents_class"`
		ContentsType     uint8 `json:"contents_type"`
		ContentsCodeType uint8 `json:"contents_code_type"`
		CopyStatus       uint8 `json:"copy_status"`
		CopyCounts       uint8 `json:"copy_counts"`
	} `json:"header"`
	Stream     []uint8 `json:"stream"`
	HasOptions bool    `json:"has_options"`
	Options    struct {
		Vendor          string `json:"vendor,omitempty"`
		Carrier         string `json:"carrier,omitempty"`
		Category        string `json:"category,omitempty"`
		Title           string `json:"title,omitempty"`
		Artist          string `json:"artist,omitempty"`
		LyricWriter     string `json:"lyric_writer,omitempty"`
		Composer        string `json:"composer,omitempty"`
		Arranger        string `json:"arranger,omitempty"`
		Copyright       string `json:"copyright,omitempty"`
		ManagementGroup string `json:"management_group,omitempty"`
		ManagementInfo  string `json:"management_info,omitempty"`
		CreatedDate     string `json:"created_date,omitempty"`
		UpdatedDate     string `json:"updated_date,omitempty"`
	} `json:"options,omitempty"`
}

func (c *ContentsInfoChunk) Traverse(fn func(Chunk)) {
	fn(c)
}

func (c *ContentsInfoChunk) String() string {
	result := "ContentsInfoChunk: " + c.ChunkHeader.String()
	sub := []string{
		fmt.Sprintf("ContentsClass: 0x%02X", c.Header.ContentsClass),
		fmt.Sprintf("ContentsType: 0x%02X", c.Header.ContentsType),
		fmt.Sprintf("ContentsCodeType: 0x%02X", c.Header.ContentsCodeType),
		fmt.Sprintf("CopyStatus: 0x%02X", c.Header.CopyStatus),
		fmt.Sprintf("CopyCounts: 0x%02X", c.Header.CopyCounts),
		fmt.Sprintf("Stream: %s", util.Escape(c.Stream)),
		fmt.Sprintf("Options: %+v", c.Options),
	}
	return result + "\n" + util.Indent(strings.Join(sub, "\n"), "\t")
}

// Read は、バイト列を読み取ってパースした結果をこの構造体に格納します。
func (c *ContentsInfoChunk) Read(rdr io.Reader) error {
	rest := int(c.ChunkHeader.Size)
	binary.Read(rdr, binary.BigEndian, &c.Header)
	rest -= int(unsafe.Sizeof(c.Header))
	c.Stream = make([]uint8, rest)
	n, err := rdr.Read(c.Stream)
	if err != nil {
		return errors.WithStack(err)
	}
	if n < len(c.Stream) {
		return errors.Errorf("Cannot read enough byte length specified in chunk header")
	}
	if c.Header.ContentsCodeType == 0x00 {
		options := util.SplitOptionalData(util.DecodeShiftJIS(c.Stream))
		c.HasOptions = true
		c.Options.Vendor = options["VN"]
		c.Options.Carrier = options["CN"]
		c.Options.Category = options["CA"]
		c.Options.Title = options["ST"]
		c.Options.Artist = options["AN"]
		c.Options.LyricWriter = options["WW"]
		c.Options.Composer = options["SW"]
		c.Options.Arranger = options["AW"]
		c.Options.Copyright = options["CR"]
		c.Options.ManagementGroup = options["GR"]
		c.Options.ManagementInfo = options["MI"]
		c.Options.CreatedDate = options["CD"]
		c.Options.UpdatedDate = options["UD"]
	}
	return nil
}
