package chunk

import (
	"fmt"
	"io"
	"strings"

	"github.com/but80/go-smaf/v2/internal/util"
	"github.com/pkg/errors"
)

type DataChunk struct {
	*ChunkHeader
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
		EditStatus      string `json:"edit_status,omitempty"`
		VCard           string `json:"vcard,omitempty"`
	} `json:"options,omitempty"`
}

func (c *DataChunk) Traverse(fn func(Chunk)) {
	fn(c)
}

func (c *DataChunk) CodeType() int {
	return int(c.ChunkHeader.Signature & 255)
}

func (c *DataChunk) String() string {
	result := "DataChunk: " + c.ChunkHeader.String()
	sub := []string{
		fmt.Sprintf("Code type: 0x%02X", c.CodeType()),
		fmt.Sprintf("Stream: %s", util.Escape(c.Stream)),
		fmt.Sprintf("Options: %+v", c.Options),
	}
	return result + "\n" + util.Indent(strings.Join(sub, "\n"), "\t")
}

// Read は、バイト列を読み取ってパースした結果をこの構造体に格納します。
func (c *DataChunk) Read(rdr io.Reader) error {
	c.Stream = make([]uint8, c.ChunkHeader.Size)
	n, err := rdr.Read(c.Stream)
	if err != nil {
		return err
	}
	if n < len(c.Stream) {
		return errors.Errorf("Cannot read enough byte length specified in chunk header")
	}
	if c.CodeType() == 0x00 {
		options := map[string]string{}
		i := 0
		for i < len(c.Stream) {
			tag := string(c.Stream[i : i+2])
			i += 2
			size := int(c.Stream[i])<<8 | int(c.Stream[i+1])
			i += 2
			value := util.DecodeShiftJIS(c.Stream[i : i+size])
			i += size
			options[tag] = value
		}
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
		c.Options.EditStatus = options["ES"]
		c.Options.VCard = options["VC"]
	}
	return nil
}
