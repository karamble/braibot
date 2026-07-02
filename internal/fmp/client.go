package fmp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const stableBaseURL = "https://financialmodelingprep.com/stable"

type Client struct {
	apiKey     string
	httpClient *http.Client
}

type SearchResult struct {
	Symbol            string `json:"symbol"`
	Name              string `json:"name"`
	Currency          string `json:"currency"`
	StockExchange     string `json:"exchangeFullName"`
	ExchangeShortName string `json:"exchange"`
}

type StockProfile struct {
	Symbol            string  `json:"symbol"`
	CompanyName       string  `json:"companyName"`
	Currency          string  `json:"currency"`
	Exchange          string  `json:"exchange"`
	ExchangeShortName string  `json:"exchangeShortName"`
	Price             float64 `json:"price"`
	MarketCap         float64 `json:"marketCap"`
	Beta              float64 `json:"beta"`
	LastDividend      float64 `json:"lastDividend"`
	Range             string  `json:"range"`
	Change            float64 `json:"change"`
	ChangePercentage  float64 `json:"changePercentage"`
	Volume            float64 `json:"volume"`
	AverageVolume     float64 `json:"averageVolume"`
	Industry          string  `json:"industry"`
	Sector            string  `json:"sector"`
	Country           string  `json:"country"`
	Description       string  `json:"description"`
	CEO               string  `json:"ceo"`
	Website           string  `json:"website"`
	Image             string  `json:"image"`
	IPODate           string  `json:"ipoDate"`
	FullTimeEmployees string  `json:"fullTimeEmployees"`
	Phone             string  `json:"phone"`
	Address           string  `json:"address"`
	City              string  `json:"city"`
	State             string  `json:"state"`
	Zip               string  `json:"zip"`
	IsETF             bool    `json:"isEtf"`
	IsActivelyTrading bool    `json:"isActivelyTrading"`
}

// IncomeStatement represents an income statement from FMP API
type IncomeStatement struct {
	Date                           string  `json:"date"`
	Symbol                         string  `json:"symbol"`
	Period                         string  `json:"period"`
	Revenue                        int64   `json:"revenue"`
	CostOfRevenue                  int64   `json:"costOfRevenue"`
	GrossProfit                    int64   `json:"grossProfit"`
	GrossProfitRatio               float64 `json:"grossProfitRatio"`
	ResearchAndDevelopmentExpenses int64   `json:"researchAndDevelopmentExpenses"`
	SellingGeneralAndAdminExpenses int64   `json:"sellingGeneralAndAdministrativeExpenses"`
	OperatingExpenses              int64   `json:"operatingExpenses"`
	OperatingIncome                int64   `json:"operatingIncome"`
	OperatingIncomeRatio           float64 `json:"operatingIncomeRatio"`
	InterestExpense                int64   `json:"interestExpense"`
	IncomeBeforeTax                int64   `json:"incomeBeforeTax"`
	IncomeTaxExpense               int64   `json:"incomeTaxExpense"`
	NetIncome                      int64   `json:"netIncome"`
	NetIncomeRatio                 float64 `json:"netIncomeRatio"`
	EPS                            float64 `json:"eps"`
	EPSDiluted                     float64 `json:"epsdiluted"`
	WeightedAvgSharesOut           int64   `json:"weightedAverageShsOut"`
	WeightedAvgSharesOutDiluted    int64   `json:"weightedAverageShsOutDil"`
	EBITDA                         int64   `json:"ebitda"`
	CalendarYear                   string  `json:"calendarYear"`
}

// BalanceSheet represents a balance sheet from FMP API
type BalanceSheet struct {
	Date                        string `json:"date"`
	Symbol                      string `json:"symbol"`
	Period                      string `json:"period"`
	CashAndCashEquivalents      int64  `json:"cashAndCashEquivalents"`
	ShortTermInvestments        int64  `json:"shortTermInvestments"`
	CashAndShortTermInvestments int64  `json:"cashAndShortTermInvestments"`
	NetReceivables              int64  `json:"netReceivables"`
	Inventory                   int64  `json:"inventory"`
	TotalCurrentAssets          int64  `json:"totalCurrentAssets"`
	PropertyPlantEquipmentNet   int64  `json:"propertyPlantEquipmentNet"`
	Goodwill                    int64  `json:"goodwill"`
	IntangibleAssets            int64  `json:"intangibleAssets"`
	LongTermInvestments         int64  `json:"longTermInvestments"`
	TotalNonCurrentAssets       int64  `json:"totalNonCurrentAssets"`
	TotalAssets                 int64  `json:"totalAssets"`
	AccountPayables             int64  `json:"accountPayables"`
	ShortTermDebt               int64  `json:"shortTermDebt"`
	TotalCurrentLiabilities     int64  `json:"totalCurrentLiabilities"`
	LongTermDebt                int64  `json:"longTermDebt"`
	TotalNonCurrentLiabilities  int64  `json:"totalNonCurrentLiabilities"`
	TotalLiabilities            int64  `json:"totalLiabilities"`
	TotalStockholdersEquity     int64  `json:"totalStockholdersEquity"`
	TotalEquity                 int64  `json:"totalEquity"`
	TotalDebt                   int64  `json:"totalDebt"`
	NetDebt                     int64  `json:"netDebt"`
	CalendarYear                string `json:"calendarYear"`
}

// CashFlowStatement represents a cash flow statement from FMP API
type CashFlowStatement struct {
	Date                                string `json:"date"`
	Symbol                              string `json:"symbol"`
	Period                              string `json:"period"`
	NetIncome                           int64  `json:"netIncome"`
	DepreciationAndAmortization         int64  `json:"depreciationAndAmortization"`
	StockBasedCompensation              int64  `json:"stockBasedCompensation"`
	ChangeInWorkingCapital              int64  `json:"changeInWorkingCapital"`
	OperatingCashFlow                   int64  `json:"operatingCashFlow"`
	CapitalExpenditure                  int64  `json:"capitalExpenditure"`
	InvestmentsInPropertyPlantEquipment int64  `json:"investmentsInPropertyPlantAndEquipment"`
	AcquisitionsNet                     int64  `json:"acquisitionsNet"`
	InvestingCashFlow                   int64  `json:"netCashUsedForInvestingActivites"`
	DebtRepayment                       int64  `json:"debtRepayment"`
	CommonStockRepurchased              int64  `json:"commonStockRepurchased"`
	DividendsPaid                       int64  `json:"dividendsPaid"`
	FinancingCashFlow                   int64  `json:"netCashUsedProvidedByFinancingActivities"`
	FreeCashFlow                        int64  `json:"freeCashFlow"`
	NetChangeInCash                     int64  `json:"netChangeInCash"`
	CashAtEndOfPeriod                   int64  `json:"cashAtEndOfPeriod"`
	CalendarYear                        string `json:"calendarYear"`
}

// KeyMetrics represents key financial metrics from FMP API
type KeyMetrics struct {
	Date                       string  `json:"date"`
	Symbol                     string  `json:"symbol"`
	Period                     string  `json:"period"`
	RevenuePerShare            float64 `json:"revenuePerShare"`
	NetIncomePerShare          float64 `json:"netIncomePerShare"`
	OperatingCashFlowPerShare  float64 `json:"operatingCashFlowPerShare"`
	FreeCashFlowPerShare       float64 `json:"freeCashFlowPerShare"`
	CashPerShare               float64 `json:"cashPerShare"`
	BookValuePerShare          float64 `json:"bookValuePerShare"`
	TangibleBookValuePerShare  float64 `json:"tangibleBookValuePerShare"`
	ShareholdersEquityPerShare float64 `json:"shareholdersEquityPerShare"`
	InterestDebtPerShare       float64 `json:"interestDebtPerShare"`
	MarketCap                  int64   `json:"marketCap"`
	EnterpriseValue            int64   `json:"enterpriseValue"`
	PERatio                    float64 `json:"peRatio"`
	PriceToSalesRatio          float64 `json:"priceToSalesRatio"`
	POCF                       float64 `json:"pocfratio"`
	PFCF                       float64 `json:"pfcfRatio"`
	PBRatio                    float64 `json:"pbRatio"`
	PTB                        float64 `json:"ptbRatio"`
	EVToSales                  float64 `json:"evToSales"`
	EVToEBITDA                 float64 `json:"enterpriseValueOverEBITDA"`
	EVToOperatingCashFlow      float64 `json:"evToOperatingCashFlow"`
	EVToFreeCashFlow           float64 `json:"evToFreeCashFlow"`
	EarningsYield              float64 `json:"earningsYield"`
	FreeCashFlowYield          float64 `json:"freeCashFlowYield"`
	DebtToEquity               float64 `json:"debtToEquity"`
	DebtToAssets               float64 `json:"debtToAssets"`
	NetDebtToEBITDA            float64 `json:"netDebtToEBITDA"`
	CurrentRatio               float64 `json:"currentRatio"`
	InterestCoverage           float64 `json:"interestCoverage"`
	IncomeQuality              float64 `json:"incomeQuality"`
	DividendYield              float64 `json:"dividendYield"`
	PayoutRatio                float64 `json:"payoutRatio"`
	ROE                        float64 `json:"roe"`
	ROIC                       float64 `json:"roic"`
	CalendarYear               string  `json:"calendarYear"`
}

// RatiosTTM represents trailing twelve months financial ratios
type RatiosTTM struct {
	Symbol                     string  `json:"symbol"`
	DividendYieldTTM           float64 `json:"dividendYielTTM"`
	DividendPerShareTTM        float64 `json:"dividendPerShareTTM"`
	PERatioTTM                 float64 `json:"peRatioTTM"`
	PEGRatioTTM                float64 `json:"pegRatioTTM"`
	PayoutRatioTTM             float64 `json:"payoutRatioTTM"`
	CurrentRatioTTM            float64 `json:"currentRatioTTM"`
	QuickRatioTTM              float64 `json:"quickRatioTTM"`
	CashRatioTTM               float64 `json:"cashRatioTTM"`
	GrossProfitMarginTTM       float64 `json:"grossProfitMarginTTM"`
	OperatingProfitMarginTTM   float64 `json:"operatingProfitMarginTTM"`
	NetProfitMarginTTM         float64 `json:"netProfitMarginTTM"`
	DebtRatioTTM               float64 `json:"debtRatioTTM"`
	DebtEquityRatioTTM         float64 `json:"debtEquityRatioTTM"`
	LongTermDebtToCapTTM       float64 `json:"longTermDebtToCapitalizationTTM"`
	TotalDebtToCapTTM          float64 `json:"totalDebtToCapitalizationTTM"`
	InterestCoverageTTM        float64 `json:"interestCoverageTTM"`
	CashFlowToDebtRatioTTM     float64 `json:"cashFlowToDebtRatioTTM"`
	ROETTM                     float64 `json:"returnOnEquityTTM"`
	ROATTM                     float64 `json:"returnOnAssetsTTM"`
	ROICTTM                    float64 `json:"returnOnCapitalEmployedTTM"`
	AssetTurnoverTTM           float64 `json:"assetTurnoverTTM"`
	InventoryTurnoverTTM       float64 `json:"inventoryTurnoverTTM"`
	ReceivablesTurnoverTTM     float64 `json:"receivablesTurnoverTTM"`
	PayablesTurnoverTTM        float64 `json:"payablesTurnoverTTM"`
	PriceBookValueRatioTTM     float64 `json:"priceBookValueRatioTTM"`
	PriceToSalesRatioTTM       float64 `json:"priceToSalesRatioTTM"`
	PriceEarningsToGrowthTTM   float64 `json:"priceEarningsToGrowthRatioTTM"`
	PriceCashFlowRatioTTM      float64 `json:"priceCashFlowRatioTTM"`
	PriceToFreeCashFlowTTM     float64 `json:"priceToFreeCashFlowsRatioTTM"`
	EnterpriseValueMultipleTTM float64 `json:"enterpriseValueMultipleTTM"`
	DividendPayoutRatioTTM     float64 `json:"dividendPayoutRatioTTM"`
}

// RevenueSegment represents revenue by product or geographic segment
type RevenueSegment struct {
	Date    string             `json:"date"`
	Symbol  string             `json:"symbol"`
	Revenue map[string]float64 `json:"-"` // Parsed from dynamic keys
}

// PriceTargetSummary represents analyst price target summary
type PriceTargetSummary struct {
	Symbol                    string  `json:"symbol"`
	LastMonth                 int     `json:"lastMonth"`
	LastMonthAvgPriceTarget   float64 `json:"lastMonthAvgPriceTarget"`
	LastQuarter               int     `json:"lastQuarter"`
	LastQuarterAvgPriceTarget float64 `json:"lastQuarterAvgPriceTarget"`
	LastYear                  int     `json:"lastYear"`
	LastYearAvgPriceTarget    float64 `json:"lastYearAvgPriceTarget"`
	AllTime                   int     `json:"allTime"`
	AllTimeAvgPriceTarget     float64 `json:"allTimeAvgPriceTarget"`
}

// AnalystEstimate represents analyst estimates for a period
type AnalystEstimate struct {
	Date                    string  `json:"date"`
	Symbol                  string  `json:"symbol"`
	EstimatedRevenueLow     float64 `json:"revenueLow"`
	EstimatedRevenueHigh    float64 `json:"revenueHigh"`
	EstimatedRevenueAvg     float64 `json:"revenueAvg"`
	EstimatedEBITDALow      float64 `json:"ebitdaLow"`
	EstimatedEBITDAHigh     float64 `json:"ebitdaHigh"`
	EstimatedEBITDAAvg      float64 `json:"ebitdaAvg"`
	EstimatedEPSLow         float64 `json:"epsLow"`
	EstimatedEPSHigh        float64 `json:"epsHigh"`
	EstimatedEPSAvg         float64 `json:"epsAvg"`
	EstimatedNetIncomeLow   float64 `json:"netIncomeLow"`
	EstimatedNetIncomeHigh  float64 `json:"netIncomeHigh"`
	EstimatedNetIncomeAvg   float64 `json:"netIncomeAvg"`
	NumberAnalystEstRevenue int     `json:"numAnalystsRevenue"`
	NumberAnalystsEstEPS    int     `json:"numAnalystsEps"`
}

// UpgradeDowngrade represents analyst upgrade/downgrade
type UpgradeDowngrade struct {
	Symbol         string `json:"symbol"`
	PublishedDate  string `json:"date"`
	NewsURL        string `json:"newsURL"`
	NewsTitle      string `json:"newsTitle"`
	NewsPublisher  string `json:"newsPublisher"`
	NewGrade       string `json:"newGrade"`
	PreviousGrade  string `json:"previousGrade"`
	GradingCompany string `json:"gradingCompany"`
	Action         string `json:"action"`
}

// HistoricalPrice represents a single day's price data
type HistoricalPrice struct {
	Date     string  `json:"date"`
	Open     float64 `json:"open"`
	High     float64 `json:"high"`
	Low      float64 `json:"low"`
	Close    float64 `json:"close"`
	AdjClose float64 `json:"adjClose"`
	Volume   int64   `json:"volume"`
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) Search(query string, limit int) ([]SearchResult, error) {
	if limit <= 0 {
		limit = 10
	}

	// Request more results for better sorting candidates
	fetchLimit := limit * 3
	if fetchLimit < 30 {
		fetchLimit = 30
	}

	// Query both the name and symbol indexes and merge: a query like
	// "apple" must surface AAPL even when the name index buries it, and
	// "tsla" must work even though no company name matches.
	byName, nameErr := c.searchByName(query, fetchLimit)
	bySymbol, symErr := c.searchBySymbol(query, fetchLimit)
	if nameErr != nil && symErr != nil {
		return nil, nameErr
	}
	seen := make(map[string]bool, len(byName)+len(bySymbol))
	results := make([]SearchResult, 0, len(byName)+len(bySymbol))
	for _, r := range append(byName, bySymbol...) {
		if seen[r.Symbol] {
			continue
		}
		seen[r.Symbol] = true
		results = append(results, r)
	}

	// Sort results to prioritize pure US symbols
	results = sortSearchResults(results, query)

	// Return only the requested limit
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// scoreResult assigns a score to a search result (lower is better)
func scoreResult(r SearchResult, query string) int {
	score := 0
	upperQuery := strings.ToUpper(query)
	upperSymbol := strings.ToUpper(r.Symbol)

	// Penalize symbols with dots (foreign exchanges like .DE, .MX)
	if strings.Contains(r.Symbol, ".") {
		score += 100
	}

	// Penalize OTC stocks
	if r.ExchangeShortName == "OTC" || strings.Contains(r.StockExchange, "OTC") {
		score += 200
	}

	// Reward major US exchanges
	switch r.ExchangeShortName {
	case "NASDAQ", "NYSE", "AMEX":
		score -= 50
	}

	// Strongly reward exact symbol match
	if upperSymbol == upperQuery {
		score -= 100
	}

	// Reward symbol starting with query
	if strings.HasPrefix(upperSymbol, upperQuery) {
		score -= 25
	}

	return score
}

// sortSearchResults filters out crypto and sorts results to prioritize pure US symbols
func sortSearchResults(results []SearchResult, query string) []SearchResult {
	// Filter out crypto results entirely
	filtered := make([]SearchResult, 0, len(results))
	for _, r := range results {
		if r.ExchangeShortName != "CRYPTO" && r.ExchangeShortName != "CCC" {
			filtered = append(filtered, r)
		}
	}

	// Sort by score
	sort.SliceStable(filtered, func(i, j int) bool {
		scoreI := scoreResult(filtered[i], query)
		scoreJ := scoreResult(filtered[j], query)
		return scoreI < scoreJ
	})
	return filtered
}

func (c *Client) searchByName(query string, limit int) ([]SearchResult, error) {
	endpoint := fmt.Sprintf("%s/search-name?query=%s&limit=%d&apikey=%s",
		stableBaseURL,
		url.QueryEscape(query),
		limit,
		c.apiKey,
	)
	return c.doSearchRequest(endpoint)
}

func (c *Client) searchBySymbol(query string, limit int) ([]SearchResult, error) {
	endpoint := fmt.Sprintf("%s/search-symbol?query=%s&limit=%d&apikey=%s",
		stableBaseURL,
		url.QueryEscape(query),
		limit,
		c.apiKey,
	)
	return c.doSearchRequest(endpoint)
}

func (c *Client) doSearchRequest(endpoint string) ([]SearchResult, error) {
	// Log the endpoint without the API key for security
	cleanEndpoint := endpoint
	if idx := strings.LastIndex(endpoint, "apikey="); idx > 0 {
		cleanEndpoint = endpoint[:idx] + "apikey=***"
	}
	log.Printf("[DEBUG] FMP search request: %s", cleanEndpoint)

	resp, err := c.httpClient.Get(endpoint)
	if err != nil {
		log.Printf("[ERROR] FMP search request failed: %v", err)
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] FMP search response status %d: %s", resp.StatusCode, string(bodyBytes))
		return nil, fmt.Errorf("search failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var results []SearchResult
	if err := json.Unmarshal(bodyBytes, &results); err != nil {
		log.Printf("[ERROR] FMP search decode failed. Body preview: %.500s", string(bodyBytes))
		return nil, fmt.Errorf("failed to decode search results: %w", err)
	}

	log.Printf("[DEBUG] FMP search returned %d results", len(results))
	return results, nil
}

func (c *Client) GetProfile(symbol string) (*StockProfile, error) {
	endpoint := fmt.Sprintf("%s/profile?symbol=%s&apikey=%s",
		stableBaseURL,
		url.QueryEscape(symbol),
		c.apiKey,
	)

	resp, err := c.httpClient.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("profile request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("profile failed with status: %d", resp.StatusCode)
	}

	var profiles []StockProfile
	if err := json.NewDecoder(resp.Body).Decode(&profiles); err != nil {
		return nil, fmt.Errorf("failed to decode profile: %w", err)
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("stock not found: %s", symbol)
	}

	return &profiles[0], nil
}

// GetIncomeStatement fetches income statements for a symbol
// period: "annual" or "quarter", limit: number of periods to fetch
func (c *Client) GetIncomeStatement(symbol, period string, limit int) ([]IncomeStatement, error) {
	endpoint := fmt.Sprintf("%s/income-statement?symbol=%s&period=%s&limit=%d&apikey=%s",
		stableBaseURL,
		url.QueryEscape(symbol),
		period,
		limit,
		c.apiKey,
	)

	log.Printf("[DEBUG] FMP GetIncomeStatement: %s period=%s limit=%d", symbol, period, limit)
	var results []IncomeStatement
	if err := c.doRequest(endpoint, &results); err != nil {
		log.Printf("[ERROR] FMP GetIncomeStatement(%s): %v", symbol, err)
		return nil, fmt.Errorf("income statement request failed: %w", err)
	}
	log.Printf("[DEBUG] FMP GetIncomeStatement(%s): got %d results", symbol, len(results))
	return results, nil
}

// GetBalanceSheet fetches balance sheets for a symbol
func (c *Client) GetBalanceSheet(symbol, period string, limit int) ([]BalanceSheet, error) {
	endpoint := fmt.Sprintf("%s/balance-sheet-statement?symbol=%s&period=%s&limit=%d&apikey=%s",
		stableBaseURL,
		url.QueryEscape(symbol),
		period,
		limit,
		c.apiKey,
	)

	log.Printf("[DEBUG] FMP GetBalanceSheet: %s period=%s limit=%d", symbol, period, limit)
	var results []BalanceSheet
	if err := c.doRequest(endpoint, &results); err != nil {
		log.Printf("[ERROR] FMP GetBalanceSheet(%s): %v", symbol, err)
		return nil, fmt.Errorf("balance sheet request failed: %w", err)
	}
	log.Printf("[DEBUG] FMP GetBalanceSheet(%s): got %d results", symbol, len(results))
	return results, nil
}

// GetCashFlowStatement fetches cash flow statements for a symbol
func (c *Client) GetCashFlowStatement(symbol, period string, limit int) ([]CashFlowStatement, error) {
	endpoint := fmt.Sprintf("%s/cash-flow-statement?symbol=%s&period=%s&limit=%d&apikey=%s",
		stableBaseURL,
		url.QueryEscape(symbol),
		period,
		limit,
		c.apiKey,
	)

	log.Printf("[DEBUG] FMP GetCashFlowStatement: %s period=%s limit=%d", symbol, period, limit)
	var results []CashFlowStatement
	if err := c.doRequest(endpoint, &results); err != nil {
		log.Printf("[ERROR] FMP GetCashFlowStatement(%s): %v", symbol, err)
		return nil, fmt.Errorf("cash flow statement request failed: %w", err)
	}
	log.Printf("[DEBUG] FMP GetCashFlowStatement(%s): got %d results", symbol, len(results))
	return results, nil
}

// GetKeyMetrics fetches key financial metrics for a symbol
func (c *Client) GetKeyMetrics(symbol, period string, limit int) ([]KeyMetrics, error) {
	endpoint := fmt.Sprintf("%s/key-metrics?symbol=%s&period=%s&limit=%d&apikey=%s",
		stableBaseURL,
		url.QueryEscape(symbol),
		period,
		limit,
		c.apiKey,
	)

	var results []KeyMetrics
	if err := c.doRequest(endpoint, &results); err != nil {
		return nil, fmt.Errorf("key metrics request failed: %w", err)
	}
	return results, nil
}

// GetRatiosTTM fetches trailing twelve months financial ratios
func (c *Client) GetRatiosTTM(symbol string) (*RatiosTTM, error) {
	endpoint := fmt.Sprintf("%s/ratios-ttm?symbol=%s&apikey=%s",
		stableBaseURL,
		url.QueryEscape(symbol),
		c.apiKey,
	)

	var results []RatiosTTM
	if err := c.doRequest(endpoint, &results); err != nil {
		return nil, fmt.Errorf("ratios TTM request failed: %w", err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no ratios found for: %s", symbol)
	}
	return &results[0], nil
}

// GetRevenueProductSegmentation fetches revenue breakdown by product
func (c *Client) GetRevenueProductSegmentation(symbol string) ([]map[string]interface{}, error) {
	endpoint := fmt.Sprintf("%s/revenue-product-segmentation?symbol=%s&apikey=%s",
		stableBaseURL,
		url.QueryEscape(symbol),
		c.apiKey,
	)

	log.Printf("[DEBUG] FMP GetRevenueProductSegmentation: %s", symbol)
	var results []map[string]interface{}
	if err := c.doRequest(endpoint, &results); err != nil {
		log.Printf("[ERROR] FMP GetRevenueProductSegmentation(%s): %v", symbol, err)
		return nil, fmt.Errorf("revenue product segmentation request failed: %w", err)
	}
	log.Printf("[DEBUG] FMP GetRevenueProductSegmentation(%s): got %d results", symbol, len(results))

	// Flatten the nested "data" structure - FMP returns products inside a "data" field
	// Only include date and symbol for metadata, plus the actual segment data
	flattened := make([]map[string]interface{}, 0, len(results))
	for _, item := range results {
		flat := make(map[string]interface{})
		// Copy only date and symbol (NOT fiscalYear - it's a number that confuses the parser)
		if date, ok := item["date"]; ok {
			flat["date"] = date
		}
		if sym, ok := item["symbol"]; ok {
			flat["symbol"] = sym
		}
		// Flatten the nested "data" object containing actual segment values
		if data, ok := item["data"].(map[string]interface{}); ok {
			for k, v := range data {
				flat[k] = v
			}
		}
		flattened = append(flattened, flat)
	}
	return flattened, nil
}

// GetRevenueGeographicSegmentation fetches revenue breakdown by geography
func (c *Client) GetRevenueGeographicSegmentation(symbol string) ([]map[string]interface{}, error) {
	endpoint := fmt.Sprintf("%s/revenue-geographic-segmentation?symbol=%s&apikey=%s",
		stableBaseURL,
		url.QueryEscape(symbol),
		c.apiKey,
	)

	log.Printf("[DEBUG] FMP GetRevenueGeographicSegmentation: %s", symbol)
	var results []map[string]interface{}
	if err := c.doRequest(endpoint, &results); err != nil {
		log.Printf("[ERROR] FMP GetRevenueGeographicSegmentation(%s): %v", symbol, err)
		return nil, fmt.Errorf("revenue geographic segmentation request failed: %w", err)
	}
	log.Printf("[DEBUG] FMP GetRevenueGeographicSegmentation(%s): got %d results", symbol, len(results))

	// Flatten the nested "data" structure - FMP returns regions inside a "data" field
	// Only include date and symbol for metadata, plus the actual segment data
	flattened := make([]map[string]interface{}, 0, len(results))
	for _, item := range results {
		flat := make(map[string]interface{})
		// Copy only date and symbol (NOT fiscalYear - it's a number that confuses the parser)
		if date, ok := item["date"]; ok {
			flat["date"] = date
		}
		if sym, ok := item["symbol"]; ok {
			flat["symbol"] = sym
		}
		// Flatten the nested "data" object containing actual segment values
		if data, ok := item["data"].(map[string]interface{}); ok {
			for k, v := range data {
				flat[k] = v
			}
		}
		flattened = append(flattened, flat)
	}
	return flattened, nil
}

// GetPriceTargetSummary fetches analyst price target summary
func (c *Client) GetPriceTargetSummary(symbol string) (*PriceTargetSummary, error) {
	endpoint := fmt.Sprintf("%s/price-target-summary?symbol=%s&apikey=%s",
		stableBaseURL,
		url.QueryEscape(symbol),
		c.apiKey,
	)

	var results []PriceTargetSummary
	if err := c.doRequest(endpoint, &results); err != nil {
		return nil, fmt.Errorf("price target summary request failed: %w", err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no price targets found for: %s", symbol)
	}
	return &results[0], nil
}

// GetAnalystEstimates fetches analyst estimates for future periods
func (c *Client) GetAnalystEstimates(symbol string, limit int) ([]AnalystEstimate, error) {
	endpoint := fmt.Sprintf("%s/analyst-estimates?symbol=%s&period=annual&limit=%d&apikey=%s",
		stableBaseURL, url.QueryEscape(symbol), limit, c.apiKey)
	var result []AnalystEstimate
	if err := c.doRequest(endpoint, &result); err != nil {
		return nil, fmt.Errorf("analyst estimates request failed: %w", err)
	}
	return result, nil
}

// GetUpgradesDowngrades fetches recent analyst upgrades and downgrades
func (c *Client) GetUpgradesDowngrades(symbol string, limit int) ([]UpgradeDowngrade, error) {
	endpoint := fmt.Sprintf("%s/grades?symbol=%s&apikey=%s",
		stableBaseURL, url.QueryEscape(symbol), c.apiKey)
	var result []UpgradeDowngrade
	if err := c.doRequest(endpoint, &result); err != nil {
		return nil, fmt.Errorf("grades request failed: %w", err)
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

// GetHistoricalPrice fetches historical price data for a symbol
func (c *Client) GetHistoricalPrice(symbol, from, to string) ([]HistoricalPrice, error) {
	endpoint := fmt.Sprintf("%s/historical-price-eod/full?symbol=%s&from=%s&to=%s&apikey=%s",
		stableBaseURL,
		url.QueryEscape(symbol),
		from,
		to,
		c.apiKey,
	)

	// The stable API returns an array directly
	var results []HistoricalPrice
	if err := c.doRequest(endpoint, &results); err != nil {
		return nil, fmt.Errorf("historical price request failed: %w", err)
	}
	return results, nil
}

// doRequest is a helper function to make HTTP requests and decode JSON responses
func (c *Client) doRequest(endpoint string, result interface{}) error {
	// Log the endpoint without the API key for security
	cleanEndpoint := endpoint
	if idx := len(endpoint) - len(c.apiKey); idx > 0 && endpoint[idx:] == c.apiKey {
		cleanEndpoint = endpoint[:idx] + "***"
	}
	log.Printf("[DEBUG] FMP API request: %s", cleanEndpoint)

	resp, err := c.httpClient.Get(endpoint)
	if err != nil {
		log.Printf("[ERROR] FMP API request failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] FMP API response status %d: %s", resp.StatusCode, string(body))
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, result); err != nil {
		log.Printf("[ERROR] FMP API decode failed. Body preview: %.500s", string(body))
		return fmt.Errorf("failed to decode response: %w", err)
	}
	return nil
}
