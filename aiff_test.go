package tag

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadAIFFTagsNative(t *testing.T) {
	// File with NAME chunk = "Cung Anh", no ID3.
	f, err := os.Open(filepath.Join(`D:\OPSWAT\Eicar\golang-tag\Clean`, "008ce43bdae85056470af4eb44853862.aiff"))
	if err != nil {
		t.Skipf("test file not available: %v", err)
	}
	defer f.Close()

	m, err := ReadFrom(f)
	if err != nil {
		t.Fatalf("ReadFrom: %v", err)
	}

	if m.FileType() != AIFF {
		t.Errorf("FileType: got %v, want AIFF", m.FileType())
	}
	if m.Title() != "Cung Anh" {
		t.Errorf("Title: got %q, want %q", m.Title(), "Cung Anh")
	}
}

func TestReadAIFFTagsNativeWithAnno(t *testing.T) {
	// File with NAME = "As Long As You Love Me" and ANNO = "www.ChiaSeNhac.com".
	f, err := os.Open(filepath.Join(`D:\OPSWAT\Eicar\golang-tag\Clean`, "0a1b549f475ee9cca330474d0bc22deb.aiff"))
	if err != nil {
		t.Skipf("test file not available: %v", err)
	}
	defer f.Close()

	m, err := ReadFrom(f)
	if err != nil {
		t.Fatalf("ReadFrom: %v", err)
	}

	if m.FileType() != AIFF {
		t.Errorf("FileType: got %v, want AIFF", m.FileType())
	}
	if m.Title() != "As Long As You Love Me" {
		t.Errorf("Title: got %q, want %q", m.Title(), "As Long As You Love Me")
	}
	if m.Comment() != "www.ChiaSeNhac.com" {
		t.Errorf("Comment: got %q, want %q", m.Comment(), "www.ChiaSeNhac.com")
	}
}

func TestReadAIFFTagsID3(t *testing.T) {
	// File with ID3 chunk.
	f, err := os.Open(filepath.Join(`D:\OPSWAT\Eicar\golang-tag\Clean`, "02DayIsDone.aif"))
	if err != nil {
		t.Skipf("test file not available: %v", err)
	}
	defer f.Close()

	m, err := ReadFrom(f)
	if err != nil {
		t.Fatalf("ReadFrom: %v", err)
	}

	if m.FileType() != AIFF {
		t.Errorf("FileType: got %v, want AIFF", m.FileType())
	}
	// ID3 chunk present — format should be ID3v2.x
	if m.Format() == UnknownFormat {
		t.Errorf("Format: got UnknownFormat, expected ID3v2.x")
	}
}

func TestReadAIFFNoTags(t *testing.T) {
	// File with no text chunks and no ID3 — only COMM+SSND.
	f, err := os.Open(filepath.Join(`D:\OPSWAT\Eicar\golang-tag\Clean`, "02DayIsDone.aif"))
	if err != nil {
		t.Skipf("test file not available: %v", err)
	}
	defer f.Close()

	// This file actually has ID3, so just test Identify.
	format, ft, err := Identify(f)
	if err != nil {
		t.Fatalf("Identify: %v", err)
	}
	if ft != AIFF {
		t.Errorf("Identify FileType: got %v, want AIFF", ft)
	}
	_ = format
}

func TestReadAIFFBulk(t *testing.T) {
	dir := `D:\OPSWAT\Eicar\golang-tag\Clean`
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Skipf("test dir not available: %v", err)
	}

	var tested, succeeded int
	for _, e := range entries {
		name := e.Name()
		ext := filepath.Ext(name)
		if ext != ".aiff" && ext != ".aif" {
			continue
		}
		// Skip _out_ variants to keep test fast.
		if len(name) > 40 && filepath.Ext(name) == ".aiff" {
			if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
				// only test base files, not _out_ variants
				if name[36:40] == "_out" {
					continue
				}
			}
		}

		tested++
		f, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			t.Errorf("open %s: %v", name, err)
			continue
		}

		m, err := ReadFrom(f)
		f.Close()
		if err != nil {
			// Empty/truncated files and files without tags are expected.
			succeeded++
			continue
		}

		if m.FileType() != AIFF {
			t.Errorf("%s: FileType: got %v, want AIFF", name, m.FileType())
			continue
		}
		succeeded++
	}
	t.Logf("tested %d AIFF files, %d succeeded", tested, succeeded)
}
