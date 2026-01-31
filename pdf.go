package facturx

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
)

//go:embed assets/sRGB-IEC61966-2.1.icc
var srgbICCProfile []byte

// pdfBuilder builds a PDF document.
type pdfBuilder struct {
	objects []pdfObject
	offsets []int
	buffer  bytes.Buffer
}

// pdfObject represents a PDF object.
type pdfObject struct {
	num     int
	gen     int
	content []byte
	stream  []byte
}

func newPDFBuilder() *pdfBuilder {
	return &pdfBuilder{
		objects: make([]pdfObject, 0, 16),
	}
}

// addObject adds an object and returns its object number.
func (b *pdfBuilder) addObject(content []byte, stream []byte) int {
	num := len(b.objects) + 1
	b.objects = append(b.objects, pdfObject{
		num:     num,
		gen:     0,
		content: content,
		stream:  stream,
	})
	return num
}

// build generates the complete PDF with a file ID.
func (b *pdfBuilder) build(fileID string) []byte {
	b.buffer.Reset()
	b.offsets = make([]int, 0, len(b.objects))

	// PDF header
	b.buffer.WriteString("%PDF-1.7\n")
	// Binary marker (required for PDF/A)
	b.buffer.Write([]byte("%\xE2\xE3\xCF\xD3\n"))

	// Write all objects
	for _, obj := range b.objects {
		b.offsets = append(b.offsets, b.buffer.Len())
		fmt.Fprintf(&b.buffer, "%d %d obj\n", obj.num, obj.gen)
		b.buffer.Write(obj.content)

		if obj.stream != nil {
			b.buffer.WriteString("\nstream\n")
			b.buffer.Write(obj.stream)
			b.buffer.WriteString("\nendstream")
		}

		b.buffer.WriteString("\nendobj\n")
	}

	// Cross-reference table
	xrefOffset := b.buffer.Len()
	b.buffer.WriteString("xref\n")
	fmt.Fprintf(&b.buffer, "0 %d\n", len(b.objects)+1)
	b.buffer.WriteString("0000000000 65535 f \n")
	for _, offset := range b.offsets {
		fmt.Fprintf(&b.buffer, "%010d 00000 n \n", offset)
	}

	// Generate file ID
	idHex := generateFileID(fileID)

	// Trailer with ID (required for PDF/A)
	b.buffer.WriteString("trailer\n")
	fmt.Fprintf(&b.buffer, "<< /Size %d /Root 1 0 R /Info 2 0 R /ID [<%s> <%s>] >>\n",
		len(b.objects)+1, idHex, idHex)
	fmt.Fprintf(&b.buffer, "startxref\n%d\n%%%%EOF\n", xrefOffset)

	return b.buffer.Bytes()
}

// generateFileID generates a 16-byte file ID as hex string from invoice identifier.
func generateFileID(input string) string {
	// Simple hash function (djb2 variant, extended to 128 bits)
	var hash [16]byte
	inputBytes := []byte(input)

	for i, b := range inputBytes {
		idx := i % 16
		hash[idx] = (hash[idx] + b) * 33
		hash[idx] += byte(i) * 7
	}

	// Mix the hash
	for i := 0; i < 16; i++ {
		hash[i] = (hash[i] + hash[(i+7)%16]) * 31
	}

	// Convert to hex
	var result strings.Builder
	for _, b := range hash {
		fmt.Fprintf(&result, "%02X", b)
	}
	return result.String()
}

// generatePDF generates complete PDF/A-3 with embedded Factur-X XML.
func generatePDF(req *InvoiceRequest, xmlContent string) []byte {
	builder := newPDFBuilder()

	// Calculate invoice totals for display
	lineTotal, taxTotal, grandTotal, vatRate, vatText := calculateTotals(req)

	// Font metrics for text layout
	metrics := getFontMetrics()
	fontDataBytes := getFontData()

	// Page dimensions (A4 in points: 595.28 x 841.89)
	pageWidth := 595.28
	pageHeight := 841.89
	margin := 50.0

	// ========================================================================
	// Create PDF objects
	// ========================================================================

	// Object 1: Catalog (root)
	catalogContent := "<< /Type /Catalog /Pages 3 0 R /MarkInfo << /Marked true >> /StructTreeRoot 4 0 R /Metadata 5 0 R /OutputIntents [6 0 R] /Names << /EmbeddedFiles << /Names [(factur-x.xml) 7 0 R] >> >> /AF [7 0 R] >>"
	builder.addObject([]byte(catalogContent), nil) // Obj 1

	// Object 2: Document Info
	infoContent := fmt.Sprintf("<< /Title (Facture %s) /Producer (facturx-go) /CreationDate (D:%s) /ModDate (D:%s) >>",
		escapePDFString(req.Number), req.Date, req.Date)
	builder.addObject([]byte(infoContent), nil) // Obj 2

	// Object 3: Pages
	pagesContent := "<< /Type /Pages /Kids [8 0 R] /Count 1 >>"
	builder.addObject([]byte(pagesContent), nil) // Obj 3

	// Object 4: StructTreeRoot (for tagged PDF)
	structTreeContent := "<< /Type /StructTreeRoot >>"
	builder.addObject([]byte(structTreeContent), nil) // Obj 4

	// Object 5: XMP Metadata
	xmp := generateXMPMetadata(req)
	xmpContent := fmt.Sprintf("<< /Type /Metadata /Subtype /XML /Length %d >>", len(xmp))
	builder.addObject([]byte(xmpContent), []byte(xmp)) // Obj 5

	// Object 6: OutputIntent for PDF/A
	outputIntentContent := "<< /Type /OutputIntent /S /GTS_PDFA1 /OutputConditionIdentifier (sRGB IEC61966-2.1) /RegistryName (http://www.color.org) /Info (sRGB IEC61966-2.1) /DestOutputProfile 9 0 R >>"
	builder.addObject([]byte(outputIntentContent), nil) // Obj 6

	// Object 7: Embedded file filespec
	filespecContent := "<< /Type /Filespec /F (factur-x.xml) /UF (factur-x.xml) /Desc (Factur-X XML invoice) /AFRelationship /Data /EF << /F 10 0 R /UF 10 0 R >> >>"
	builder.addObject([]byte(filespecContent), nil) // Obj 7

	// Object 8: Page
	pageContent := fmt.Sprintf("<< /Type /Page /Parent 3 0 R /MediaBox [0 0 %.2f %.2f] /Contents 11 0 R /Resources << /Font << /F1 12 0 R >> >> >>",
		pageWidth, pageHeight)
	builder.addObject([]byte(pageContent), nil) // Obj 8

	// Object 9: ICC Profile
	iccHex := bytesToHex(srgbICCProfile)
	iccContent := fmt.Sprintf("<< /N 3 /Length %d /Filter /ASCIIHexDecode >>", len(iccHex))
	builder.addObject([]byte(iccContent), iccHex) // Obj 9

	// Object 10: Embedded XML file
	xmlBytes := []byte(xmlContent)
	embeddedFileContent := fmt.Sprintf("<< /Type /EmbeddedFile /Subtype /text#2Fxml /Length %d /Params << /Size %d >> >>",
		len(xmlBytes), len(xmlBytes))
	builder.addObject([]byte(embeddedFileContent), xmlBytes) // Obj 10

	// Object 11: Page content stream
	contentStream := generatePageContent(req, lineTotal, taxTotal, grandTotal, vatRate, vatText, metrics, pageWidth, pageHeight, margin)
	contentObj := fmt.Sprintf("<< /Length %d >>", len(contentStream))
	builder.addObject([]byte(contentObj), contentStream) // Obj 11

	// Object 12: Font dictionary
	fontDictContent := "<< /Type /Font /Subtype /TrueType /BaseFont /LiberationSans /FirstChar 32 /LastChar 255 /FontDescriptor 13 0 R /Encoding /WinAnsiEncoding /Widths 14 0 R >>"
	builder.addObject([]byte(fontDictContent), nil) // Obj 12

	// Object 13: Font descriptor
	fontDescriptorContent := fmt.Sprintf("<< /Type /FontDescriptor /FontName /LiberationSans /Flags 32 /FontBBox [-543 -303 1300 979] /ItalicAngle 0 /Ascent %d /Descent %d /CapHeight 729 /StemV 80 /FontFile2 15 0 R >>",
		metrics.ascender, metrics.descender)
	builder.addObject([]byte(fontDescriptorContent), nil) // Obj 13

	// Object 14: Font widths array (characters 32-255)
	widths := generateFontWidths(metrics)
	widthsContent := fmt.Sprintf("[%s]", widths)
	builder.addObject([]byte(widthsContent), nil) // Obj 14

	// Object 15: Embedded font file (raw binary)
	fontContent := fmt.Sprintf("<< /Length %d /Length1 %d >>", len(fontDataBytes), len(fontDataBytes))
	builder.addObject([]byte(fontContent), fontDataBytes) // Obj 15

	// Generate file ID from invoice number and date
	fileID := fmt.Sprintf("%s_%s", req.Number, req.Date)
	return builder.build(fileID)
}

// generateFontWidths generates font widths for characters 32-255 (scaled to 1000 units).
func generateFontWidths(metrics *fontMetrics) string {
	scale := 1000.0 / float64(metrics.unitsPerEM)
	var widths strings.Builder

	for code := 32; code <= 255; code++ {
		if code > 32 {
			widths.WriteByte(' ')
		}
		width := metrics.charWidth(rune(code))
		scaled := int(float64(width)*scale + 0.5)
		fmt.Fprintf(&widths, "%d", scaled)
	}

	return widths.String()
}

// bytesToHex converts bytes to ASCII hex encoding.
func bytesToHex(data []byte) []byte {
	hex := make([]byte, 0, len(data)*2+1)
	for _, b := range data {
		hex = append(hex, fmt.Sprintf("%02X", b)...)
	}
	hex = append(hex, '>')
	return hex
}

// escapePDFString escapes a string for PDF.
func escapePDFString(s string) string {
	var result strings.Builder
	result.Grow(len(s))
	for _, c := range s {
		switch c {
		case '(':
			result.WriteString("\\(")
		case ')':
			result.WriteString("\\)")
		case '\\':
			result.WriteString("\\\\")
		default:
			result.WriteRune(c)
		}
	}
	return result.String()
}

// generateXMPMetadata generates XMP metadata for PDF/A-3 and Factur-X.
func generateXMPMetadata(req *InvoiceRequest) string {
	return fmt.Sprintf(`<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
  <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
    <rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/">
      <dc:title>
        <rdf:Alt>
          <rdf:li xml:lang="x-default">Facture %s</rdf:li>
        </rdf:Alt>
      </dc:title>
      <dc:creator>
        <rdf:Seq>
          <rdf:li>%s</rdf:li>
        </rdf:Seq>
      </dc:creator>
    </rdf:Description>
    <rdf:Description rdf:about="" xmlns:pdf="http://ns.adobe.com/pdf/1.3/">
      <pdf:Producer>facturx-go</pdf:Producer>
    </rdf:Description>
    <rdf:Description rdf:about="" xmlns:xmp="http://ns.adobe.com/xap/1.0/">
      <xmp:CreateDate>%s-%s-%sT00:00:00+00:00</xmp:CreateDate>
      <xmp:ModifyDate>%s-%s-%sT00:00:00+00:00</xmp:ModifyDate>
    </rdf:Description>
    <rdf:Description rdf:about="" xmlns:pdfaid="http://www.aiim.org/pdfa/ns/id/">
      <pdfaid:part>3</pdfaid:part>
      <pdfaid:conformance>B</pdfaid:conformance>
    </rdf:Description>
    <rdf:Description rdf:about="" xmlns:pdfaExtension="http://www.aiim.org/pdfa/ns/extension/" xmlns:pdfaSchema="http://www.aiim.org/pdfa/ns/schema#" xmlns:pdfaProperty="http://www.aiim.org/pdfa/ns/property#">
      <pdfaExtension:schemas>
        <rdf:Bag>
          <rdf:li rdf:parseType="Resource">
            <pdfaSchema:schema>Factur-X PDFA Extension Schema</pdfaSchema:schema>
            <pdfaSchema:namespaceURI>urn:factur-x:pdfa:CrossIndustryDocument:invoice:1p0#</pdfaSchema:namespaceURI>
            <pdfaSchema:prefix>fx</pdfaSchema:prefix>
            <pdfaSchema:property>
              <rdf:Seq>
                <rdf:li rdf:parseType="Resource">
                  <pdfaProperty:name>DocumentFileName</pdfaProperty:name>
                  <pdfaProperty:valueType>Text</pdfaProperty:valueType>
                  <pdfaProperty:category>external</pdfaProperty:category>
                  <pdfaProperty:description>Name of the embedded XML invoice file</pdfaProperty:description>
                </rdf:li>
                <rdf:li rdf:parseType="Resource">
                  <pdfaProperty:name>DocumentType</pdfaProperty:name>
                  <pdfaProperty:valueType>Text</pdfaProperty:valueType>
                  <pdfaProperty:category>external</pdfaProperty:category>
                  <pdfaProperty:description>Type of the hybrid document</pdfaProperty:description>
                </rdf:li>
                <rdf:li rdf:parseType="Resource">
                  <pdfaProperty:name>Version</pdfaProperty:name>
                  <pdfaProperty:valueType>Text</pdfaProperty:valueType>
                  <pdfaProperty:category>external</pdfaProperty:category>
                  <pdfaProperty:description>Version of the Factur-X standard</pdfaProperty:description>
                </rdf:li>
                <rdf:li rdf:parseType="Resource">
                  <pdfaProperty:name>ConformanceLevel</pdfaProperty:name>
                  <pdfaProperty:valueType>Text</pdfaProperty:valueType>
                  <pdfaProperty:category>external</pdfaProperty:category>
                  <pdfaProperty:description>Conformance level of the Factur-X document</pdfaProperty:description>
                </rdf:li>
              </rdf:Seq>
            </pdfaSchema:property>
          </rdf:li>
        </rdf:Bag>
      </pdfaExtension:schemas>
    </rdf:Description>
    <rdf:Description rdf:about="" xmlns:fx="urn:factur-x:pdfa:CrossIndustryDocument:invoice:1p0#">
      <fx:DocumentFileName>factur-x.xml</fx:DocumentFileName>
      <fx:DocumentType>INVOICE</fx:DocumentType>
      <fx:Version>1.0</fx:Version>
      <fx:ConformanceLevel>BASIC</fx:ConformanceLevel>
    </rdf:Description>
  </rdf:RDF>
</x:xmpmeta>
<?xpacket end="w"?>`,
		escapeXMLAttr(req.Number),
		escapeXMLAttr(req.Seller.Name),
		req.Date[0:4], req.Date[4:6], req.Date[6:8],
		req.Date[0:4], req.Date[4:6], req.Date[6:8])
}

// escapeXMLAttr escapes string for XML attribute.
func escapeXMLAttr(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

// calculateTotals calculates invoice totals.
func calculateTotals(req *InvoiceRequest) (lineTotal, taxTotal, grandTotal, vatRate, vatText string) {
	var lineTotalVal float64
	for _, line := range req.Lines {
		lineTotalVal += line.Quantity * line.UnitPrice
	}

	var vatRateVal float64
	var vatTextVal string
	switch req.Regime.kind {
	case vatFranchiseAuto:
		vatRateVal = 0
		vatTextVal = "TVA non applicable, art. 293 B du CGI"
	case vatExemptHealth:
		vatRateVal = 0
		vatTextVal = "Exonération de TVA, art. 261-4-1° du CGI"
	default:
		vatRateVal = req.Regime.rate
		vatTextVal = fmt.Sprintf("TVA %.0f%%", req.Regime.rate)
	}

	taxTotalVal := lineTotalVal * vatRateVal / 100.0
	grandTotalVal := lineTotalVal + taxTotalVal

	return fmt.Sprintf("%.2f", lineTotalVal),
		fmt.Sprintf("%.2f", taxTotalVal),
		fmt.Sprintf("%.2f", grandTotalVal),
		fmt.Sprintf("%.2f", vatRateVal),
		vatTextVal
}

// generatePageContent generates page content stream (visual invoice layout).
func generatePageContent(req *InvoiceRequest, lineTotal, taxTotal, grandTotal, vatRate, vatText string,
	metrics *fontMetrics, pageWidth, pageHeight, margin float64) []byte {

	var content bytes.Buffer

	// Color definitions (RGB 0-1)
	const (
		primaryR, primaryG, primaryB = 0.173, 0.243, 0.314 // Dark blue-gray #2C3E50
		accentR, accentG, accentB    = 0.204, 0.596, 0.859 // Blue accent #3498DB
		grayR, grayG, grayB          = 0.584, 0.647, 0.651 // Gray text #95A5A6
		lightBgR, lightBgG, lightBgB = 0.925, 0.941, 0.945 // Light gray bg #ECF0F1
	)

	// Start with graphics state
	content.WriteString("q\n")

	// ========================================================================
	// Header band with accent color
	// ========================================================================
	headerHeight := 70.0
	fmt.Fprintf(&content, "%.3f %.3f %.3f rg\n", primaryR, primaryG, primaryB)
	fmt.Fprintf(&content, "0 %.2f %.2f %.2f re f\n", pageHeight-headerHeight, pageWidth, headerHeight)

	// Title in white
	writeTextColored(&content, "FACTURE", margin, pageHeight-45, 28.0, 1, 1, 1)

	// Invoice number in white (smaller, right-aligned concept)
	invoiceInfo := fmt.Sprintf("N° %s", req.Number)
	writeTextColored(&content, invoiceInfo, margin, pageHeight-62, 11.0, 0.8, 0.8, 0.8)

	// ========================================================================
	// Date badge
	// ========================================================================
	dateStr := fmt.Sprintf("%s/%s/%s", req.Date[6:8], req.Date[4:6], req.Date[0:4])
	dateX := pageWidth - margin - 80
	fmt.Fprintf(&content, "1 1 1 rg\n") // White background
	fmt.Fprintf(&content, "%.2f %.2f 75 20 re f\n", dateX-5, pageHeight-50)
	writeTextColored(&content, dateStr, dateX, pageHeight-44, 10.0, primaryR, primaryG, primaryB)

	// ========================================================================
	// Accent line under header
	// ========================================================================
	fmt.Fprintf(&content, "%.3f %.3f %.3f RG\n", accentR, accentG, accentB)
	fmt.Fprintf(&content, "3 w\n") // 3pt line width
	fmt.Fprintf(&content, "%.2f %.2f m %.2f %.2f l S\n", 0.0, pageHeight-headerHeight, pageWidth, pageHeight-headerHeight)
	fmt.Fprintf(&content, "1 w\n") // Reset line width

	// ========================================================================
	// Seller and Buyer blocks
	// ========================================================================
	yParties := pageHeight - 110.0
	blockWidth := (pageWidth - 2*margin - 30) / 2

	// Seller block - left with subtle background
	fmt.Fprintf(&content, "%.3f %.3f %.3f rg\n", lightBgR, lightBgG, lightBgB)
	fmt.Fprintf(&content, "%.2f %.2f %.2f 85 re f\n", margin-10, yParties-70, blockWidth+20)

	writeTextColored(&content, "VENDEUR", margin, yParties, 11.0, primaryR, primaryG, primaryB)
	sellerName := req.Seller.Name
	if req.AddEISuffix {
		sellerName = req.Seller.Name + ", EI"
	}
	writeTextColored(&content, sellerName, margin, yParties-18, 10.0, 0.2, 0.2, 0.2)
	writeTextColored(&content, req.Seller.Address, margin, yParties-33, 9.0, grayR, grayG, grayB)
	writeTextColored(&content, fmt.Sprintf("%s %s", req.Seller.ZipCode, req.Seller.City), margin, yParties-46, 9.0, grayR, grayG, grayB)
	writeTextColored(&content, fmt.Sprintf("SIRET: %s", req.Seller.Siret), margin, yParties-59, 9.0, grayR, grayG, grayB)

	// Buyer block - right with subtle background
	buyerX := pageWidth/2.0 + 15.0
	fmt.Fprintf(&content, "%.3f %.3f %.3f rg\n", lightBgR, lightBgG, lightBgB)
	fmt.Fprintf(&content, "%.2f %.2f %.2f 85 re f\n", buyerX-10, yParties-70, blockWidth+20)

	writeTextColored(&content, "CLIENT", buyerX, yParties, 11.0, primaryR, primaryG, primaryB)
	writeTextColored(&content, req.Buyer.Name, buyerX, yParties-18, 10.0, 0.2, 0.2, 0.2)
	writeTextColored(&content, req.Buyer.Address, buyerX, yParties-33, 9.0, grayR, grayG, grayB)
	writeTextColored(&content, fmt.Sprintf("%s %s", req.Buyer.ZipCode, req.Buyer.City), buyerX, yParties-46, 9.0, grayR, grayG, grayB)
	writeTextColored(&content, fmt.Sprintf("SIRET: %s", req.Buyer.Siret), buyerX, yParties-59, 9.0, grayR, grayG, grayB)

	// ========================================================================
	// Table
	// ========================================================================
	tableTop := pageHeight - 230.0
	colDesc := margin
	colQty := margin + 280.0
	colPrice := margin + 350.0
	colTotal := margin + 440.0
	rowHeight := 22.0

	// Table header background
	fmt.Fprintf(&content, "%.3f %.3f %.3f rg\n", primaryR, primaryG, primaryB)
	fmt.Fprintf(&content, "%.2f %.2f %.2f %.2f re f\n", margin-10, tableTop-5, pageWidth-2*margin+20, 25.0)

	// Table header text in white
	writeTextColored(&content, "Description", colDesc, tableTop+3, 10.0, 1, 1, 1)
	writeTextColored(&content, "Qté", colQty, tableTop+3, 10.0, 1, 1, 1)
	writeTextColored(&content, "Prix unit.", colPrice, tableTop+3, 10.0, 1, 1, 1)
	writeTextColored(&content, "Total HT", colTotal, tableTop+3, 10.0, 1, 1, 1)

	// Table rows with alternating backgrounds
	y := tableTop - 25.0
	for i, line := range req.Lines {
		lineAmount := line.Quantity * line.UnitPrice

		// Alternating row background
		if i%2 == 0 {
			fmt.Fprintf(&content, "%.3f %.3f %.3f rg\n", lightBgR, lightBgG, lightBgB)
			fmt.Fprintf(&content, "%.2f %.2f %.2f %.2f re f\n", margin-10, y-5, pageWidth-2*margin+20, rowHeight)
		}

		// Truncate description if too long
		desc := line.Description
		if len(desc) > 45 {
			desc = desc[:42] + "..."
		}

		writeTextColored(&content, desc, colDesc, y+3, 10.0, 0.2, 0.2, 0.2)
		writeTextColored(&content, fmt.Sprintf("%.2f", line.Quantity), colQty, y+3, 10.0, 0.2, 0.2, 0.2)
		writeTextColored(&content, fmt.Sprintf("%.2f EUR", line.UnitPrice), colPrice, y+3, 10.0, 0.2, 0.2, 0.2)
		writeTextColored(&content, fmt.Sprintf("%.2f EUR", lineAmount), colTotal, y+3, 10.0, 0.2, 0.2, 0.2)

		y -= rowHeight
	}

	// Bottom line of table
	fmt.Fprintf(&content, "%.3f %.3f %.3f RG\n", primaryR, primaryG, primaryB)
	fmt.Fprintf(&content, "0.5 w\n")
	fmt.Fprintf(&content, "%.2f %.2f m %.2f %.2f l S\n", margin-10, y+rowHeight-5, pageWidth-margin+10, y+rowHeight-5)

	// ========================================================================
	// Totals box
	// ========================================================================
	totalsBoxX := colPrice - 40
	totalsBoxY := y - 85
	totalsBoxW := pageWidth - margin - totalsBoxX + 25
	totalsBoxH := 80.0

	// Totals background
	fmt.Fprintf(&content, "%.3f %.3f %.3f rg\n", lightBgR, lightBgG, lightBgB)
	fmt.Fprintf(&content, "%.2f %.2f %.2f %.2f re f\n", totalsBoxX, totalsBoxY, totalsBoxW, totalsBoxH)

	// Totals border
	fmt.Fprintf(&content, "%.3f %.3f %.3f RG\n", primaryR, primaryG, primaryB)
	fmt.Fprintf(&content, "1 w\n")
	fmt.Fprintf(&content, "%.2f %.2f %.2f %.2f re S\n", totalsBoxX, totalsBoxY, totalsBoxW, totalsBoxH)

	// Totals content
	totalsLabelX := colPrice - 20
	totalsValueX := colTotal
	totalsY := totalsBoxY + totalsBoxH - 20

	writeTextColored(&content, "Total HT:", totalsLabelX, totalsY, 10.0, 0.2, 0.2, 0.2)
	writeTextColored(&content, fmt.Sprintf("%s EUR", lineTotal), totalsValueX, totalsY, 10.0, 0.2, 0.2, 0.2)

	writeTextColored(&content, fmt.Sprintf("TVA (%s%%):", vatRate), totalsLabelX, totalsY-18, 10.0, 0.2, 0.2, 0.2)
	writeTextColored(&content, fmt.Sprintf("%s EUR", taxTotal), totalsValueX, totalsY-18, 10.0, 0.2, 0.2, 0.2)

	// Grand total highlight
	fmt.Fprintf(&content, "%.3f %.3f %.3f rg\n", primaryR, primaryG, primaryB)
	fmt.Fprintf(&content, "%.2f %.2f %.2f 22 re f\n", totalsBoxX, totalsBoxY, totalsBoxW)
	writeTextColored(&content, "Total TTC:", totalsLabelX, totalsBoxY+6, 11.0, 1, 1, 1)
	writeTextColored(&content, fmt.Sprintf("%s EUR", grandTotal), totalsValueX, totalsBoxY+6, 11.0, 1, 1, 1)

	// ========================================================================
	// Legal mentions
	// ========================================================================
	mentionsY := 110.0

	// Small accent line
	fmt.Fprintf(&content, "%.3f %.3f %.3f RG\n", accentR, accentG, accentB)
	fmt.Fprintf(&content, "2 w\n")
	fmt.Fprintf(&content, "%.2f %.2f m %.2f %.2f l S\n", margin, mentionsY+15, margin+40, mentionsY+15)
	fmt.Fprintf(&content, "1 w\n")

	writeTextColored(&content, "Mentions légales", margin, mentionsY, 9.0, primaryR, primaryG, primaryB)
	writeTextColored(&content, vatText, margin, mentionsY-14, 8.0, grayR, grayG, grayB)

	if req.CustomMentions != "" {
		cmY := mentionsY - 28.0
		for _, line := range strings.Split(req.CustomMentions, "\n") {
			writeTextColored(&content, line, margin, cmY, 8.0, grayR, grayG, grayB)
			cmY -= 11.0
		}
	}

	// ========================================================================
	// Footer
	// ========================================================================
	fmt.Fprintf(&content, "%.3f %.3f %.3f rg\n", lightBgR, lightBgG, lightBgB)
	fmt.Fprintf(&content, "0 0 %.2f 35 re f\n", pageWidth)
	writeTextColored(&content, "Document généré conformément à la norme Factur-X 1.0 (Profil BASIC)", margin, 14, 7.0, grayR, grayG, grayB)

	// End graphics state
	content.WriteString("Q\n")

	return content.Bytes()
}

// writeText writes text at position in black.
func writeText(content *bytes.Buffer, text string, x, y, size float64) {
	writeTextColored(content, text, x, y, size, 0, 0, 0)
}

// writeTextColored writes text at position with specified RGB color (0-1 range).
func writeTextColored(content *bytes.Buffer, text string, x, y, size, r, g, b float64) {
	encoded := encodeWinAnsi(text)
	content.WriteString("BT\n")
	fmt.Fprintf(content, "%.3f %.3f %.3f rg\n", r, g, b)
	fmt.Fprintf(content, "/F1 %.0f Tf\n", size)
	fmt.Fprintf(content, "%.2f %.2f Td\n", x, y)
	fmt.Fprintf(content, "(%s) Tj\n", encoded)
	content.WriteString("ET\n")
}

// encodeWinAnsi encodes text to WinAnsiEncoding and escapes for PDF string.
// Uses octal escapes for non-ASCII characters.
func encodeWinAnsi(s string) string {
	var result strings.Builder
	result.Grow(len(s) * 2)

	for _, c := range s {
		switch c {
		case '(':
			result.WriteString("\\(")
		case ')':
			result.WriteString("\\)")
		case '\\':
			result.WriteString("\\\\")
		case '\n':
			result.WriteString("\\n")
		case '\r':
			result.WriteString("\\r")
		case '\t':
			result.WriteString("\\t")
		// Map common Unicode to WinAnsi using octal escapes
		case 'é':
			result.WriteString("\\351")
		case 'è':
			result.WriteString("\\350")
		case 'ê':
			result.WriteString("\\352")
		case 'ë':
			result.WriteString("\\353")
		case 'à':
			result.WriteString("\\340")
		case 'â':
			result.WriteString("\\342")
		case 'ä':
			result.WriteString("\\344")
		case 'ù':
			result.WriteString("\\371")
		case 'û':
			result.WriteString("\\373")
		case 'ü':
			result.WriteString("\\374")
		case 'ô':
			result.WriteString("\\364")
		case 'ö':
			result.WriteString("\\366")
		case 'î':
			result.WriteString("\\356")
		case 'ï':
			result.WriteString("\\357")
		case 'ç':
			result.WriteString("\\347")
		case 'œ':
			result.WriteString("oe") // No direct mapping in WinAnsi
		case 'æ':
			result.WriteString("\\346")
		case '€':
			result.WriteString("\\200")
		case '°':
			result.WriteString("\\260")
		case '²':
			result.WriteString("\\262")
		case '³':
			result.WriteString("\\263")
		case 'É':
			result.WriteString("\\311")
		case 'È':
			result.WriteString("\\310")
		case 'Ê':
			result.WriteString("\\312")
		case 'À':
			result.WriteString("\\300")
		case 'Ç':
			result.WriteString("\\307")
		case 'Ô':
			result.WriteString("\\324")
		case 'Ù':
			result.WriteString("\\331")
		case 'Û':
			result.WriteString("\\333")
		case 'Î':
			result.WriteString("\\316")
		case 'Ï':
			result.WriteString("\\317")
		default:
			if c >= 32 && c < 127 {
				result.WriteRune(c)
			} else {
				// Other characters: replace with ?
				result.WriteByte('?')
			}
		}
	}
	return result.String()
}
