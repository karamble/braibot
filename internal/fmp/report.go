package fmp

import (
	"fmt"
	"html"
	"math"
	"strings"
)

// RenderHTML produces the single-file fundamentals report: tabbed layout,
// inline SVG charts, embedded logo, no external assets. Palette and mark
// rules follow the validated dark reference set (series blue/aqua/yellow/red
// on surface #1a1a19, direct labels on all multi-series charts).
func (f *Fundamentals) RenderHTML() string {
	var b strings.Builder
	p := f.Profile
	name := esc(p.CompanyName)
	sym := esc(f.Symbol)

	b.WriteString(`<!DOCTYPE html><html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>` + sym + ` Fundamentals - ` + name + `</title>
<style>` + reportCSS + `</style></head><body><div class="wrap">`)

	// ---- header ----
	b.WriteString(`<header class="head">`)
	if f.LogoData != "" {
		b.WriteString(`<img class="logo" src="` + f.LogoData + `" alt="">`)
	} else {
		mono := "?"
		if len(f.Symbol) > 0 {
			mono = f.Symbol[:1]
		}
		b.WriteString(`<div class="logo mono">` + esc(mono) + `</div>`)
	}
	chClass, chSign := "up", "+"
	if p.ChangePercentage < 0 {
		chClass, chSign = "down", ""
	}
	b.WriteString(`<div class="head-id"><h1>` + name + `</h1><div class="sub">` +
		sym + ` · ` + esc(p.ExchangeShortName) + ` · ` + esc(p.Sector) + ` · ` + esc(p.Industry) +
		`</div></div><div class="head-px"><div class="px">` + money(p.Price, p.Currency) +
		`</div><div class="chg ` + chClass + `">` + chSign + fmt.Sprintf("%.2f%%", p.ChangePercentage) +
		`</div></div></header>`)

	// ---- stat tiles ----
	b.WriteString(`<section class="tiles">` +
		tile("Market cap", bigNum(p.MarketCap)) +
		tile("Beta", fmt.Sprintf("%.2f", p.Beta)) +
		tile("52w range", esc(p.Range)) +
		tile("Avg volume", bigNum(p.AverageVolume)) +
		tile("Dividend", fmt.Sprintf("%.2f", p.LastDividend)) +
		tile("Employees", esc(p.FullTimeEmployees)) +
		`</section>`)

	// ---- tabs ----
	tabs := []struct{ id, label string }{
		{"overview", "Overview"}, {"financials", "Financials"},
		{"ratios", "Ratios & Metrics"}, {"revenue", "Revenue"},
		{"analysts", "Analysts"}, {"price", "Price Chart"},
	}
	for i, t := range tabs {
		checked := ""
		if i == 0 {
			checked = " checked"
		}
		b.WriteString(`<input type="radio" name="tab" id="tab-` + t.id + `"` + checked + `>`)
	}
	b.WriteString(`<nav class="tabbar">`)
	for _, t := range tabs {
		b.WriteString(`<label for="tab-` + t.id + `">` + t.label + `</label>`)
	}
	b.WriteString(`</nav>`)

	b.WriteString(`<section class="tab tab-overview">` + f.overviewTab() + `</section>`)
	b.WriteString(`<section class="tab tab-financials">` + f.financialsTab() + `</section>`)
	b.WriteString(`<section class="tab tab-ratios">` + f.ratiosTab() + `</section>`)
	b.WriteString(`<section class="tab tab-revenue">` + f.revenueTab() + `</section>`)
	b.WriteString(`<section class="tab tab-analysts">` + f.analystsTab() + `</section>`)
	b.WriteString(`<section class="tab tab-price">` + f.priceTab() + `</section>`)

	b.WriteString(`<footer>Generated ` + f.GeneratedAt.Format("2006-01-02 15:04 UTC") +
		` · Data: Financial Modeling Prep · braibot</footer></div>` + tooltipJS + `</body></html>`)
	return b.String()
}

// ---- tabs ----

func (f *Fundamentals) overviewTab() string {
	var b strings.Builder
	p := f.Profile
	b.WriteString(`<div class="cards"><div class="card grow"><h2>About</h2><p class="desc">` +
		esc(p.Description) + `</p><table class="kv">` +
		kv("CEO", p.CEO) + kv("Country", p.Country) + kv("IPO", p.IPODate) +
		kv("Website", p.Website) + kv("Currency", p.Currency) +
		`</table></div>`)
	if len(f.Income) > 0 {
		b.WriteString(`<div class="card grow"><h2>Revenue & Net income</h2>` + f.revNetChart() + `</div>`)
	}
	b.WriteString(`</div>`)
	return b.String()
}

func (f *Fundamentals) financialsTab() string {
	if len(f.Income) == 0 {
		return `<div class="card"><p class="muted">No financial statements available.</p></div>`
	}
	var b strings.Builder
	b.WriteString(`<div class="card"><h2>Margin trend</h2>` + f.marginChart() + `</div>`)

	b.WriteString(`<div class="card"><h2>Income statement (` + esc(f.Period) + `)</h2>`)
	b.WriteString(finTable(f.Income, []finRow{
		{"Revenue", func(i IncomeStatement) float64 { return float64(i.Revenue) }},
		{"Gross profit", func(i IncomeStatement) float64 { return float64(i.GrossProfit) }},
		{"Operating income", func(i IncomeStatement) float64 { return float64(i.OperatingIncome) }},
		{"Net income", func(i IncomeStatement) float64 { return float64(i.NetIncome) }},
		{"EBITDA", func(i IncomeStatement) float64 { return float64(i.EBITDA) }},
		{"EPS", func(i IncomeStatement) float64 { return i.EPS }},
	}))
	b.WriteString(`</div>`)

	if len(f.Balance) > 0 {
		b.WriteString(`<div class="card"><h2>Balance sheet</h2>`)
		b.WriteString(balTable(f.Balance))
		b.WriteString(`</div>`)
	}
	if len(f.CashFlow) > 0 {
		b.WriteString(`<div class="card"><h2>Cash flow</h2>`)
		b.WriteString(cfTable(f.CashFlow))
		b.WriteString(`</div>`)
	}
	return b.String()
}

func (f *Fundamentals) ratiosTab() string {
	var b strings.Builder
	if r := f.RatiosTTM; r != nil {
		b.WriteString(`<div class="card"><h2>Trailing twelve months</h2><div class="grid">` +
			tile("P/E", num2(orDerived(r.PERatioTTM, f.derivedPE()))) +
			tile("P/B", num2(orDerived(r.PriceBookValueRatioTTM, f.derivedPB()))) +
			tile("P/S", num2(r.PriceToSalesRatioTTM)) +
			tile("EV multiple", num2(r.EnterpriseValueMultipleTTM)) +
			tile("Gross margin", pct(r.GrossProfitMarginTTM)) + tile("Operating margin", pct(r.OperatingProfitMarginTTM)) +
			tile("Net margin", pct(r.NetProfitMarginTTM)) + tile("Current ratio", num2(r.CurrentRatioTTM)) +
			tile("Quick ratio", num2(r.QuickRatioTTM)) + tile("Asset turnover", num2(r.AssetTurnoverTTM)) +
			tile("ROE", pct(orDerived(r.ROETTM, f.derivedROE()))) +
			tile("Dividend yield", pct(orDerived(r.DividendYieldTTM, f.derivedDivYield()))) +
			`</div></div>`)
	}
	if len(f.KeyMetrics) > 0 {
		b.WriteString(`<div class="card"><h2>Key metrics (annual)</h2><table class="data"><thead><tr><th>Metric</th>`)
		for _, m := range f.KeyMetrics {
			b.WriteString(`<th>` + esc(yearOf(m.CalendarYear, m.Date)) + `</th>`)
		}
		b.WriteString(`</tr></thead><tbody>`)
		rows := []struct {
			label string
			get   func(KeyMetrics) float64
			fmt   func(float64) string
		}{
			{"Market cap", func(m KeyMetrics) float64 { return float64(m.MarketCap) }, bigNum},
			{"Enterprise value", func(m KeyMetrics) float64 { return float64(m.EnterpriseValue) }, bigNum},
			{"EV / sales", func(m KeyMetrics) float64 { return m.EVToSales }, num2},
			{"EV / free cash flow", func(m KeyMetrics) float64 { return m.EVToFreeCashFlow }, num2},
			{"Earnings yield", func(m KeyMetrics) float64 { return m.EarningsYield }, pct},
			{"Free cash flow yield", func(m KeyMetrics) float64 { return m.FreeCashFlowYield }, pct},
			{"Net debt / EBITDA", func(m KeyMetrics) float64 { return m.NetDebtToEBITDA }, num2},
			{"Current ratio", func(m KeyMetrics) float64 { return m.CurrentRatio }, num2},
			{"Income quality", func(m KeyMetrics) float64 { return m.IncomeQuality }, num2},
		}
		for _, r := range rows {
			b.WriteString(`<tr><td>` + r.label + `</td>`)
			for _, m := range f.KeyMetrics {
				b.WriteString(`<td class="n">` + r.fmt(r.get(m)) + `</td>`)
			}
			b.WriteString(`</tr>`)
		}
		b.WriteString(`</tbody></table></div>`)
	}
	if f.RatiosTTM == nil && len(f.KeyMetrics) == 0 {
		b.WriteString(`<div class="card"><p class="muted">No ratio data available.</p></div>`)
	}
	return b.String()
}

func (f *Fundamentals) revenueTab() string {
	var b strings.Builder
	prod := f.segBars("By product", f.RevenueProd)
	geo := f.segBars("By geography", f.RevenueGeo)
	if prod == "" && geo == "" {
		return `<div class="card"><p class="muted">No revenue segmentation available.</p></div>`
	}
	b.WriteString(prod)
	b.WriteString(geo)
	return b.String()
}

func (f *Fundamentals) analystsTab() string {
	var b strings.Builder
	if t := f.Targets; t != nil && t.LastMonthAvgPriceTarget > 0 {
		b.WriteString(`<div class="card"><h2>Price targets</h2>` + f.targetChart(t) + `</div>`)
	}
	if len(f.Upgrades) > 0 {
		b.WriteString(`<div class="card"><h2>Recent analyst actions</h2><table class="data"><thead><tr>` +
			`<th>Date</th><th>Firm</th><th>Action</th><th>Grade</th></tr></thead><tbody>`)
		max := len(f.Upgrades)
		if max > 12 {
			max = 12
		}
		for _, u := range f.Upgrades[:max] {
			cls := "muted"
			switch strings.ToLower(u.Action) {
			case "upgrade":
				cls = "up"
			case "downgrade":
				cls = "down"
			}
			date := u.PublishedDate
			if len(date) > 10 {
				date = date[:10]
			}
			grade := u.NewGrade
			if u.PreviousGrade != "" && u.PreviousGrade != u.NewGrade {
				grade = u.PreviousGrade + " → " + u.NewGrade
			}
			b.WriteString(`<tr><td>` + esc(date) + `</td><td>` + esc(u.GradingCompany) +
				`</td><td class="` + cls + `">` + esc(strings.Title(u.Action)) + `</td><td>` + esc(grade) + `</td></tr>`)
		}
		b.WriteString(`</tbody></table></div>`)
	}
	if len(f.Estimates) > 0 {
		b.WriteString(`<div class="card"><h2>Analyst estimates</h2><table class="data"><thead><tr>` +
			`<th>Period</th><th>Revenue avg</th><th>EPS avg</th><th>Analysts</th></tr></thead><tbody>`)
		for _, e := range f.Estimates {
			date := e.Date
			if len(date) > 10 {
				date = date[:10]
			}
			b.WriteString(`<tr><td>` + esc(date) + `</td><td class="n">` + bigNum(e.EstimatedRevenueAvg) +
				`</td><td class="n">` + num2(e.EstimatedEPSAvg) + `</td><td class="n">` +
				fmt.Sprintf("%d", e.NumberAnalystsEstEPS) + `</td></tr>`)
		}
		b.WriteString(`</tbody></table></div>`)
	}
	if b.Len() == 0 {
		return `<div class="card"><p class="muted">No analyst coverage data available.</p></div>`
	}
	return b.String()
}

func (f *Fundamentals) priceTab() string {
	if len(f.History) < 2 {
		return `<div class="card"><p class="muted">No price history available.</p></div>`
	}
	return `<div class="card"><h2>1 year price</h2>` + f.priceChart() + `</div>`
}

// ---- charts (inline SVG, one axis, thin marks, direct labels) ----

const (
	cBlue  = "#3987e5"
	cAqua  = "#199e70"
	cAmber = "#c98500"
	cRed   = "#e66767"
)

// revNetChart draws grouped bars of revenue and net income per period with a
// legend and direct value labels.
func (f *Fundamentals) revNetChart() string {
	n := len(f.Income)
	if n == 0 {
		return ""
	}
	// FMP returns newest first; draw oldest to newest.
	inc := make([]IncomeStatement, n)
	for i, v := range f.Income {
		inc[n-1-i] = v
	}
	W, H, padL, padB, padT := 640.0, 260.0, 8.0, 26.0, 26.0
	maxV := 1.0
	minV := 0.0
	for _, v := range inc {
		maxV = math.Max(maxV, math.Max(float64(v.Revenue), float64(v.NetIncome)))
		minV = math.Min(minV, float64(v.NetIncome))
	}
	span := maxV - minV
	y := func(v float64) float64 { return padT + (H-padT-padB)*(1-(v-minV)/span) }
	group := (W - padL - 8) / float64(n)
	barW := math.Min(42, group/2.6)
	var s strings.Builder
	s.WriteString(svgOpen(W, H))
	zero := y(0)
	s.WriteString(line(padL, zero, W-4, zero, "#3a3a38", 1))
	for i, v := range inc {
		x := padL + group*float64(i) + group/2
		label := esc(v.CalendarYear)
		if f.Period == "quarter" && len(v.Date) >= 7 {
			label = esc(v.Date[:7])
		}
		rev, ni := float64(v.Revenue), float64(v.NetIncome)
		s.WriteString(bar(x-barW-1, rev, zero, y(rev), barW, cBlue,
			label+" revenue "+bigNum(rev)))
		s.WriteString(bar(x+1, ni, zero, y(ni), barW, cAqua,
			label+" net income "+bigNum(ni)))
		s.WriteString(text(x, H-8, label, "mid", "#c3c2b7", 12))
		s.WriteString(text(x-barW/2-1, y(math.Max(rev, 0))-5, bigNum(rev), "mid", "#c3c2b7", 10))
	}
	s.WriteString(`</svg>`)
	return s.String() + legend([2][2]string{{cBlue, "Revenue"}, {cAqua, "Net income"}})
}

// marginChart draws gross/operating/net margin percentage lines.
func (f *Fundamentals) marginChart() string {
	n := len(f.Income)
	if n < 2 {
		return ""
	}
	inc := make([]IncomeStatement, n)
	for i, v := range f.Income {
		inc[n-1-i] = v
	}
	type series struct {
		name  string
		color string
		vals  []float64
	}
	mk := func(get func(IncomeStatement) float64) []float64 {
		out := make([]float64, n)
		for i, v := range inc {
			if v.Revenue != 0 {
				out[i] = get(v) / float64(v.Revenue) * 100
			}
		}
		return out
	}
	ss := []series{
		{"Gross", cBlue, mk(func(i IncomeStatement) float64 { return float64(i.GrossProfit) })},
		{"Operating", cAqua, mk(func(i IncomeStatement) float64 { return float64(i.OperatingIncome) })},
		{"Net", cAmber, mk(func(i IncomeStatement) float64 { return float64(i.NetIncome) })},
	}
	W, H, padL, padB, padT, padR := 640.0, 240.0, 8.0, 26.0, 16.0, 70.0
	minV, maxV := 0.0, 10.0
	for _, sr := range ss {
		for _, v := range sr.vals {
			minV = math.Min(minV, v)
			maxV = math.Max(maxV, v)
		}
	}
	span := maxV - minV
	x := func(i int) float64 { return padL + (W-padL-padR)*float64(i)/float64(n-1) }
	y := func(v float64) float64 { return padT + (H-padT-padB)*(1-(v-minV)/span) }
	var s strings.Builder
	s.WriteString(svgOpen(W, H))
	zero := y(0)
	s.WriteString(line(padL, zero, W-padR, zero, "#3a3a38", 1))
	for _, sr := range ss {
		var pts strings.Builder
		for i, v := range sr.vals {
			if i > 0 {
				pts.WriteString(" ")
			}
			pts.WriteString(fmt.Sprintf("%.1f,%.1f", x(i), y(v)))
		}
		s.WriteString(`<polyline points="` + pts.String() + `" fill="none" stroke="` + sr.color + `" stroke-width="2"/>`)
		last := sr.vals[n-1]
		s.WriteString(text(W-padR+6, y(last)+4, sr.name+" "+fmt.Sprintf("%.0f%%", last), "start", sr.color, 11))
		for i, v := range sr.vals {
			s.WriteString(dot(x(i), y(v), 3.5, sr.color, labelFor(inc[i], f.Period)+" "+sr.name+" "+fmt.Sprintf("%.1f%%", v)))
		}
	}
	for i, v := range inc {
		s.WriteString(text(x(i), H-8, labelFor(v, f.Period), "mid", "#c3c2b7", 12))
	}
	s.WriteString(`</svg>`)
	return s.String()
}

// segBars renders the latest year's revenue segmentation as horizontal bars
// on a single-hue sequential scale.
func (f *Fundamentals) segBars(title string, segs []map[string]interface{}) string {
	if len(segs) == 0 {
		return ""
	}
	latest := segs[0]
	type seg struct {
		name string
		val  float64
	}
	var rows []seg
	var date string
	for k, v := range latest {
		switch k {
		case "date", "symbol":
			if k == "date" {
				date, _ = v.(string)
			}
			continue
		}
		if fv, ok := v.(float64); ok && fv > 0 {
			rows = append(rows, seg{k, fv})
		}
	}
	if len(rows) == 0 {
		return ""
	}
	for i := range rows {
		for j := i + 1; j < len(rows); j++ {
			if rows[j].val > rows[i].val {
				rows[i], rows[j] = rows[j], rows[i]
			}
		}
	}
	if len(rows) > 8 {
		var other float64
		for _, r := range rows[8:] {
			other += r.val
		}
		rows = append(rows[:8], seg{"Other", other})
	}
	maxV := rows[0].val
	var s strings.Builder
	s.WriteString(`<div class="card"><h2>` + esc(title))
	if date != "" {
		s.WriteString(` <span class="muted">(` + esc(date) + `)</span>`)
	}
	s.WriteString(`</h2><div class="segbars">`)
	for _, r := range rows {
		w := r.val / maxV * 100
		s.WriteString(`<div class="segrow" data-tip="` + esc(r.name) + ` ` + bigNum(r.val) +
			`"><span class="seglabel">` + esc(r.name) + `</span><span class="segtrack"><span class="segfill" style="width:` +
			fmt.Sprintf("%.1f", w) + `%"></span></span><span class="segval">` + bigNum(r.val) + `</span></div>`)
	}
	s.WriteString(`</div></div>`)
	return s.String()
}

// targetChart draws the analyst target range with the current price marked.
func (f *Fundamentals) targetChart(t *PriceTargetSummary) string {
	price := f.Profile.Price
	lo := math.Min(t.LastMonthAvgPriceTarget, price) * 0.9
	hi := math.Max(t.LastYearAvgPriceTarget, math.Max(t.LastMonthAvgPriceTarget, price)) * 1.1
	if hi <= lo {
		return ""
	}
	W, H := 640.0, 110.0
	x := func(v float64) float64 { return 20 + (W-40)*(v-lo)/(hi-lo) }
	mid := H/2 + 6
	var s strings.Builder
	s.WriteString(svgOpen(W, H))
	s.WriteString(line(20, mid, W-20, mid, "#3a3a38", 2))
	marks := []struct {
		v     float64
		label string
		color string
	}{
		{price, "Price " + num2(price), "#ffffff"},
		{t.LastMonthAvgPriceTarget, "1m avg target " + num2(t.LastMonthAvgPriceTarget), cBlue},
		{t.LastQuarterAvgPriceTarget, "3m avg " + num2(t.LastQuarterAvgPriceTarget), cAqua},
		{t.LastYearAvgPriceTarget, "1y avg " + num2(t.LastYearAvgPriceTarget), cAmber},
	}
	up := true
	for _, m := range marks {
		if m.v <= 0 {
			continue
		}
		s.WriteString(dot(x(m.v), mid, 5, m.color, m.label))
		ty := mid - 14.0
		if !up {
			ty = mid + 22.0
		}
		s.WriteString(text(x(m.v), ty, m.label, "mid", m.color, 11))
		up = !up
	}
	s.WriteString(`</svg>`)
	if price > 0 && t.LastMonthAvgPriceTarget > 0 {
		upside := (t.LastMonthAvgPriceTarget - price) / price * 100
		cls := "up"
		if upside < 0 {
			cls = "down"
		}
		return s.String() + `<p>Consensus (1m avg) implies <span class="` + cls + `">` +
			fmt.Sprintf("%+.1f%%", upside) + `</span> vs the current price.</p>`
	}
	return s.String()
}

// priceChart draws the 1y close line with a min/max band of direct labels.
func (f *Fundamentals) priceChart() string {
	n := len(f.History)
	hist := make([]HistoricalPrice, n)
	for i, v := range f.History {
		hist[n-1-i] = v
	}
	W, H, padL, padB, padT, padR := 640.0, 280.0, 8.0, 26.0, 18.0, 64.0
	minV, maxV := hist[0].Close, hist[0].Close
	for _, h := range hist {
		minV = math.Min(minV, h.Close)
		maxV = math.Max(maxV, h.Close)
	}
	span := maxV - minV
	if span == 0 {
		span = 1
	}
	x := func(i int) float64 { return padL + (W-padL-padR)*float64(i)/float64(n-1) }
	y := func(v float64) float64 { return padT + (H-padT-padB)*(1-(v-minV)/span) }
	var s strings.Builder
	s.WriteString(svgOpen(W, H))
	var pts strings.Builder
	for i, h := range hist {
		if i > 0 {
			pts.WriteString(" ")
		}
		pts.WriteString(fmt.Sprintf("%.1f,%.1f", x(i), y(h.Close)))
	}
	s.WriteString(`<polyline points="` + pts.String() + `" fill="none" stroke="` + cBlue + `" stroke-width="2"/>`)
	last := hist[n-1]
	s.WriteString(dot(x(n-1), y(last.Close), 4, cBlue, esc(last.Date)+" close "+num2(last.Close)))
	s.WriteString(text(W-padR+6, y(last.Close)+4, num2(last.Close), "start", cBlue, 12))
	// Sparse month labels along the baseline.
	lastMonth := ""
	for i, h := range hist {
		if len(h.Date) < 7 {
			continue
		}
		m := h.Date[:7]
		if m != lastMonth && i%2 == 0 {
			if lastMonth != "" && len(m) == 7 {
				s.WriteString(text(x(i), H-8, m[5:], "mid", "#7a7974", 10))
			}
			lastMonth = m
		}
	}
	// Hover samples: one dot per ~week keeps the DOM small.
	step := n / 52
	if step < 1 {
		step = 1
	}
	for i := 0; i < n; i += step {
		h := hist[i]
		s.WriteString(`<circle cx="` + fmt.Sprintf("%.1f", x(i)) + `" cy="` + fmt.Sprintf("%.1f", y(h.Close)) +
			`" r="7" fill="transparent" data-tip="` + esc(h.Date) + ` close ` + num2(h.Close) + `"/>`)
	}
	s.WriteString(text(padL, y(maxV)+4, "high "+num2(maxV), "start", "#c3c2b7", 10))
	s.WriteString(text(padL, y(minV)-6, "low "+num2(minV), "start", "#c3c2b7", 10))
	s.WriteString(`</svg>`)
	return s.String()
}

// ---- svg + table helpers ----

func svgOpen(w, h float64) string {
	return fmt.Sprintf(`<svg viewBox="0 0 %.0f %.0f" role="img" xmlns="http://www.w3.org/2000/svg">`, w, h)
}

func line(x1, y1, x2, y2 float64, color string, w float64) string {
	return fmt.Sprintf(`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="%.1f"/>`, x1, y1, x2, y2, color, w)
}

func bar(x, v, zero, top, w float64, color, tip string) string {
	y, h := top, zero-top
	if v < 0 {
		y, h = zero, top-zero
	}
	if h < 1 {
		h = 1
	}
	return fmt.Sprintf(`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="3" fill="%s" data-tip="%s"/>`,
		x, y, w, h, color, tip)
}

func dot(x, y, r float64, color, tip string) string {
	return fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="%.1f" fill="%s" stroke="#1a1a19" stroke-width="2" data-tip="%s"/>`,
		x, y, r, color, tip)
}

func text(x, y float64, s, anchor, color string, size int) string {
	a := map[string]string{"mid": "middle", "start": "start", "end": "end"}[anchor]
	return fmt.Sprintf(`<text x="%.1f" y="%.1f" text-anchor="%s" fill="%s" font-size="%d">%s</text>`,
		x, y, a, color, size, s)
}

func legend(items [2][2]string) string {
	var b strings.Builder
	b.WriteString(`<div class="legend">`)
	for _, it := range items {
		b.WriteString(`<span class="key"><span class="swatch" style="background:` + it[0] + `"></span>` + it[1] + `</span>`)
	}
	b.WriteString(`</div>`)
	return b.String()
}

func labelFor(i IncomeStatement, period string) string {
	if period == "quarter" && len(i.Date) >= 7 {
		return esc(i.Date[:7])
	}
	return esc(yearOf(i.CalendarYear, i.Date))
}

// yearOf falls back to the fiscal date when the stable API omits
// calendarYear.
func yearOf(calendarYear, date string) string {
	if calendarYear != "" {
		return calendarYear
	}
	if len(date) >= 4 {
		return date[:4]
	}
	return date
}

type finRow struct {
	label string
	get   func(IncomeStatement) float64
}

func finTable(inc []IncomeStatement, rows []finRow) string {
	var b strings.Builder
	b.WriteString(`<table class="data"><thead><tr><th></th>`)
	for _, i := range inc {
		b.WriteString(`<th>` + labelFor(i, "") + `</th>`)
	}
	b.WriteString(`</tr></thead><tbody>`)
	for _, r := range rows {
		b.WriteString(`<tr><td>` + r.label + `</td>`)
		for _, i := range inc {
			v := r.get(i)
			if r.label == "EPS" {
				b.WriteString(`<td class="n">` + num2(v) + `</td>`)
			} else {
				b.WriteString(`<td class="n">` + bigNum(v) + `</td>`)
			}
		}
		b.WriteString(`</tr>`)
	}
	b.WriteString(`</tbody></table>`)
	return b.String()
}

func balTable(bal []BalanceSheet) string {
	var b strings.Builder
	b.WriteString(`<table class="data"><thead><tr><th></th>`)
	for _, v := range bal {
		b.WriteString(`<th>` + esc(yearOf(v.CalendarYear, v.Date)) + `</th>`)
	}
	b.WriteString(`</tr></thead><tbody>`)
	rows := []struct {
		label string
		get   func(BalanceSheet) float64
	}{
		{"Total assets", func(v BalanceSheet) float64 { return float64(v.TotalAssets) }},
		{"Total liabilities", func(v BalanceSheet) float64 { return float64(v.TotalLiabilities) }},
		{"Total equity", func(v BalanceSheet) float64 { return float64(v.TotalEquity) }},
		{"Cash & equivalents", func(v BalanceSheet) float64 { return float64(v.CashAndCashEquivalents) }},
		{"Total debt", func(v BalanceSheet) float64 { return float64(v.TotalDebt) }},
		{"Net debt", func(v BalanceSheet) float64 { return float64(v.NetDebt) }},
	}
	for _, r := range rows {
		b.WriteString(`<tr><td>` + r.label + `</td>`)
		for _, v := range bal {
			b.WriteString(`<td class="n">` + bigNum(r.get(v)) + `</td>`)
		}
		b.WriteString(`</tr>`)
	}
	b.WriteString(`</tbody></table>`)
	return b.String()
}

func cfTable(cf []CashFlowStatement) string {
	var b strings.Builder
	b.WriteString(`<table class="data"><thead><tr><th></th>`)
	for _, v := range cf {
		b.WriteString(`<th>` + esc(yearOf(v.CalendarYear, v.Date)) + `</th>`)
	}
	b.WriteString(`</tr></thead><tbody>`)
	rows := []struct {
		label string
		get   func(CashFlowStatement) float64
	}{
		{"Operating cash flow", func(v CashFlowStatement) float64 { return float64(v.OperatingCashFlow) }},
		{"Capital expenditure", func(v CashFlowStatement) float64 { return float64(v.CapitalExpenditure) }},
		{"Free cash flow", func(v CashFlowStatement) float64 { return float64(v.FreeCashFlow) }},
		{"Dividends paid", func(v CashFlowStatement) float64 { return float64(v.DividendsPaid) }},
	}
	for _, r := range rows {
		b.WriteString(`<tr><td>` + r.label + `</td>`)
		for _, v := range cf {
			b.WriteString(`<td class="n">` + bigNum(r.get(v)) + `</td>`)
		}
		b.WriteString(`</tr>`)
	}
	b.WriteString(`</tbody></table>`)
	return b.String()
}

// orDerived prefers the API value, falling back to a locally derived one
// when the stable schema left the field zero.
func orDerived(v, derived float64) float64 {
	if v != 0 {
		return v
	}
	return derived
}

func (f *Fundamentals) derivedPE() float64 {
	if len(f.Income) > 0 && f.Income[0].EPS != 0 {
		return f.Profile.Price / f.Income[0].EPS
	}
	return 0
}

func (f *Fundamentals) derivedPB() float64 {
	if len(f.Balance) > 0 && f.Balance[0].TotalEquity != 0 {
		return f.Profile.MarketCap / float64(f.Balance[0].TotalEquity)
	}
	return 0
}

func (f *Fundamentals) derivedROE() float64 {
	if len(f.Income) > 0 && len(f.Balance) > 0 && f.Balance[0].TotalEquity != 0 {
		return float64(f.Income[0].NetIncome) / float64(f.Balance[0].TotalEquity)
	}
	return 0
}

func (f *Fundamentals) derivedDivYield() float64 {
	if f.Profile.Price != 0 {
		return f.Profile.LastDividend / f.Profile.Price
	}
	return 0
}

// ---- formatting ----

func esc(s string) string { return html.EscapeString(s) }

func tile(label, value string) string {
	return `<div class="tile"><div class="tl">` + esc(label) + `</div><div class="tv">` + value + `</div></div>`
}

func kv(k, v string) string {
	if v == "" {
		return ""
	}
	return `<tr><td class="k">` + esc(k) + `</td><td>` + esc(v) + `</td></tr>`
}

func money(v float64, currency string) string {
	sym := "$"
	if currency != "" && currency != "USD" {
		sym = esc(currency) + " "
	}
	return sym + fmt.Sprintf("%.2f", v)
}

func num2(v float64) string {
	if v == 0 {
		return "-"
	}
	return fmt.Sprintf("%.2f", v)
}

func pct(v float64) string {
	if v == 0 {
		return "-"
	}
	return fmt.Sprintf("%.1f%%", v*100)
}

// bigNum renders 1234567890 as 1.23B etc.
func bigNum(v float64) string {
	a := math.Abs(v)
	switch {
	case a >= 1e12:
		return fmt.Sprintf("%.2fT", v/1e12)
	case a >= 1e9:
		return fmt.Sprintf("%.2fB", v/1e9)
	case a >= 1e6:
		return fmt.Sprintf("%.2fM", v/1e6)
	case a >= 1e3:
		return fmt.Sprintf("%.1fK", v/1e3)
	case a == 0:
		return "-"
	default:
		return fmt.Sprintf("%.2f", v)
	}
}

// ---- static assets ----

const reportCSS = `
:root{--bg:#111110;--surface:#1a1a19;--card:#222221;--line:#33332f;--ink:#ffffff;--ink2:#c3c2b7;--ink3:#7a7974;--blue:#3987e5;--aqua:#199e70;--red:#e66767;--green:#4caf50}
*{box-sizing:border-box;margin:0;padding:0}
body{background:var(--bg);color:var(--ink);font:15px/1.5 -apple-system,"Segoe UI",Roboto,Helvetica,Arial,sans-serif;padding:24px 12px}
.wrap{max-width:960px;margin:0 auto}
.head{display:flex;align-items:center;gap:16px;margin-bottom:18px}
.logo{width:56px;height:56px;border-radius:12px;background:var(--card);object-fit:contain;padding:6px}
.logo.mono{display:flex;align-items:center;justify-content:center;font-size:26px;font-weight:700;color:var(--blue)}
.head-id{flex:1;min-width:0}
h1{font-size:22px;line-height:1.2}
.sub{color:var(--ink2);font-size:13px;margin-top:2px}
.head-px{text-align:right}
.px{font-size:24px;font-weight:700}
.chg{font-size:14px;font-weight:600}
.up{color:var(--green)}.down{color:var(--red)}
.tiles,.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(130px,1fr));gap:8px;margin-bottom:18px}
.grid{margin-bottom:0}
.tile{background:var(--card);border:1px solid var(--line);border-radius:10px;padding:10px 12px}
.tl{color:var(--ink3);font-size:11px;text-transform:uppercase;letter-spacing:.4px}
.tv{font-size:16px;font-weight:600;margin-top:2px}
input[name=tab]{display:none}
.tabbar{display:flex;gap:4px;border-bottom:1px solid var(--line);margin-bottom:16px;overflow-x:auto}
.tabbar label{padding:9px 14px;color:var(--ink2);cursor:pointer;border-radius:8px 8px 0 0;white-space:nowrap;font-size:14px}
.tabbar label:hover{color:var(--ink)}
.tab{display:none}
#tab-overview:checked~.tab-overview,#tab-financials:checked~.tab-financials,#tab-ratios:checked~.tab-ratios,#tab-revenue:checked~.tab-revenue,#tab-analysts:checked~.tab-analysts,#tab-price:checked~.tab-price{display:block}
#tab-overview:checked~.tabbar label[for=tab-overview],#tab-financials:checked~.tabbar label[for=tab-financials],#tab-ratios:checked~.tabbar label[for=tab-ratios],#tab-revenue:checked~.tabbar label[for=tab-revenue],#tab-analysts:checked~.tabbar label[for=tab-analysts],#tab-price:checked~.tabbar label[for=tab-price]{color:var(--ink);background:var(--card);box-shadow:inset 0 -2px 0 var(--blue)}
.cards{display:grid;grid-template-columns:repeat(auto-fit,minmax(300px,1fr));gap:12px}
.card{background:var(--surface);border:1px solid var(--line);border-radius:12px;padding:16px;margin-bottom:12px}
.card.grow{margin-bottom:0}
h2{font-size:15px;margin-bottom:10px;color:var(--ink)}
.desc{color:var(--ink2);font-size:13.5px;margin-bottom:10px;max-height:180px;overflow-y:auto}
.kv{width:100%;font-size:13px}
.kv td{padding:3px 0;border-bottom:1px solid var(--line)}
.kv .k{color:var(--ink3);width:35%}
table.data{width:100%;border-collapse:collapse;font-size:13px}
table.data th{color:var(--ink3);font-weight:600;text-align:right;padding:6px 8px;border-bottom:1px solid var(--line)}
table.data th:first-child,table.data td:first-child{text-align:left}
table.data td{padding:6px 8px;border-bottom:1px solid var(--line);color:var(--ink2)}
table.data td.n{text-align:right;font-variant-numeric:tabular-nums;color:var(--ink)}
svg{width:100%;height:auto;display:block}
svg text{font-family:inherit}
.legend{display:flex;gap:16px;margin-top:8px;font-size:12.5px;color:var(--ink2)}
.key{display:inline-flex;align-items:center;gap:6px}
.swatch{width:10px;height:10px;border-radius:3px;display:inline-block}
.segbars{display:flex;flex-direction:column;gap:8px}
.segrow{display:flex;align-items:center;gap:10px;font-size:13px}
.seglabel{width:30%;min-width:110px;color:var(--ink2);overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.segtrack{flex:1;background:var(--card);border-radius:6px;height:16px;overflow:hidden}
.segfill{display:block;height:100%;background:var(--blue);border-radius:6px 3px 3px 6px}
.segval{width:70px;text-align:right;font-variant-numeric:tabular-nums}
.muted{color:var(--ink3)}
footer{color:var(--ink3);font-size:12px;margin-top:20px;text-align:center}
#tip{position:fixed;pointer-events:none;background:#000;color:#fff;border:1px solid var(--line);border-radius:8px;padding:5px 9px;font-size:12.5px;opacity:0;transition:opacity .08s;z-index:9;max-width:260px}
p{color:var(--ink2);font-size:13.5px;margin-top:8px}
`

const tooltipJS = `<div id="tip"></div><script>
(function(){var tip=document.getElementById('tip');
document.addEventListener('mousemove',function(e){var t=e.target.closest('[data-tip]');
if(t){tip.textContent=t.getAttribute('data-tip');tip.style.opacity=1;
var x=Math.min(e.clientX+14,window.innerWidth-tip.offsetWidth-8);
tip.style.left=x+'px';tip.style.top=(e.clientY+16)+'px';}else{tip.style.opacity=0;}});
})();
</script>`
