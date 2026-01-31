// Package main provides a simple HTTP server for the Factur-X web interface.
// Uses the Go library directly for PDF generation.
// The frontend is embedded in the binary for single-binary deployment.
package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/audrenbdb/facturx"
)

//go:embed dist/*
var distFS embed.FS

// Rate limiter: 10 requests per hour per IP
const (
	rateLimitWindow   = time.Hour
	rateLimitRequests = 10
)

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
}

var limiter = &rateLimiter{
	requests: make(map[string][]time.Time),
}

func (rl *rateLimiter) allow(ip string) (bool, int, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rateLimitWindow)

	// Clean old requests
	var recent []time.Time
	for _, t := range rl.requests[ip] {
		if t.After(windowStart) {
			recent = append(recent, t)
		}
	}
	rl.requests[ip] = recent

	remaining := rateLimitRequests - len(recent)
	if remaining <= 0 {
		// Find when the oldest request will expire
		resetIn := rl.requests[ip][0].Add(rateLimitWindow).Sub(now)
		return false, 0, resetIn
	}

	// Allow and record the request
	rl.requests[ip] = append(rl.requests[ip], now)
	return true, remaining - 1, rateLimitWindow
}

type GenerateRequest struct {
	Number       string      `json:"number"`
	Date         string      `json:"date"`
	Seller       ContactJSON `json:"seller"`
	Buyer        ContactJSON `json:"buyer"`
	Lines        []LineJSON  `json:"lines"`
	PaymentTerms PaymentJSON `json:"paymentTerms"`
	Note         string      `json:"note"`
}

type ContactJSON struct {
	Name       string `json:"name"`
	SIRET      string `json:"siret"`
	VATNumber  string `json:"vatNumber"`
	Street     string `json:"street"`
	PostalCode string `json:"postalCode"`
	City       string `json:"city"`
	Email      string `json:"email"`
}

type LineJSON struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
	VATRegime   int     `json:"vatRegime"`
}

type PaymentJSON struct {
	DueDate string `json:"dueDate"`
	IBAN    string `json:"iban"`
	BIC     string `json:"bic"`
	Note    string `json:"note"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func main() {
	// API routes
	http.HandleFunc("/api/generate", handleGenerate)
	http.HandleFunc("/api/health", handleHealth)

	// Serve embedded frontend
	distContent, err := fs.Sub(distFS, "dist")
	if err != nil {
		log.Fatal(err)
	}
	fileServer := http.FileServer(http.FS(distContent))

	// SPA handler: serve static files, fallback to index.html
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Try to serve the file directly
		if path != "/" {
			// Check if file exists
			if f, err := distContent.Open(strings.TrimPrefix(path, "/")); err == nil {
				f.Close()
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// Fallback to index.html for SPA routing
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})

	addr := ":9473"
	log.Printf("Factur-X server starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "backend": "go-native"})
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Apply rate limiting
	ip := getClientIP(r)
	allowed, remaining, resetIn := limiter.allow(ip)

	w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rateLimitRequests))
	w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
	w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", int(resetIn.Seconds())))

	if !allowed {
		log.Printf("Rate limit exceeded for IP %s", ip)
		sendError(w, fmt.Sprintf("Rate limit dépassé. Limite: %d factures par heure. Réessayez dans %d minutes.", rateLimitRequests, int(resetIn.Minutes())+1), http.StatusTooManyRequests)
		return
	}

	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Format de requête invalide: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Convert to facturx library format
	invoiceReq, err := convertToFacturxFormat(req)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate PDF using Go library directly
	pdfData, err := facturx.Generate(invoiceReq)
	if err != nil {
		sendError(w, "Erreur de génération: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Generated invoice %s (%d bytes)", req.Number, len(pdfData))

	// Send PDF response
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="facture-%s.pdf"`, req.Number))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(pdfData)))
	w.Write(pdfData)
}

func convertToFacturxFormat(req GenerateRequest) (facturx.InvoiceRequest, error) {
	// Convert date from YYYY-MM-DD to YYYYMMDD
	date := strings.ReplaceAll(req.Date, "-", "")
	if len(date) != 8 {
		return facturx.InvoiceRequest{}, fmt.Errorf("format de date invalide")
	}

	// Determine VAT regime from lines
	var regime facturx.VatRegime
	if len(req.Lines) > 0 {
		firstRegime := req.Lines[0].VATRegime
		switch firstRegime {
		case 4: // franchise_auto
			regime = facturx.VatFranchiseAuto()
		case 5: // exempt_health
			regime = facturx.VatExemptHealth()
		default:
			// Standard VAT - get rate from regime code
			rate := getVATRate(firstRegime)
			regime = facturx.VatStandard(rate)
		}
	} else {
		regime = facturx.VatStandard(20.0)
	}

	// Build custom mentions from payment info
	var mentions []string
	if req.PaymentTerms.Note != "" {
		mentions = append(mentions, req.PaymentTerms.Note)
	}
	if req.PaymentTerms.IBAN != "" {
		mentions = append(mentions, fmt.Sprintf("IBAN: %s", req.PaymentTerms.IBAN))
	}
	if req.Note != "" {
		mentions = append(mentions, req.Note)
	}

	invoiceReq := facturx.InvoiceRequest{
		Number: req.Number,
		Date:   date,
		Seller: facturx.Contact{
			Name:        req.Seller.Name,
			Address:     req.Seller.Street,
			ZipCode:     req.Seller.PostalCode,
			City:        req.Seller.City,
			CountryCode: "FR",
			Siret:       strings.ReplaceAll(req.Seller.SIRET, " ", ""),
			VatNumber:   req.Seller.VATNumber,
		},
		Buyer: facturx.Contact{
			Name:        req.Buyer.Name,
			Address:     req.Buyer.Street,
			ZipCode:     req.Buyer.PostalCode,
			City:        req.Buyer.City,
			CountryCode: "FR",
			Siret:       strings.ReplaceAll(req.Buyer.SIRET, " ", ""),
		},
		Regime:         regime,
		AddEISuffix:    false,
		CustomMentions: strings.Join(mentions, "\n"),
	}

	// Convert lines
	for _, line := range req.Lines {
		invoiceReq.Lines = append(invoiceReq.Lines, facturx.InvoiceLine{
			Description: line.Description,
			Quantity:    line.Quantity,
			UnitPrice:   line.UnitPrice,
		})
	}

	return invoiceReq, nil
}

func getVATRate(regime int) float64 {
	switch regime {
	case 0: // standard
		return 20.0
	case 1: // reduced
		return 10.0
	case 2: // super_reduced
		return 5.5
	case 3: // minimal
		return 2.1
	case 4, 5, 6, 7: // exempt
		return 0.0
	default:
		return 20.0
	}
}

func sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Message: message})
}
