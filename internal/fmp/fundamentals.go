package fmp

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Per-call limits sized to the FMP starter plan.
const (
	financialsLimit = 5
	keyMetricsLimit = 4
	estimatesLimit  = 8
	upgradesLimit   = 20
	logoMaxBytes    = 1 << 20
)

// Service wraps the raw client with the shared cache and the fundamentals
// assembler used by the MCP tools.
type Service struct {
	client *Client
	cache  *memCache
	logf   func(format string, args ...interface{})
}

func NewService(apiKey string, cacheTTL time.Duration, logf func(string, ...interface{})) *Service {
	if cacheTTL <= 0 {
		cacheTTL = 24 * time.Hour
	}
	if logf == nil {
		logf = func(string, ...interface{}) {}
	}
	return &Service{
		client: NewClient(apiKey),
		cache:  newMemCache(cacheTTL),
		logf:   logf,
	}
}

// Search returns ranked ticker matches, cached per normalized query.
func (s *Service) Search(query string, limit int) ([]SearchResult, error) {
	if limit <= 0 || limit > 25 {
		limit = 10
	}
	key := fmt.Sprintf("search|%s|%d", strings.ToLower(strings.TrimSpace(query)), limit)
	if v, ok := s.cache.get(key); ok {
		s.logf("fmp: search cache hit for %q", query)
		return v.([]SearchResult), nil
	}
	res, err := s.client.Search(query, limit)
	if err != nil {
		return nil, err
	}
	s.cache.put(key, res)
	return res, nil
}

// Fundamentals is the assembled report bundle for one symbol.
type Fundamentals struct {
	Symbol      string                   `json:"symbol"`
	Period      string                   `json:"period"`
	GeneratedAt time.Time                `json:"generatedAt"`
	Profile     *StockProfile            `json:"profile"`
	Income      []IncomeStatement        `json:"income"`
	Balance     []BalanceSheet           `json:"balance"`
	CashFlow    []CashFlowStatement      `json:"cashFlow"`
	KeyMetrics  []KeyMetrics             `json:"keyMetrics"`
	RatiosTTM   *RatiosTTM               `json:"ratiosTtm"`
	RevenueProd []map[string]interface{} `json:"revenueByProduct"`
	RevenueGeo  []map[string]interface{} `json:"revenueByGeography"`
	Targets     *PriceTargetSummary      `json:"priceTargets"`
	Estimates   []AnalystEstimate        `json:"analystEstimates"`
	Upgrades    []UpgradeDowngrade       `json:"upgradesDowngrades"`
	History     []HistoricalPrice        `json:"history"`

	// LogoData is the company logo fetched from the profile image URL,
	// embedded so the HTML report stays a single self-contained file.
	// Not serialized into the json tool result.
	LogoData string `json:"-"`
}

// Fundamentals assembles the full bundle for a symbol, serving repeat
// requests from the shared cache. period is "annual" or "quarter".
func (s *Service) Fundamentals(symbol, period string) (*Fundamentals, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if period != "quarter" {
		period = "annual"
	}
	key := "fund|" + symbol + "|" + period
	if v, ok := s.cache.get(key); ok {
		s.logf("fmp: fundamentals cache hit for %s (%s)", symbol, period)
		return v.(*Fundamentals), nil
	}

	profile, err := s.client.GetProfile(symbol)
	if err != nil {
		return nil, fmt.Errorf("profile for %s: %w", symbol, err)
	}
	if profile == nil {
		return nil, fmt.Errorf("no data for symbol %s", symbol)
	}

	f := &Fundamentals{
		Symbol:      symbol,
		Period:      period,
		GeneratedAt: time.Now().UTC(),
		Profile:     profile,
	}
	// The statement and analyst sections degrade independently: a sparse
	// listing (fund, foreign ticker) still yields a useful report.
	if inc, err := s.client.GetIncomeStatement(symbol, period, financialsLimit); err == nil {
		f.Income = inc
	} else {
		s.logf("fmp: income %s: %v", symbol, err)
	}
	if bal, err := s.client.GetBalanceSheet(symbol, period, financialsLimit); err == nil {
		f.Balance = bal
	} else {
		s.logf("fmp: balance %s: %v", symbol, err)
	}
	if cf, err := s.client.GetCashFlowStatement(symbol, period, financialsLimit); err == nil {
		f.CashFlow = cf
	} else {
		s.logf("fmp: cashflow %s: %v", symbol, err)
	}
	if km, err := s.client.GetKeyMetrics(symbol, "annual", keyMetricsLimit); err == nil {
		f.KeyMetrics = km
	} else {
		s.logf("fmp: key metrics %s: %v", symbol, err)
	}
	if rt, err := s.client.GetRatiosTTM(symbol); err == nil {
		f.RatiosTTM = rt
	} else {
		s.logf("fmp: ratios ttm %s: %v", symbol, err)
	}
	if rp, err := s.client.GetRevenueProductSegmentation(symbol); err == nil {
		f.RevenueProd = rp
	} else {
		s.logf("fmp: revenue product %s: %v", symbol, err)
	}
	if rg, err := s.client.GetRevenueGeographicSegmentation(symbol); err == nil {
		f.RevenueGeo = rg
	} else {
		s.logf("fmp: revenue geo %s: %v", symbol, err)
	}
	if pt, err := s.client.GetPriceTargetSummary(symbol); err == nil {
		f.Targets = pt
	} else {
		s.logf("fmp: price targets %s: %v", symbol, err)
	}
	if est, err := s.client.GetAnalystEstimates(symbol, estimatesLimit); err == nil {
		f.Estimates = est
	} else {
		s.logf("fmp: estimates %s: %v", symbol, err)
	}
	if ud, err := s.client.GetUpgradesDowngrades(symbol, upgradesLimit); err == nil {
		f.Upgrades = ud
	} else {
		s.logf("fmp: upgrades %s: %v", symbol, err)
	}
	to := time.Now()
	from := to.AddDate(-1, 0, 0)
	if hist, err := s.client.GetHistoricalPrice(symbol,
		from.Format("2006-01-02"), to.Format("2006-01-02")); err == nil {
		f.History = hist
	} else {
		s.logf("fmp: history %s: %v", symbol, err)
	}
	f.LogoData = fetchLogoDataURI(profile.Image)

	s.cache.put(key, f)
	return f, nil
}

// MonthlyHistory downsamples the daily series to month-end closes for the
// compact json output.
func (f *Fundamentals) MonthlyHistory() []HistoricalPrice {
	var out []HistoricalPrice
	lastMonth := ""
	// History arrives newest-first from FMP; walk it as-is and keep the
	// first (latest) sample seen for each month.
	for _, h := range f.History {
		if len(h.Date) < 7 {
			continue
		}
		m := h.Date[:7]
		if m != lastMonth {
			out = append(out, h)
			lastMonth = m
		}
	}
	return out
}

// fetchLogoDataURI downloads the profile image and returns it as a data URI
// for embedding, or "" when unavailable.
func fetchLogoDataURI(imgURL string) string {
	if imgURL == "" || !strings.HasPrefix(imgURL, "https://") {
		return ""
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(imgURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, logoMaxBytes))
	if err != nil || len(data) == 0 {
		return ""
	}
	mime := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(mime, "image/") {
		return ""
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data)
}
