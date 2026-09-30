// Copyright 2015, David Howden
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tag

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"
)

// ReadAIFFTags reads AIFF metadata from the io.ReadSeeker, returning the resulting
// metadata in a Metadata implementation, or non-nil error if there was a problem.
func ReadAIFFTags(r io.ReadSeeker) (Metadata, error) {
	form, err := readString(r, 4)
	if err != nil {
		return nil, err
	}
	if form != "FORM" {
		return nil, errors.New("expected 'FORM'")
	}

	// Skip FORM chunk size (4 bytes).
	if _, err := r.Seek(4, io.SeekCurrent); err != nil {
		return nil, err
	}

	formType, err := readString(r, 4)
	if err != nil {
		return nil, err
	}
	if formType != "AIFF" && formType != "AIFC" {
		return nil, fmt.Errorf("expected 'AIFF' or 'AIFC', got '%s'", formType)
	}

	var (
		native = make(map[string]string)
		id3    Metadata
	)

	for {
		chunkID, err := readString(r, 4)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return nil, err
		}

		var chunkSize uint32
		if err := binary.Read(r, binary.BigEndian, &chunkSize); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return nil, err
		}

		switch chunkID {
		case "ID3 ":
			id3, err = ReadID3v2Tags(r)
			if err != nil {
				return nil, fmt.Errorf("error reading ID3 chunk: %v", err)
			}
			return metadataAIFF{id3: id3}, nil

		case "NAME":
			s, err := readString(r, uint(chunkSize))
			if err != nil {
				return nil, err
			}
			native["NAME"] = strings.TrimRight(s, "\x00")
			if chunkSize%2 != 0 {
				r.Seek(1, io.SeekCurrent)
			}

		case "AUTH":
			s, err := readString(r, uint(chunkSize))
			if err != nil {
				return nil, err
			}
			native["AUTH"] = strings.TrimRight(s, "\x00")
			if chunkSize%2 != 0 {
				r.Seek(1, io.SeekCurrent)
			}

		case "(c) ":
			s, err := readString(r, uint(chunkSize))
			if err != nil {
				return nil, err
			}
			native["(c) "] = strings.TrimRight(s, "\x00")
			if chunkSize%2 != 0 {
				r.Seek(1, io.SeekCurrent)
			}

		case "ANNO":
			s, err := readString(r, uint(chunkSize))
			if err != nil {
				return nil, err
			}
			native["ANNO"] = strings.TrimRight(s, "\x00")
			if chunkSize%2 != 0 {
				r.Seek(1, io.SeekCurrent)
			}

		default:
			skip := int64(chunkSize)
			if chunkSize%2 != 0 {
				skip++
			}
			if _, err := r.Seek(skip, io.SeekCurrent); err != nil {
				return nil, err
			}
		}
	}

	if len(native) == 0 {
		return nil, ErrNoTagsFound
	}

	return metadataAIFF{native: native}, nil
}

type metadataAIFF struct {
	id3    Metadata
	native map[string]string
}

func (m metadataAIFF) Format() Format {
	if m.id3 != nil {
		return m.id3.Format()
	}
	return UnknownFormat
}

func (m metadataAIFF) FileType() FileType {
	return AIFF
}

func (m metadataAIFF) Title() string {
	if m.id3 != nil {
		return m.id3.Title()
	}
	return m.native["NAME"]
}

func (m metadataAIFF) Album() string {
	if m.id3 != nil {
		return m.id3.Album()
	}
	return ""
}

func (m metadataAIFF) Artist() string {
	if m.id3 != nil {
		return m.id3.Artist()
	}
	return m.native["AUTH"]
}

func (m metadataAIFF) AlbumArtist() string {
	if m.id3 != nil {
		return m.id3.AlbumArtist()
	}
	return ""
}

func (m metadataAIFF) Composer() string {
	if m.id3 != nil {
		return m.id3.Composer()
	}
	return ""
}

func (m metadataAIFF) Year() int {
	if m.id3 != nil {
		return m.id3.Year()
	}
	return 0
}

func (m metadataAIFF) Genre() string {
	if m.id3 != nil {
		return m.id3.Genre()
	}
	return ""
}

func (m metadataAIFF) Track() (int, int) {
	if m.id3 != nil {
		return m.id3.Track()
	}
	return 0, 0
}

func (m metadataAIFF) Disc() (int, int) {
	if m.id3 != nil {
		return m.id3.Disc()
	}
	return 0, 0
}

func (m metadataAIFF) Picture() *Picture {
	if m.id3 != nil {
		return m.id3.Picture()
	}
	return nil
}

func (m metadataAIFF) Lyrics() string {
	if m.id3 != nil {
		return m.id3.Lyrics()
	}
	return ""
}

func (m metadataAIFF) Comment() string {
	if m.id3 != nil {
		return m.id3.Comment()
	}
	if c, ok := m.native["ANNO"]; ok {
		return c
	}
	return m.native["(c) "]
}

func (m metadataAIFF) Raw() map[string]interface{} {
	if m.id3 != nil {
		return m.id3.Raw()
	}
	raw := make(map[string]interface{})
	for k, v := range m.native {
		raw[k] = v
	}
	return raw
}
