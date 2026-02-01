// Package facturx generates Factur-X 1.0 (BASIC profile) PDF/A-3 invoices.
//
// A zero-dependency Go library for generating electronic invoices conforming to:
//   - EN 16931-1 semantic model
//   - UN/CEFACT CII D16B syntax
//   - PDF/A-3 (ISO 19005-3) hybrid format
//   - Factur-X 1.0 BASIC profile
//
// Example:
//
//	req := facturx.InvoiceRequest{
//	    Number: "FA-2024-001",
//	    Date:   "20240115",
//	    Seller: facturx.Contact{
//	        Name:        "ACME Corp",
//	        Address:     "123 Rue de Paris",
//	        ZipCode:     "75001",
//	        City:        "Paris",
//	        CountryCode: "FR",
//	        Siret:       "12345678901234",
//	        VatNumber:   "FR12345678901",
//	    },
//	    Buyer: facturx.Contact{...},
//	    Lines: []facturx.InvoiceLine{
//	        {Description: "Prestation de conseil", Quantity: 10, UnitPrice: 100},
//	    },
//	    Regime: facturx.VatStandard(20.0),
//	}
//	pdfBytes, err := facturx.Generate(req)
package facturx

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

// VatRegime represents the VAT regime for the invoice.
type VatRegime struct {
	kind           vatKind
	rate           float64
	categoryCode   string
	exemptionCode  string
	exemptionText  string
}

type vatKind int

const (
	vatStandard vatKind = iota
	vatFranchiseAuto
	vatExemptHealth
)

// VatStandard creates a standard VAT regime with the given rate (e.g., 20.0 for 20%).
func VatStandard(rate float64) VatRegime {
	return VatRegime{
		kind:         vatStandard,
		rate:         rate,
		categoryCode: "S",
	}
}

// VatFranchiseAuto creates a VAT regime for franchise en base de TVA (Art. 293 B du CGI).
// Code: VATEX-FR-FRANCHISE
func VatFranchiseAuto() VatRegime {
	return VatRegime{
		kind:          vatFranchiseAuto,
		rate:          0,
		categoryCode:  "E",
		exemptionCode: "VATEX-FR-FRANCHISE",
		exemptionText: "TVA non applicable, art. 293 B du CGI",
	}
}

// VatExemptHealth creates a VAT regime for health activities exemption (Art. 261-4-1° du CGI).
// Code: VATEX-EU-O
func VatExemptHealth() VatRegime {
	return VatRegime{
		kind:          vatExemptHealth,
		rate:          0,
		categoryCode:  "E",
		exemptionCode: "VATEX-EU-O",
		exemptionText: "Exonération de TVA, art. 261-4-1° du CGI",
	}
}

// ProfessionalId represents a professional identifier (ADELI, RPPS, etc.).
type ProfessionalId struct {
	// Type of identifier (e.g., "ADELI", "RPPS").
	Type string
	// Value is the identifier value.
	Value string
}

// Contact represents contact information for seller or buyer.
type Contact struct {
	// Name is the full name (company or individual).
	Name string
	// Address is the street address.
	Address string
	// ZipCode is the postal code.
	ZipCode string
	// City is the city name.
	City string
	// CountryCode is the ISO 3166-1 alpha-2 country code (e.g., "FR").
	CountryCode string
	// Siret is the SIRET number (14 digits for French companies).
	Siret string
	// VatNumber is the VAT number (e.g., "FR12345678901"). Optional for exempt regimes.
	VatNumber string
	// ProfessionalIds contains professional identifiers (ADELI, RPPS, etc.).
	ProfessionalIds []ProfessionalId
}

// PaymentMethod represents the payment method for a paid invoice.
type PaymentMethod string

const (
	PaymentCash     PaymentMethod = "cash"
	PaymentCheck    PaymentMethod = "check"
	PaymentCard     PaymentMethod = "card"
	PaymentTransfer PaymentMethod = "transfer"
)

// PaymentMethodLabel returns the French label for the payment method.
func (m PaymentMethod) Label() string {
	switch m {
	case PaymentCash:
		return "espèces"
	case PaymentCheck:
		return "chèque"
	case PaymentCard:
		return "carte bancaire"
	case PaymentTransfer:
		return "virement"
	default:
		return string(m)
	}
}

// Payment contains payment information for paid invoices.
type Payment struct {
	// Date is the payment date in DD/MM/YYYY format.
	Date string
	// Method is the payment method.
	Method PaymentMethod
}

// InvoiceLine represents a single invoice line item.
type InvoiceLine struct {
	// Description of the product or service.
	Description string
	// Quantity (number of units).
	Quantity float64
	// UnitPrice in EUR (excluding tax).
	UnitPrice float64
	// Date is the service/delivery date in DD/MM/YYYY format (optional).
	Date string
}

// InvoiceRequest contains all data needed to generate an invoice.
type InvoiceRequest struct {
	// Number is the unique invoice identifier.
	Number string
	// Date in YYYYMMDD format (CII format code 102).
	Date string
	// Seller information.
	Seller Contact
	// Buyer information.
	Buyer Contact
	// Lines contains the invoice line items.
	Lines []InvoiceLine
	// Regime is the VAT regime.
	Regime VatRegime
	// AddEISuffix adds "Entrepreneur Individuel" suffix to seller name.
	AddEISuffix bool
	// CustomMentions is free text for legal mentions (can contain newlines).
	CustomMentions string
	// Payment contains payment info. If set, displays "Payée le [date] par [method]".
	Payment *Payment
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// validate checks the invoice request for errors.
func validate(req *InvoiceRequest) error {
	// Invoice number
	if strings.TrimSpace(req.Number) == "" {
		return ValidationError{Field: "Number", Message: "invoice number cannot be empty"}
	}

	// Date format: YYYYMMDD
	if len(req.Date) != 8 {
		return ValidationError{Field: "Date", Message: "date must be in YYYYMMDD format"}
	}
	for _, c := range req.Date {
		if !unicode.IsDigit(c) {
			return ValidationError{Field: "Date", Message: "date must contain only digits"}
		}
	}

	// Validate date values
	year := parseInt(req.Date[0:4])
	month := parseInt(req.Date[4:6])
	day := parseInt(req.Date[6:8])
	if year < 2000 || year > 2100 || month < 1 || month > 12 || day < 1 || day > 31 {
		return ValidationError{Field: "Date", Message: "invalid date values"}
	}

	// Lines
	if len(req.Lines) == 0 {
		return ValidationError{Field: "Lines", Message: "invoice must have at least one line"}
	}

	for i, line := range req.Lines {
		if line.Quantity <= 0 {
			return ValidationError{Field: fmt.Sprintf("Lines[%d].Quantity", i), Message: "quantity must be positive"}
		}
		if line.UnitPrice < 0 {
			return ValidationError{Field: fmt.Sprintf("Lines[%d].UnitPrice", i), Message: "unit price cannot be negative"}
		}
	}

	// Seller
	if strings.TrimSpace(req.Seller.Name) == "" {
		return ValidationError{Field: "Seller.Name", Message: "seller name cannot be empty"}
	}
	if err := validateContact(&req.Seller, "Seller", true); err != nil {
		return err
	}

	// Buyer (SIRET optional for B2C)
	if strings.TrimSpace(req.Buyer.Name) == "" {
		return ValidationError{Field: "Buyer.Name", Message: "buyer name cannot be empty"}
	}
	if err := validateContact(&req.Buyer, "Buyer", false); err != nil {
		return err
	}

	// VAT rate
	if req.Regime.kind == vatStandard && req.Regime.rate < 0 {
		return ValidationError{Field: "Regime", Message: "VAT rate cannot be negative"}
	}

	return nil
}

func validateContact(c *Contact, prefix string, requireSiret bool) error {
	// SIRET: 14 digits (optional for buyer in B2C)
	if c.Siret != "" || requireSiret {
		if len(c.Siret) != 14 {
			return ValidationError{Field: prefix + ".Siret", Message: "SIRET must be 14 digits"}
		}
		for _, ch := range c.Siret {
			if !unicode.IsDigit(ch) {
				return ValidationError{Field: prefix + ".Siret", Message: "SIRET must contain only digits"}
			}
		}
		if !validateSiretLuhn(c.Siret) {
			return ValidationError{Field: prefix + ".Siret", Message: "SIRET checksum invalid (Luhn)"}
		}
	}

	// Country code: 2 letters
	if len(c.CountryCode) != 2 {
		return ValidationError{Field: prefix + ".CountryCode", Message: "country code must be 2 letters"}
	}
	for _, ch := range c.CountryCode {
		if !unicode.IsLetter(ch) {
			return ValidationError{Field: prefix + ".CountryCode", Message: "country code must contain only letters"}
		}
	}

	return nil
}

// validateSiretLuhn validates a 14-digit SIRET using the Luhn algorithm.
// Assumes the input has already been validated as 14 numeric digits.
func validateSiretLuhn(siret string) bool {
	sum := 0
	for i := 0; i < 14; i++ {
		digit := int(siret[i] - '0')
		// Double every second digit (0-indexed: positions 1, 3, 5, 7, 9, 11, 13)
		if i%2 == 1 {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	return sum%10 == 0
}

func parseInt(s string) int {
	n := 0
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n
}

// Generate creates a Factur-X PDF/A-3 invoice.
//
// Returns the PDF file bytes on success, or an error on failure.
func Generate(req InvoiceRequest) ([]byte, error) {
	// Validate input
	if err := validate(&req); err != nil {
		return nil, err
	}

	// Generate CII XML
	xml := generateCIIXML(&req)

	// Generate PDF/A-3 with embedded XML
	pdf := generatePDF(&req, xml)

	return pdf, nil
}

// GenerateXMLOnly generates only the CII XML for an invoice (useful for debugging).
func GenerateXMLOnly(req *InvoiceRequest) (string, error) {
	if err := validate(req); err != nil {
		return "", err
	}
	return generateCIIXML(req), nil
}

// ErrValidation is returned when the invoice request fails validation.
var ErrValidation = errors.New("validation error")
