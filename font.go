package facturx

import (
	_ "embed"
	"encoding/binary"
	"sync"
)

//go:embed assets/LiberationSans-Subset.ttf
var fontData []byte

// fontMetrics holds parsed font metrics.
type fontMetrics struct {
	unitsPerEM   uint16
	glyphWidths  map[uint32]uint16
	defaultWidth uint16
	ascender     int16
	descender    int16
}

var (
	cachedMetrics *fontMetrics
	metricsOnce   sync.Once
)

// getFontMetrics returns cached font metrics (parses on first call).
func getFontMetrics() *fontMetrics {
	metricsOnce.Do(func() {
		var err error
		cachedMetrics, err = parseTTF(fontData)
		if err != nil {
			panic("failed to parse embedded font: " + err.Error())
		}
	})
	return cachedMetrics
}

// getFontData returns raw font data for PDF embedding.
func getFontData() []byte {
	return fontData
}

// charWidth returns the advance width for a character in font units.
func (m *fontMetrics) charWidth(c rune) uint16 {
	if w, ok := m.glyphWidths[uint32(c)]; ok {
		return w
	}
	return m.defaultWidth
}

// stringWidth calculates the width of a string at the given font size in points.
func (m *fontMetrics) stringWidth(s string, fontSize float64) float64 {
	var totalWidth uint32
	for _, c := range s {
		totalWidth += uint32(m.charWidth(c))
	}
	return float64(totalWidth) * fontSize / float64(m.unitsPerEM)
}

// ============================================================================
// TTF Parser Implementation
// ============================================================================

// tableEntry represents a TTF table directory entry.
type tableEntry struct {
	offset uint32
	length uint32
}

// findTable finds a table by its 4-byte tag.
func findTable(data []byte, tag string) (tableEntry, bool) {
	if len(data) < 12 {
		return tableEntry{}, false
	}
	numTables := int(binary.BigEndian.Uint16(data[4:6]))

	for i := 0; i < numTables; i++ {
		entryOffset := 12 + i*16
		if entryOffset+16 > len(data) {
			break
		}
		if string(data[entryOffset:entryOffset+4]) == tag {
			return tableEntry{
				offset: binary.BigEndian.Uint32(data[entryOffset+8 : entryOffset+12]),
				length: binary.BigEndian.Uint32(data[entryOffset+12 : entryOffset+16]),
			}, true
		}
	}
	return tableEntry{}, false
}

// parseHead parses the 'head' table to get unitsPerEm.
func parseHead(data []byte, table tableEntry) (uint16, error) {
	offset := int(table.offset)
	if table.length < 54 || offset+54 > len(data) {
		return 0, errTableTooSmall
	}
	// unitsPerEm is at offset 18 within head table
	return binary.BigEndian.Uint16(data[offset+18 : offset+20]), nil
}

// parseHhea parses the 'hhea' table to get numberOfHMetrics and ascender/descender.
func parseHhea(data []byte, table tableEntry) (numHMetrics uint16, ascender, descender int16, err error) {
	offset := int(table.offset)
	if table.length < 36 || offset+36 > len(data) {
		return 0, 0, 0, errTableTooSmall
	}
	// ascender at offset 4, descender at offset 6
	ascender = int16(binary.BigEndian.Uint16(data[offset+4 : offset+6]))
	descender = int16(binary.BigEndian.Uint16(data[offset+6 : offset+8]))
	// numberOfHMetrics is at offset 34 within hhea table
	numHMetrics = binary.BigEndian.Uint16(data[offset+34 : offset+36])
	return numHMetrics, ascender, descender, nil
}

// parseHmtx parses the 'hmtx' table to get glyph advance widths.
func parseHmtx(data []byte, table tableEntry, numHMetrics uint16) []uint16 {
	offset := int(table.offset)
	widths := make([]uint16, 0, numHMetrics)

	for i := 0; i < int(numHMetrics); i++ {
		pos := offset + i*4
		if pos+2 > len(data) {
			break
		}
		// Each longHorMetric is 4 bytes: advanceWidth (u16) + leftSideBearing (i16)
		advanceWidth := binary.BigEndian.Uint16(data[pos : pos+2])
		widths = append(widths, advanceWidth)
	}

	return widths
}

// parseCmapFormat4 parses a cmap format 4 subtable (Unicode BMP).
func parseCmapFormat4(data []byte, subtableOffset int, glyphWidthsRaw []uint16) (map[uint32]uint16, error) {
	if subtableOffset+14 > len(data) {
		return nil, errTableTooSmall
	}

	segCountX2 := int(binary.BigEndian.Uint16(data[subtableOffset+6 : subtableOffset+8]))
	segCount := segCountX2 / 2

	// Table layout after header (14 bytes):
	// endCode[segCount], reservedPad, startCode[segCount], idDelta[segCount], idRangeOffset[segCount], glyphIdArray[]
	endCodesOffset := subtableOffset + 14
	startCodesOffset := endCodesOffset + segCountX2 + 2 // +2 for reservedPad
	idDeltaOffset := startCodesOffset + segCountX2
	idRangeOffsetOffset := idDeltaOffset + segCountX2

	charToWidth := make(map[uint32]uint16)
	defaultWidth := uint16(600)
	if len(glyphWidthsRaw) > 0 {
		defaultWidth = glyphWidthsRaw[0]
	}

	for seg := 0; seg < segCount; seg++ {
		endCodePos := endCodesOffset + seg*2
		startCodePos := startCodesOffset + seg*2
		idDeltaPos := idDeltaOffset + seg*2
		idRangeOffsetPos := idRangeOffsetOffset + seg*2

		if idRangeOffsetPos+2 > len(data) {
			break
		}

		endCode := uint32(binary.BigEndian.Uint16(data[endCodePos : endCodePos+2]))
		startCode := uint32(binary.BigEndian.Uint16(data[startCodePos : startCodePos+2]))
		idDelta := int32(int16(binary.BigEndian.Uint16(data[idDeltaPos : idDeltaPos+2])))
		idRangeOffset := int(binary.BigEndian.Uint16(data[idRangeOffsetPos : idRangeOffsetPos+2]))

		if startCode == 0xFFFF {
			break
		}

		for code := startCode; code <= endCode; code++ {
			var glyphIndex uint16
			if idRangeOffset == 0 {
				glyphIndex = uint16((int32(code) + idDelta) & 0xFFFF)
			} else {
				// Calculate offset into glyphIdArray
				glyphOffset := idRangeOffsetPos + idRangeOffset + int(code-startCode)*2

				if glyphOffset+2 <= len(data) {
					glyphID := binary.BigEndian.Uint16(data[glyphOffset : glyphOffset+2])
					if glyphID != 0 {
						glyphIndex = uint16((int32(glyphID) + idDelta) & 0xFFFF)
					}
				}
			}

			var width uint16
			if int(glyphIndex) < len(glyphWidthsRaw) {
				width = glyphWidthsRaw[glyphIndex]
			} else if len(glyphWidthsRaw) > 0 {
				// For glyphs beyond numberOfHMetrics, use the last width
				width = glyphWidthsRaw[len(glyphWidthsRaw)-1]
			} else {
				width = defaultWidth
			}

			charToWidth[code] = width
		}
	}

	return charToWidth, nil
}

// parseCmap parses the 'cmap' table to build character -> glyph width mapping.
func parseCmap(data []byte, table tableEntry, glyphWidthsRaw []uint16) (map[uint32]uint16, error) {
	offset := int(table.offset)
	if offset+4 > len(data) {
		return nil, errTableTooSmall
	}
	numTables := int(binary.BigEndian.Uint16(data[offset+2 : offset+4]))

	// Look for platform 3 (Windows), encoding 1 (Unicode BMP) - format 4
	// Or platform 0 (Unicode), encoding 3 (Unicode BMP) - format 4
	for i := 0; i < numTables; i++ {
		recordOffset := offset + 4 + i*8
		if recordOffset+8 > len(data) {
			break
		}
		platformID := binary.BigEndian.Uint16(data[recordOffset : recordOffset+2])
		encodingID := binary.BigEndian.Uint16(data[recordOffset+2 : recordOffset+4])
		subtableOffset := offset + int(binary.BigEndian.Uint32(data[recordOffset+4:recordOffset+8]))

		if subtableOffset+2 > len(data) {
			continue
		}
		format := binary.BigEndian.Uint16(data[subtableOffset : subtableOffset+2])

		// Accept format 4 tables for Unicode BMP
		if format == 4 {
			if (platformID == 3 && encodingID == 1) || (platformID == 0 && encodingID == 3) {
				return parseCmapFormat4(data, subtableOffset, glyphWidthsRaw)
			}
		}
	}

	// Fallback: try any format 4 table
	for i := 0; i < numTables; i++ {
		recordOffset := offset + 4 + i*8
		if recordOffset+8 > len(data) {
			break
		}
		subtableOffset := offset + int(binary.BigEndian.Uint32(data[recordOffset+4:recordOffset+8]))

		if subtableOffset+2 > len(data) {
			continue
		}
		format := binary.BigEndian.Uint16(data[subtableOffset : subtableOffset+2])

		if format == 4 {
			return parseCmapFormat4(data, subtableOffset, glyphWidthsRaw)
		}
	}

	return nil, errNoCmapSubtable
}

// parseTTF parses a TTF font and extracts metrics.
func parseTTF(data []byte) (*fontMetrics, error) {
	if len(data) < 12 {
		return nil, errInvalidTTF
	}

	// Verify TrueType signature
	sfntVersion := binary.BigEndian.Uint32(data[0:4])
	if sfntVersion != 0x00010000 && sfntVersion != 0x4F54544F {
		return nil, errInvalidTTF
	}

	// Parse required tables
	head, ok := findTable(data, "head")
	if !ok {
		return nil, errMissingTable
	}
	hhea, ok := findTable(data, "hhea")
	if !ok {
		return nil, errMissingTable
	}
	hmtx, ok := findTable(data, "hmtx")
	if !ok {
		return nil, errMissingTable
	}
	cmap, ok := findTable(data, "cmap")
	if !ok {
		return nil, errMissingTable
	}

	unitsPerEM, err := parseHead(data, head)
	if err != nil {
		return nil, err
	}

	numHMetrics, ascender, descender, err := parseHhea(data, hhea)
	if err != nil {
		return nil, err
	}

	glyphWidthsRaw := parseHmtx(data, hmtx, numHMetrics)

	defaultWidth := uint16(600)
	if len(glyphWidthsRaw) > 0 {
		defaultWidth = glyphWidthsRaw[0]
	}

	glyphWidths, err := parseCmap(data, cmap, glyphWidthsRaw)
	if err != nil {
		return nil, err
	}

	return &fontMetrics{
		unitsPerEM:   unitsPerEM,
		glyphWidths:  glyphWidths,
		defaultWidth: defaultWidth,
		ascender:     ascender,
		descender:    descender,
	}, nil
}

// Errors
type fontError string

func (e fontError) Error() string { return string(e) }

const (
	errInvalidTTF     fontError = "invalid TTF signature"
	errMissingTable   fontError = "missing required table"
	errTableTooSmall  fontError = "table too small"
	errNoCmapSubtable fontError = "no suitable cmap subtable found"
)
