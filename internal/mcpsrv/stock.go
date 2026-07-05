// Copyright (c) 2026 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package mcpsrv

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/karamble/brmcp/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	kit "github.com/vctt94/bisonbotkit"

	"github.com/karamble/braibot/internal/fmp"
)

// stockFundamentalsUSD is the flat resale price per fundamentals call. FMP
// usage is metered on the operator's plan and the shared 24h cache makes
// repeat symbols pure margin.
const stockFundamentalsUSD = 0.05

type stockSearchIn struct {
	Query string `json:"query" jsonschema:"company name or ticker fragment to search"`
	Limit int    `json:"limit,omitempty" jsonschema:"max results (default 10, max 25)"`
}

type stockFundamentalsIn struct {
	Symbol string `json:"symbol" jsonschema:"stock ticker symbol, e.g. AAPL"`
	Period string `json:"period,omitempty" jsonschema:"statement period: annual (default) or quarter"`
	Format string `json:"format,omitempty" jsonschema:"json (structured result, default) or html (single-file report delivered to your Bison Relay DM)"`
}

// AttachStock registers the FMP-backed stock tools on the harness.
func AttachStock(h *server.Harness, svc *fmp.Service, bot *kit.Bot) {
	server.AddTool(h, &mcp.Tool{
		Name:        "stock_search",
		Description: "Search stock tickers by company name or symbol. Returns ranked matches from major exchanges.",
	}, 0, func(_ context.Context, _ string, in stockSearchIn) (any, error) {
		if strings.TrimSpace(in.Query) == "" {
			return nil, errors.New("query is required")
		}
		res, err := svc.Search(in.Query, in.Limit)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]any, 0, len(res))
		for _, r := range res {
			out = append(out, map[string]any{
				"symbol":   r.Symbol,
				"name":     r.Name,
				"exchange": r.ExchangeShortName,
				"currency": r.Currency,
			})
		}
		return map[string]any{"results": out}, nil
	})

	server.AddToolPriced(h, &mcp.Tool{
		Name: "stock_fundamentals",
		Description: "Full fundamentals for a ticker: profile, five periods of income/balance/cash flow, " +
			"key metrics, TTM ratios, revenue segmentation, analyst targets and actions, one year of prices. " +
			"format json returns the data in this result; format html delivers a self-contained tabbed report " +
			"file to your Bison Relay DM and this result only confirms delivery.",
	}, func(_ context.Context, _ string, _ stockFundamentalsIn) (int64, error) {
		return usdToAtoms(stockFundamentalsUSD)
	}, func(ctx context.Context, peer string, in stockFundamentalsIn) (any, error) {
		f, err := svc.Fundamentals(in.Symbol, in.Period)
		if err != nil {
			return nil, err
		}
		if strings.EqualFold(in.Format, "html") {
			dir, err := os.MkdirTemp("", "braibot-report-*")
			if err != nil {
				return nil, err
			}
			defer os.RemoveAll(dir)
			name := fmt.Sprintf("%s-fundamentals-%s.html", f.Symbol, time.Now().Format("2006-01-02"))
			path := filepath.Join(dir, name)
			if err := os.WriteFile(path, []byte(f.RenderHTML()), 0o600); err != nil {
				return nil, err
			}
			if err := bot.SendFile(ctx, peer, path); err != nil {
				return nil, fmt.Errorf("send report: %w", err)
			}
			return delivered("fmp", "report", map[string]any{
				"symbol": f.Symbol,
				"file":   name,
			}), nil
		}
		return map[string]any{
			"symbol":             f.Symbol,
			"period":             f.Period,
			"generatedAt":        f.GeneratedAt,
			"profile":            f.Profile,
			"income":             f.Income,
			"balance":            f.Balance,
			"cashFlow":           f.CashFlow,
			"keyMetrics":         f.KeyMetrics,
			"ratiosTtm":          f.RatiosTTM,
			"revenueByProduct":   f.RevenueProd,
			"revenueByGeography": f.RevenueGeo,
			"priceTargets":       f.Targets,
			"analystEstimates":   f.Estimates,
			"upgradesDowngrades": f.Upgrades,
			"historyMonthly":     f.MonthlyHistory(),
		}, nil
	})
}
