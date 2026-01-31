package facturx

import (
	"fmt"
	"strings"
)

// Factur-X BASIC profile URN (EN 16931 compliant)
const profileURN = "urn:cen.eu:en16931:2017#compliant#urn:factur-x.eu:1p0:basic"

// CII namespace declarations
const (
	nsRSM = "urn:un:unece:uncefact:data:standard:CrossIndustryInvoice:100"
	nsRAM = "urn:un:unece:uncefact:data:standard:ReusableAggregateBusinessInformationEntity:100"
	nsUDT = "urn:un:unece:uncefact:data:standard:UnqualifiedDataType:100"
	nsQDT = "urn:un:unece:uncefact:data:standard:QualifiedDataType:100"
)

// escapeXML escapes special characters for XML content.
func escapeXML(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, c := range s {
		switch c {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		case '"':
			b.WriteString("&quot;")
		case '\'':
			b.WriteString("&apos;")
		default:
			b.WriteRune(c)
		}
	}
	return b.String()
}

// fmtAmount formats a float with 2 decimal places.
func fmtAmount(value float64) string {
	return fmt.Sprintf("%.2f", value)
}

// fmtPrice formats a float with 4 decimal places (for unit prices).
func fmtPrice(value float64) string {
	return fmt.Sprintf("%.4f", value)
}

// fmtQuantity formats a float with 4 decimal places (for quantities).
func fmtQuantity(value float64) string {
	return fmt.Sprintf("%.4f", value)
}

// invoiceCalculation holds calculated invoice values.
type invoiceCalculation struct {
	lineTotal         float64
	taxBase           float64
	taxTotal          float64
	grandTotal        float64
	dueAmount         float64
	vatRate           float64
	vatCategoryCode   string
	vatExemptionCode  string
	vatExemptionText  string
}

// calculateInvoice computes invoice totals according to EN 16931 business rules.
func calculateInvoice(req *InvoiceRequest) invoiceCalculation {
	// BR-CO-10: Sum of line net amounts
	var lineTotal float64
	for _, line := range req.Lines {
		lineTotal += line.Quantity * line.UnitPrice
	}

	// Tax base is the sum of line amounts for simple invoices (no allowances/charges)
	taxBase := lineTotal

	// Determine VAT treatment
	vatRate := req.Regime.rate
	vatCategoryCode := req.Regime.categoryCode
	vatExemptionCode := req.Regime.exemptionCode
	vatExemptionText := req.Regime.exemptionText

	// BR-CO-14: VAT amount calculation
	taxTotal := taxBase * vatRate / 100.0

	// BR-CO-15: Grand total = tax base + tax
	grandTotal := taxBase + taxTotal

	// For simple invoices without prepayment, due = grand total
	dueAmount := grandTotal

	return invoiceCalculation{
		lineTotal:         lineTotal,
		taxBase:           taxBase,
		taxTotal:          taxTotal,
		grandTotal:        grandTotal,
		dueAmount:         dueAmount,
		vatRate:           vatRate,
		vatCategoryCode:   vatCategoryCode,
		vatExemptionCode:  vatExemptionCode,
		vatExemptionText:  vatExemptionText,
	}
}

// generateCIIXML generates the complete CII XML document.
func generateCIIXML(req *InvoiceRequest) string {
	calc := calculateInvoice(req)
	var xml strings.Builder
	xml.Grow(8192)

	// XML declaration
	xml.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	xml.WriteByte('\n')

	// Root element with namespaces
	fmt.Fprintf(&xml, `<rsm:CrossIndustryInvoice xmlns:rsm="%s" xmlns:ram="%s" xmlns:udt="%s" xmlns:qdt="%s">`,
		nsRSM, nsRAM, nsUDT, nsQDT)
	xml.WriteByte('\n')

	// ExchangedDocumentContext - identifies profile
	writeDocumentContext(&xml)

	// ExchangedDocument - invoice header
	writeExchangedDocument(&xml, req)

	// SupplyChainTradeTransaction - the main content
	writeSupplyChainTradeTransaction(&xml, req, &calc)

	xml.WriteString("</rsm:CrossIndustryInvoice>\n")

	return xml.String()
}

// writeDocumentContext writes the ExchangedDocumentContext element.
func writeDocumentContext(xml *strings.Builder) {
	xml.WriteString("  <rsm:ExchangedDocumentContext>\n")

	// Business process (optional but recommended)
	xml.WriteString("    <ram:BusinessProcessSpecifiedDocumentContextParameter>\n")
	xml.WriteString("      <ram:ID>A1</ram:ID>\n")
	xml.WriteString("    </ram:BusinessProcessSpecifiedDocumentContextParameter>\n")

	// Guideline - MUST be Factur-X BASIC
	xml.WriteString("    <ram:GuidelineSpecifiedDocumentContextParameter>\n")
	fmt.Fprintf(xml, "      <ram:ID>%s</ram:ID>\n", profileURN)
	xml.WriteString("    </ram:GuidelineSpecifiedDocumentContextParameter>\n")

	xml.WriteString("  </rsm:ExchangedDocumentContext>\n")
}

// writeExchangedDocument writes the ExchangedDocument element (invoice header).
func writeExchangedDocument(xml *strings.Builder, req *InvoiceRequest) {
	xml.WriteString("  <rsm:ExchangedDocument>\n")

	// Invoice number (BT-1)
	fmt.Fprintf(xml, "    <ram:ID>%s</ram:ID>\n", escapeXML(req.Number))

	// Type code: 380 = Commercial Invoice (BT-3)
	xml.WriteString("    <ram:TypeCode>380</ram:TypeCode>\n")

	// Issue date (BT-2) - format code 102 = YYYYMMDD
	xml.WriteString("    <ram:IssueDateTime>\n")
	fmt.Fprintf(xml, "      <udt:DateTimeString format=\"102\">%s</udt:DateTimeString>\n", escapeXML(req.Date))
	xml.WriteString("    </ram:IssueDateTime>\n")

	xml.WriteString("  </rsm:ExchangedDocument>\n")
}

// writeSupplyChainTradeTransaction writes the main transaction content.
func writeSupplyChainTradeTransaction(xml *strings.Builder, req *InvoiceRequest, calc *invoiceCalculation) {
	xml.WriteString("  <rsm:SupplyChainTradeTransaction>\n")

	// Line items
	for i, line := range req.Lines {
		writeLineItem(xml, &line, i+1, calc)
	}

	// Trade agreement (seller, buyer)
	writeApplicableHeaderTradeAgreement(xml, req)

	// Trade delivery
	writeApplicableHeaderTradeDelivery(xml, req.Date)

	// Trade settlement (payment, totals)
	writeApplicableHeaderTradeSettlement(xml, calc)

	xml.WriteString("  </rsm:SupplyChainTradeTransaction>\n")
}

// writeLineItem writes a single line item.
func writeLineItem(xml *strings.Builder, line *InvoiceLine, lineNum int, calc *invoiceCalculation) {
	lineAmount := line.Quantity * line.UnitPrice

	xml.WriteString("    <ram:IncludedSupplyChainTradeLineItem>\n")

	// Line ID (BT-126)
	xml.WriteString("      <ram:AssociatedDocumentLineDocument>\n")
	fmt.Fprintf(xml, "        <ram:LineID>%d</ram:LineID>\n", lineNum)
	xml.WriteString("      </ram:AssociatedDocumentLineDocument>\n")

	// Product information
	xml.WriteString("      <ram:SpecifiedTradeProduct>\n")
	fmt.Fprintf(xml, "        <ram:Name>%s</ram:Name>\n", escapeXML(line.Description))
	xml.WriteString("      </ram:SpecifiedTradeProduct>\n")

	// Line trade agreement (price)
	xml.WriteString("      <ram:SpecifiedLineTradeAgreement>\n")
	xml.WriteString("        <ram:NetPriceProductTradePrice>\n")
	fmt.Fprintf(xml, "          <ram:ChargeAmount>%s</ram:ChargeAmount>\n", fmtPrice(line.UnitPrice))
	xml.WriteString("        </ram:NetPriceProductTradePrice>\n")
	xml.WriteString("      </ram:SpecifiedLineTradeAgreement>\n")

	// Line trade delivery (quantity)
	xml.WriteString("      <ram:SpecifiedLineTradeDelivery>\n")
	fmt.Fprintf(xml, "        <ram:BilledQuantity unitCode=\"C62\">%s</ram:BilledQuantity>\n", fmtQuantity(line.Quantity))
	xml.WriteString("      </ram:SpecifiedLineTradeDelivery>\n")

	// Line trade settlement
	xml.WriteString("      <ram:SpecifiedLineTradeSettlement>\n")

	// Line VAT
	xml.WriteString("        <ram:ApplicableTradeTax>\n")
	xml.WriteString("          <ram:TypeCode>VAT</ram:TypeCode>\n")
	fmt.Fprintf(xml, "          <ram:CategoryCode>%s</ram:CategoryCode>\n", calc.vatCategoryCode)
	fmt.Fprintf(xml, "          <ram:RateApplicablePercent>%s</ram:RateApplicablePercent>\n", fmtAmount(calc.vatRate))
	xml.WriteString("        </ram:ApplicableTradeTax>\n")

	// Line net amount (BT-131)
	xml.WriteString("        <ram:SpecifiedTradeSettlementLineMonetarySummation>\n")
	fmt.Fprintf(xml, "          <ram:LineTotalAmount>%s</ram:LineTotalAmount>\n", fmtAmount(lineAmount))
	xml.WriteString("        </ram:SpecifiedTradeSettlementLineMonetarySummation>\n")

	xml.WriteString("      </ram:SpecifiedLineTradeSettlement>\n")

	xml.WriteString("    </ram:IncludedSupplyChainTradeLineItem>\n")
}

// writeApplicableHeaderTradeAgreement writes seller and buyer information.
func writeApplicableHeaderTradeAgreement(xml *strings.Builder, req *InvoiceRequest) {
	xml.WriteString("    <ram:ApplicableHeaderTradeAgreement>\n")

	// Seller (BG-4)
	writeTradeParty(xml, &req.Seller, "SellerTradeParty", req.AddEISuffix)

	// Buyer (BG-7)
	writeTradeParty(xml, &req.Buyer, "BuyerTradeParty", false)

	xml.WriteString("    </ram:ApplicableHeaderTradeAgreement>\n")
}

// writeTradeParty writes a trade party (seller or buyer).
func writeTradeParty(xml *strings.Builder, contact *Contact, elementName string, addEISuffix bool) {
	fmt.Fprintf(xml, "      <ram:%s>\n", elementName)

	// Name (BT-27 for seller, BT-44 for buyer)
	name := contact.Name
	if addEISuffix {
		name = contact.Name + ", Entrepreneur Individuel"
	}
	fmt.Fprintf(xml, "        <ram:Name>%s</ram:Name>\n", escapeXML(name))

	// Legal organization with SIRET
	xml.WriteString("        <ram:SpecifiedLegalOrganization>\n")
	fmt.Fprintf(xml, "          <ram:ID schemeID=\"0002\">%s</ram:ID>\n", escapeXML(contact.Siret))
	xml.WriteString("        </ram:SpecifiedLegalOrganization>\n")

	// Postal address (BG-5 for seller, BG-8 for buyer)
	xml.WriteString("        <ram:PostalTradeAddress>\n")
	fmt.Fprintf(xml, "          <ram:PostcodeCode>%s</ram:PostcodeCode>\n", escapeXML(contact.ZipCode))
	fmt.Fprintf(xml, "          <ram:LineOne>%s</ram:LineOne>\n", escapeXML(contact.Address))
	fmt.Fprintf(xml, "          <ram:CityName>%s</ram:CityName>\n", escapeXML(contact.City))
	fmt.Fprintf(xml, "          <ram:CountryID>%s</ram:CountryID>\n", escapeXML(contact.CountryCode))
	xml.WriteString("        </ram:PostalTradeAddress>\n")

	// Tax registration (VAT number) if present
	if contact.VatNumber != "" {
		xml.WriteString("        <ram:SpecifiedTaxRegistration>\n")
		fmt.Fprintf(xml, "          <ram:ID schemeID=\"VA\">%s</ram:ID>\n", escapeXML(contact.VatNumber))
		xml.WriteString("        </ram:SpecifiedTaxRegistration>\n")
	}

	fmt.Fprintf(xml, "      </ram:%s>\n", elementName)
}

// writeApplicableHeaderTradeDelivery writes delivery information.
func writeApplicableHeaderTradeDelivery(xml *strings.Builder, date string) {
	xml.WriteString("    <ram:ApplicableHeaderTradeDelivery>\n")

	// Actual delivery date (BT-72) - using invoice date as default
	xml.WriteString("      <ram:ActualDeliverySupplyChainEvent>\n")
	xml.WriteString("        <ram:OccurrenceDateTime>\n")
	fmt.Fprintf(xml, "          <udt:DateTimeString format=\"102\">%s</udt:DateTimeString>\n", date)
	xml.WriteString("        </ram:OccurrenceDateTime>\n")
	xml.WriteString("      </ram:ActualDeliverySupplyChainEvent>\n")

	xml.WriteString("    </ram:ApplicableHeaderTradeDelivery>\n")
}

// writeApplicableHeaderTradeSettlement writes payment and totals.
func writeApplicableHeaderTradeSettlement(xml *strings.Builder, calc *invoiceCalculation) {
	xml.WriteString("    <ram:ApplicableHeaderTradeSettlement>\n")

	// Invoice currency (BT-5)
	xml.WriteString("      <ram:InvoiceCurrencyCode>EUR</ram:InvoiceCurrencyCode>\n")

	// VAT breakdown (BG-23)
	xml.WriteString("      <ram:ApplicableTradeTax>\n")
	fmt.Fprintf(xml, "        <ram:CalculatedAmount>%s</ram:CalculatedAmount>\n", fmtAmount(calc.taxTotal))
	xml.WriteString("        <ram:TypeCode>VAT</ram:TypeCode>\n")

	// Exemption reason if applicable
	if calc.vatExemptionText != "" {
		fmt.Fprintf(xml, "        <ram:ExemptionReason>%s</ram:ExemptionReason>\n", escapeXML(calc.vatExemptionText))
	}

	fmt.Fprintf(xml, "        <ram:BasisAmount>%s</ram:BasisAmount>\n", fmtAmount(calc.taxBase))
	fmt.Fprintf(xml, "        <ram:CategoryCode>%s</ram:CategoryCode>\n", calc.vatCategoryCode)

	// Exemption reason code if applicable
	if calc.vatExemptionCode != "" {
		fmt.Fprintf(xml, "        <ram:ExemptionReasonCode>%s</ram:ExemptionReasonCode>\n", calc.vatExemptionCode)
	}

	fmt.Fprintf(xml, "        <ram:RateApplicablePercent>%s</ram:RateApplicablePercent>\n", fmtAmount(calc.vatRate))
	xml.WriteString("      </ram:ApplicableTradeTax>\n")

	// Payment terms (BT-20) - required when DuePayableAmount > 0
	xml.WriteString("      <ram:SpecifiedTradePaymentTerms>\n")
	xml.WriteString("        <ram:Description>Paiement à réception de facture</ram:Description>\n")
	xml.WriteString("      </ram:SpecifiedTradePaymentTerms>\n")

	// Monetary summation (BG-22)
	xml.WriteString("      <ram:SpecifiedTradeSettlementHeaderMonetarySummation>\n")

	// Sum of line net amounts (BT-106)
	fmt.Fprintf(xml, "        <ram:LineTotalAmount>%s</ram:LineTotalAmount>\n", fmtAmount(calc.lineTotal))

	// Tax basis total (BT-109)
	fmt.Fprintf(xml, "        <ram:TaxBasisTotalAmount>%s</ram:TaxBasisTotalAmount>\n", fmtAmount(calc.taxBase))

	// Tax total (BT-110)
	fmt.Fprintf(xml, "        <ram:TaxTotalAmount currencyID=\"EUR\">%s</ram:TaxTotalAmount>\n", fmtAmount(calc.taxTotal))

	// Grand total (BT-112)
	fmt.Fprintf(xml, "        <ram:GrandTotalAmount>%s</ram:GrandTotalAmount>\n", fmtAmount(calc.grandTotal))

	// Due payable amount (BT-115)
	fmt.Fprintf(xml, "        <ram:DuePayableAmount>%s</ram:DuePayableAmount>\n", fmtAmount(calc.dueAmount))

	xml.WriteString("      </ram:SpecifiedTradeSettlementHeaderMonetarySummation>\n")

	xml.WriteString("    </ram:ApplicableHeaderTradeSettlement>\n")
}
