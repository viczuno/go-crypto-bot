// Package markdown provides README generation for the crypto bot.
package markdown

import (
	"fmt"
	"html"
	"net/url"
	"strings"
	"time"

	"github.com/viczuno/go-crypto-bot/internal/domain"
)

// ReadmeBuilder implements domain.ReadmeGenerator.
type ReadmeBuilder struct{}

// NewReadmeBuilder creates a new README builder.
func NewReadmeBuilder() *ReadmeBuilder {
	return &ReadmeBuilder{}
}

var _ domain.ReadmeGenerator = (*ReadmeBuilder)(nil)

// Generate creates the README content from coin statistics.
func (b *ReadmeBuilder) Generate(stats []domain.CoinStats, coins []domain.CoinMetadata) string {
	var sb strings.Builder
	now := time.Now().UTC()

	b.writeHeader(&sb, now)
	b.writePriceTable(&sb, stats, coins)
	b.writeTechnicalAnalysis(&sb, stats, coins)
	b.writePerformanceChart(&sb, stats, coins)
	b.writeFooter(&sb)

	return sb.String()
}

func (b *ReadmeBuilder) writeHeader(sb *strings.Builder, now time.Time) {
	sb.WriteString("<div align=\"center\">\n\n")
	sb.WriteString("# 🚀 Crypto Market Tracker\n\n")
	sb.WriteString("[![Update Status](https://img.shields.io/badge/auto--update-every%2012h-brightgreen)]()\n")
	sb.WriteString("[![Data Source](https://img.shields.io/badge/data-CoinGecko-orange)](https://coingecko.com)\n")
	sb.WriteString("[![Built with Go](https://img.shields.io/badge/built%20with-Go-00ADD8?logo=go)](https://golang.org)\n\n")
	sb.WriteString("**Real-time cryptocurrency tracking powered by GitHub Actions**\n\n")
	fmt.Fprintf(sb, "🕐 *Last updated: %s*\n\n", now.Format("Monday, January 2, 2006 at 15:04 UTC"))
	sb.WriteString("</div>\n\n")
	sb.WriteString("---\n\n")
}

func (b *ReadmeBuilder) writePriceTable(sb *strings.Builder, stats []domain.CoinStats, coins []domain.CoinMetadata) {
	sb.WriteString("## 💰 Live Prices & Trends\n\n")
	sb.WriteString("<table>\n")
	sb.WriteString("<thead>\n")
	sb.WriteString("<tr>\n")
	sb.WriteString("<th align=\"left\">Asset</th>\n")
	sb.WriteString("<th align=\"right\">Price (USD)</th>\n")
	sb.WriteString("<th align=\"center\">24h</th>\n")
	sb.WriteString("<th align=\"center\">7 Days</th>\n")
	sb.WriteString("<th align=\"center\">30 Days</th>\n")
	sb.WriteString("</tr>\n")
	sb.WriteString("</thead>\n")
	sb.WriteString("<tbody>\n")

	coinMap := make(map[string]domain.CoinMetadata)
	for _, c := range coins {
		coinMap[c.ID] = c
	}

	for _, s := range stats {
		meta := coinMap[s.ID]
		priceStr := b.formatPrice(s.Price)
		change24h := b.formatChangeWithColor(s.Change24h)
		change7d := b.formatHistoricalChange(s.Change7d)
		change30d := b.formatHistoricalChange(s.Change30d)

		sb.WriteString("<tr>\n")
		fmt.Fprintf(sb, "<td><b>%s %s</b><br/></td>\n", html.EscapeString(meta.Name), html.EscapeString(meta.Symbol))
		fmt.Fprintf(sb, "<td align=\"right\"><code>%s</code></td>\n", priceStr)
		fmt.Fprintf(sb, "<td align=\"center\">%s</td>\n", change24h)
		fmt.Fprintf(sb, "<td align=\"center\">%s</td>\n", change7d)
		fmt.Fprintf(sb, "<td align=\"center\">%s</td>\n", change30d)
		sb.WriteString("</tr>\n")
	}

	sb.WriteString("</tbody>\n")
	sb.WriteString("</table>\n\n")
}

func (b *ReadmeBuilder) writePerformanceChart(sb *strings.Builder, stats []domain.CoinStats, coins []domain.CoinMetadata) {
	coinMap := make(map[string]domain.CoinMetadata)
	for _, c := range coins {
		coinMap[c.ID] = c
	}

	var labels []string
	var data []string
	var colors []string

	for _, s := range stats {
		meta := coinMap[s.ID]
		labels = append(labels, fmt.Sprintf("'%s'", meta.Symbol))
		data = append(data, fmt.Sprintf("%.2f", s.Change24h))
		if s.Change24h >= 0 {
			colors = append(colors, "'rgba(34, 197, 94, 0.8)'")
		} else {
			colors = append(colors, "'rgba(239, 68, 68, 0.8)'")
		}
	}

	chartConfig := fmt.Sprintf(`{
  type: 'bar',
  data: {
    labels: [%s],
    datasets: [{
      label: '24h Change',
      data: [%s],
      backgroundColor: [%s],
      borderRadius: 5
    }]
  },
  options: {
    plugins: {
      title: {display: true, text: '24-Hour Performance (%%)', font: {size: 16}},
      legend: {display: false}
    },
    scales: {
      y: {
        beginAtZero: true,
        grid: {color: 'rgba(0,0,0,0.1)'}
      }
    }
  }
}`, strings.Join(labels, ", "), strings.Join(data, ", "), strings.Join(colors, ", "))

	chartURL := "https://quickchart.io/chart?w=700&h=350&c=" + url.QueryEscape(chartConfig)

	sb.WriteString("## 24-Hour Performance\n\n")
	sb.WriteString("<div align=\"center\">\n\n")
	fmt.Fprintf(sb, "![24h Performance Chart](%s)\n\n", chartURL)
	sb.WriteString("</div>\n\n")
}

func (b *ReadmeBuilder) writeFooter(sb *strings.Builder) {
	sb.WriteString("---\n\n")
	sb.WriteString("<details>\n")
	sb.WriteString("<summary><b>ℹ️ About This Project</b></summary>\n\n")
	sb.WriteString("This automated tracker runs every 12 hours via GitHub Actions.\n\n")
	sb.WriteString("**Features:**\n")
	sb.WriteString("- Auto-updates twice daily\n")
	sb.WriteString("- Historical trend tracking using SQLite\n")
	sb.WriteString("- Dynamic chart generation\n")
	sb.WriteString("- No external server required\n\n")
	sb.WriteString("**Tech Stack:** Go • SQLite • GitHub Actions • CoinGecko API\n\n")
	sb.WriteString("</details>\n\n")
	sb.WriteString("<div align=\"center\">\n\n")
	sb.WriteString("*Data provided by [CoinGecko](https://coingecko.com)*\n\n")
	sb.WriteString("</div>\n")
}

func (b *ReadmeBuilder) formatPrice(price float64) string {
	if price >= 1 {
		return fmt.Sprintf("$%.2f", price)
	}
	return fmt.Sprintf("$%.4f", price)
}

func (b *ReadmeBuilder) formatChangeWithColor(change float64) string {
	if change > 0 {
		return fmt.Sprintf("🟢 +%.2f%%", change)
	} else if change < 0 {
		return fmt.Sprintf("🔴 %.2f%%", change)
	}
	return "⚪ 0.00%"
}

func (b *ReadmeBuilder) formatHistoricalChange(pc domain.PriceChange) string {
	if !pc.HasData {
		return "<sub>📊 Collecting...</sub>"
	}
	return b.formatChangeWithColor(pc.PctChange)
}

func (b *ReadmeBuilder) writeTechnicalAnalysis(sb *strings.Builder, stats []domain.CoinStats, coins []domain.CoinMetadata) {
	// Check if any coin has indicator data
	hasIndicators := false
	for _, s := range stats {
		if s.Indicators != nil && !s.Indicators.InsufficientData {
			hasIndicators = true
			break
		}
	}

	if !hasIndicators {
		return
	}

	coinMap := make(map[string]domain.CoinMetadata)
	for _, c := range coins {
		coinMap[c.ID] = c
	}

	sb.WriteString("## 📈 Technical Analysis\n\n")
	sb.WriteString("<table>\n")
	sb.WriteString("<thead>\n")
	sb.WriteString("<tr>\n")
	sb.WriteString("<th align=\"left\">Asset</th>\n")
	sb.WriteString("<th align=\"center\">Signal</th>\n")
	sb.WriteString("<th align=\"center\">RSI</th>\n")
	sb.WriteString("<th align=\"center\">MACD</th>\n")
	sb.WriteString("<th align=\"center\">MA</th>\n")
	sb.WriteString("<th align=\"center\">Bollinger</th>\n")
	sb.WriteString("</tr>\n")
	sb.WriteString("</thead>\n")
	sb.WriteString("<tbody>\n")

	for _, s := range stats {
		meta := coinMap[s.ID]

		if s.Indicators == nil || s.Indicators.InsufficientData {
			sb.WriteString("<tr>\n")
			fmt.Fprintf(sb, "<td><b>%s</b></td>\n", html.EscapeString(meta.Symbol))
			sb.WriteString("<td align=\"center\" colspan=\"5\"><sub>📊 Collecting data...</sub></td>\n")
			sb.WriteString("</tr>\n")
			continue
		}

		// Format consensus signal
		consensusStr := b.formatSignal(s.Indicators.Consensus, s.Indicators.Confidence)

		// Format individual indicators
		var rsiStr, macdStr, maStr, bbStr string
		for _, ind := range s.Indicators.Indicators {
			switch ind.Indicator {
			case domain.IndicatorRSI:
				rsiStr = b.formatIndicatorValue(ind, "%.1f")
			case domain.IndicatorMACD:
				macdStr = b.formatIndicatorSignal(ind)
			case domain.IndicatorMovingAverage:
				maStr = b.formatMAIndicator(ind)
			case domain.IndicatorBollingerBands:
				bbStr = b.formatBBIndicator(ind)
			}
		}

		sb.WriteString("<tr>\n")
		fmt.Fprintf(sb, "<td><b>%s</b></td>\n", html.EscapeString(meta.Symbol))
		fmt.Fprintf(sb, "<td align=\"center\">%s</td>\n", consensusStr)
		fmt.Fprintf(sb, "<td align=\"center\">%s</td>\n", rsiStr)
		fmt.Fprintf(sb, "<td align=\"center\">%s</td>\n", macdStr)
		fmt.Fprintf(sb, "<td align=\"center\">%s</td>\n", maStr)
		fmt.Fprintf(sb, "<td align=\"center\">%s</td>\n", bbStr)
		sb.WriteString("</tr>\n")
	}

	sb.WriteString("</tbody>\n")
	sb.WriteString("</table>\n\n")

	// Add legend
	sb.WriteString("<details>\n")
	sb.WriteString("<summary><b>📖 Signal Legend</b></summary>\n\n")
	sb.WriteString("| Signal | Meaning |\n")
	sb.WriteString("|--------|--------|\n")
	sb.WriteString("| 🟢 Strong Buy | Multiple indicators suggest strong upward momentum |\n")
	sb.WriteString("| 🔵 Buy | Indicators lean bullish |\n")
	sb.WriteString("| ⚪ Hold | Mixed or neutral signals |\n")
	sb.WriteString("| 🟠 Sell | Indicators lean bearish |\n")
	sb.WriteString("| 🔴 Strong Sell | Multiple indicators suggest strong downward momentum |\n\n")
	sb.WriteString("**Indicators used:** RSI (14), MACD (12/26/9), Moving Averages (50/200), Bollinger Bands (20, 2σ)\n\n")
	sb.WriteString("</details>\n\n")
}

func (b *ReadmeBuilder) formatSignal(signal domain.Signal, confidence float64) string {
	emoji := signal.Emoji()
	name := signal.String()
	confPct := int(confidence * 100)
	return fmt.Sprintf("%s <b>%s</b><br/><sub>%d%% conf</sub>", emoji, name, confPct)
}

func (b *ReadmeBuilder) formatIndicatorValue(ind domain.IndicatorResult, format string) string {
	if ind.Error != nil {
		return "<sub>N/A</sub>"
	}
	emoji := ind.Signal.Emoji()
	value := fmt.Sprintf(format, ind.Value)
	return fmt.Sprintf("%s %s", emoji, value)
}

func (b *ReadmeBuilder) formatIndicatorSignal(ind domain.IndicatorResult) string {
	if ind.Error != nil {
		return "<sub>N/A</sub>"
	}
	emoji := ind.Signal.Emoji()
	hist := ind.Metadata["histogram"]
	var direction string
	if hist > 0 {
		direction = "↑"
	} else if hist < 0 {
		direction = "↓"
	} else {
		direction = "→"
	}
	return fmt.Sprintf("%s %s", emoji, direction)
}

func (b *ReadmeBuilder) formatMAIndicator(ind domain.IndicatorResult) string {
	if ind.Error != nil {
		return "<sub>N/A</sub>"
	}
	emoji := ind.Signal.Emoji()

	// Check for golden/death cross
	if ind.Metadata["golden_cross"] == 1.0 {
		return fmt.Sprintf("%s ✨ Golden", emoji)
	}
	if ind.Metadata["death_cross"] == 1.0 {
		return fmt.Sprintf("%s ☠️ Death", emoji)
	}

	// Show trend direction
	sma50 := ind.Metadata["sma_50"]
	sma200 := ind.Metadata["sma_200"]
	if sma50 > sma200 {
		return fmt.Sprintf("%s ↑ Bullish", emoji)
	}
	return fmt.Sprintf("%s ↓ Bearish", emoji)
}

func (b *ReadmeBuilder) formatBBIndicator(ind domain.IndicatorResult) string {
	if ind.Error != nil {
		return "<sub>N/A</sub>"
	}
	emoji := ind.Signal.Emoji()
	percentB := ind.Metadata["percent_b"]

	var position string
	switch {
	case percentB <= 0.2:
		position = "Lower"
	case percentB >= 0.8:
		position = "Upper"
	default:
		position = "Middle"
	}

	return fmt.Sprintf("%s %s", emoji, position)
}
