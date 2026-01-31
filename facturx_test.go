package facturx

import (
	"bytes"
	"strings"
	"testing"
)

func sampleRequest() InvoiceRequest {
	return InvoiceRequest{
		Number: "FA-2024-001",
		Date:   "20240115",
		Seller: Contact{
			Name:        "ACME Corp",
			Address:     "123 Rue de Paris",
			ZipCode:     "75001",
			City:        "Paris",
			CountryCode: "FR",
			Siret:       "12345678900006", // Valid Luhn checksum
			VatNumber:   "FR12345678901",
		},
		Buyer: Contact{
			Name:        "Client SA",
			Address:     "456 Avenue des Champs",
			ZipCode:     "69001",
			City:        "Lyon",
			CountryCode: "FR",
			Siret:       "98765432100006", // Valid Luhn checksum
			VatNumber:   "FR98765432109",
		},
		Lines: []InvoiceLine{
			{
				Description: "Prestation de conseil",
				Quantity:    10.0,
				UnitPrice:   100.0,
			},
		},
		Regime:      VatStandard(20.0),
		AddEISuffix: false,
	}
}

func TestGenerateFacturx(t *testing.T) {
	req := sampleRequest()
	pdf, err := Generate(req)
	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	if len(pdf) < 1000 {
		t.Error("PDF too small")
	}

	if !bytes.HasPrefix(pdf, []byte("%PDF-1.7")) {
		t.Error("PDF header missing")
	}
}

func TestValidationEmptyNumber(t *testing.T) {
	req := sampleRequest()
	req.Number = ""
	_, err := Generate(req)
	if err == nil {
		t.Error("Expected validation error for empty number")
	}
}

func TestValidationInvalidDate(t *testing.T) {
	req := sampleRequest()
	req.Date = "2024-01-15" // Wrong format
	_, err := Generate(req)
	if err == nil {
		t.Error("Expected validation error for invalid date")
	}
}

func TestValidationNoLines(t *testing.T) {
	req := sampleRequest()
	req.Lines = nil
	_, err := Generate(req)
	if err == nil {
		t.Error("Expected validation error for no lines")
	}
}

func TestValidationInvalidSiret(t *testing.T) {
	req := sampleRequest()
	req.Seller.Siret = "123"
	_, err := Generate(req)
	if err == nil {
		t.Error("Expected validation error for invalid SIRET")
	}
}

func TestValidationInvalidSiretLuhn(t *testing.T) {
	req := sampleRequest()
	req.Seller.Siret = "12345678901234" // 14 digits but invalid Luhn checksum
	_, err := Generate(req)
	if err == nil {
		t.Error("Expected validation error for invalid SIRET checksum")
	}
	ve, ok := err.(ValidationError)
	if !ok {
		t.Errorf("Expected ValidationError, got %T", err)
	}
	if ve.Field != "Seller.Siret" {
		t.Errorf("Expected field Seller.Siret, got %s", ve.Field)
	}
}

func TestSiretLuhnValidation(t *testing.T) {
	tests := []struct {
		siret string
		valid bool
	}{
		{"12345678900006", true},  // Valid checksum
		{"98765432100006", true},  // Valid checksum
		{"00000000000000", true},  // All zeros is valid
		{"12345678901234", false}, // Invalid checksum
		{"11111111111111", false}, // Invalid checksum
	}

	for _, tt := range tests {
		result := validateSiretLuhn(tt.siret)
		if result != tt.valid {
			t.Errorf("validateSiretLuhn(%s) = %v, want %v", tt.siret, result, tt.valid)
		}
	}
}

func TestFranchiseAuto(t *testing.T) {
	req := sampleRequest()
	req.Regime = VatFranchiseAuto()
	pdf, err := Generate(req)
	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}
	if len(pdf) < 1000 {
		t.Error("PDF too small")
	}
}

func TestEISuffix(t *testing.T) {
	req := sampleRequest()
	req.AddEISuffix = true
	xml, err := GenerateXMLOnly(&req)
	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}
	if !strings.Contains(xml, "Entrepreneur Individuel") {
		t.Error("EI suffix not found in XML")
	}
}

func TestXMLGeneration(t *testing.T) {
	req := sampleRequest()
	xml, err := GenerateXMLOnly(&req)
	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	// Check required elements
	checks := []string{
		"CrossIndustryInvoice",
		profileURN,
		"<ram:ID>FA-2024-001</ram:ID>",
		"<ram:TypeCode>380</ram:TypeCode>",
		"<ram:InvoiceCurrencyCode>EUR</ram:InvoiceCurrencyCode>",
	}

	for _, check := range checks {
		if !strings.Contains(xml, check) {
			t.Errorf("XML missing: %s", check)
		}
	}
}

func TestXMLCalculations(t *testing.T) {
	req := sampleRequest()
	req.Lines = []InvoiceLine{
		{Description: "Service 1", Quantity: 10, UnitPrice: 100},
		{Description: "Service 2", Quantity: 2, UnitPrice: 500},
	}
	xml, err := GenerateXMLOnly(&req)
	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	// Line total: 10*100 + 2*500 = 2000
	if !strings.Contains(xml, "<ram:LineTotalAmount>2000.00</ram:LineTotalAmount>") {
		t.Error("Incorrect line total")
	}

	// Tax: 2000 * 20% = 400
	if !strings.Contains(xml, `<ram:TaxTotalAmount currencyID="EUR">400.00</ram:TaxTotalAmount>`) {
		t.Error("Incorrect tax total")
	}

	// Grand total: 2000 + 400 = 2400
	if !strings.Contains(xml, "<ram:GrandTotalAmount>2400.00</ram:GrandTotalAmount>") {
		t.Error("Incorrect grand total")
	}
}

func TestXMLEscaping(t *testing.T) {
	req := sampleRequest()
	req.Lines[0].Description = "Test <>&\"' special chars"
	xml, err := GenerateXMLOnly(&req)
	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	if !strings.Contains(xml, "Test &lt;&gt;&amp;&quot;&apos; special chars") {
		t.Error("XML escaping failed")
	}
}

func TestPDFGeneration(t *testing.T) {
	req := sampleRequest()
	pdf, err := Generate(req)
	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	// Check PDF header
	if !bytes.HasPrefix(pdf, []byte("%PDF-1.7")) {
		t.Error("PDF header missing")
	}
	if !bytes.HasSuffix(pdf, []byte("%%EOF\n")) {
		t.Error("PDF footer missing")
	}

	// Check for required elements
	pdfStr := string(pdf)
	checks := []string{
		"/Type /Catalog",
		"factur-x.xml",
		"/AFRelationship /Data",
	}
	for _, check := range checks {
		if !strings.Contains(pdfStr, check) {
			t.Errorf("PDF missing: %s", check)
		}
	}
}

func TestWinAnsiEncoding(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Test", "Test"},
		{"(test)", "\\(test\\)"},
		{"été", "\\351t\\351"},
		{"€100", "\\200100"},
	}

	for _, tt := range tests {
		result := encodeWinAnsi(tt.input)
		if result != tt.expected {
			t.Errorf("encodeWinAnsi(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFontMetrics(t *testing.T) {
	metrics := getFontMetrics()

	// Liberation Sans should have 2048 units per em
	if metrics.unitsPerEM != 2048 {
		t.Errorf("Unexpected unitsPerEM: %d", metrics.unitsPerEM)
	}

	// Space should have a width
	if metrics.charWidth(' ') == 0 {
		t.Error("Space width is 0")
	}

	// 'M' should be wider than 'i'
	if metrics.charWidth('M') <= metrics.charWidth('i') {
		t.Error("'M' should be wider than 'i'")
	}
}

func BenchmarkGenerate(b *testing.B) {
	req := sampleRequest()
	for i := 0; i < b.N; i++ {
		_, err := Generate(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateXML(b *testing.B) {
	req := sampleRequest()
	for i := 0; i < b.N; i++ {
		_, err := GenerateXMLOnly(&req)
		if err != nil {
			b.Fatal(err)
		}
	}
}
