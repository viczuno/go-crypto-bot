package markdown

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/viczuno/go-crypto-bot/internal/domain"
)

// ReadmeBuilder implements domain.ReadmeGenerator
type ReadmeBuilder struct{}

// NewReadmeBuilder creates a new README builder
func NewReadmeBuilder() *ReadmeBuilder {
	return &ReadmeBuilder{}
}

// Generate creates the README content from coin statistics
func (b *ReadmeBuilder) Generate(stats []domain.CoinStats, coins []domain.CoinMetadata) string {
	var sb strings.Builder
	now := time.Now().UTC()

	b.writeHeader(&sb, now)
	b.writeMarketOverview(&sb, stats, coins)
	b.writePriceTable(&sb, stats, coins)
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
	sb.WriteString(fmt.Sprintf("🕐 *Last updated: %s*\n\n", now.Format("Monday, January 2, 2006 at 15:04 UTC")))
	sb.WriteString("</div>\n\n")
	sb.WriteString("---\n\n")
}

func (b *ReadmeBuilder) writeMarketOverview(sb *strings.Builder, stats []domain.CoinStats, coins []domain.CoinMetadata) {
	// Calculate market sentiment
	gainers := 0
	losers := 0
	totalChange := 0.0

	for _, s := range stats {
		totalChange += s.Change24h
		if s.Change24h > 0 {
			gainers++
		} else {
			losers--
		}
	}

	avgChange := totalChange / float64(len(stats))
	sentiment := "🔴 Bearish"
	if avgChange > 2 {
		sentiment = "🟢 Bullish"
	} else if avgChange > 0 {
		sentiment = "🟡 Neutral"
	}

	sb.WriteString("## 📊 Market Overview\n\n")
	sb.WriteString("<table>\n<tr>\n")
	sb.WriteString(fmt.Sprintf("<td align=\"center\"><b>Market Sentiment</b><br/>%s</td>\n", sentiment))
	sb.WriteString(fmt.Sprintf("<td align=\"center\"><b>Avg 24h Change</b><br/>%s</td>\n", b.formatChangeWithColor(avgChange)))
	sb.WriteString(fmt.Sprintf("<td align=\"center\"><b>Gainers</b><br/>🟢 %d</td>\n", gainers))
	sb.WriteString(fmt.Sprintf("<td align=\"center\"><b>Losers</b><br/>🔴 %d</td>\n", len(stats)-gainers))
	sb.WriteString("</tr>\n</table>\n\n")
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
	sb.WriteString("<th align=\"center\">Trend</th>\n")
	sb.WriteString("</tr>\n")
	sb.WriteString("</thead>\n")
	sb.WriteString("<tbody>\n")

	coinMap := make(map[string]domain.CoinMetadata)
	for _, c := range coins {
		coinMap[c.ID] = c
	}

	for _, s := range stats {
		meta := coinMap[s.Name]

		// Format price with proper formatting
		priceStr := b.formatPrice(s.Price)

		// Format changes
		change24h := b.formatChangeWithColor(s.Change24h)
		change7d := b.formatHistoricalChange(s.Change7d)
		change30d := b.formatHistoricalChange(s.Change30d)

		// Determine trend
		trend := b.calculateTrend(s)

		sb.WriteString("<tr>\n")
		sb.WriteString(fmt.Sprintf("<td><b>%s %s</b><br/><sub>%s</sub></td>\n", meta.Name, meta.Symbol))
		sb.WriteString(fmt.Sprintf("<td align=\"right\"><code>%s</code></td>\n", priceStr))
		sb.WriteString(fmt.Sprintf("<td align=\"center\">%s</td>\n", change24h))
		sb.WriteString(fmt.Sprintf("<td align=\"center\">%s</td>\n", change7d))
		sb.WriteString(fmt.Sprintf("<td align=\"center\">%s</td>\n", change30d))
		sb.WriteString(fmt.Sprintf("<td align=\"center\">%s</td>\n", trend))
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
		meta := coinMap[s.Name]
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
	sb.WriteString(fmt.Sprintf("![24h Performance Chart](%s)\n\n", chartURL))
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
	if price >= 1000 {
		return fmt.Sprintf("$%.2f", price)
	} else if price >= 1 {
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

func (b *ReadmeBuilder) calculateTrend(s domain.CoinStats) string {
	// Simple trend calculation based on available data
	score := 0

	if s.Change24h > 0 {
		score++
	} else if s.Change24h < 0 {
		score--
	}

	if s.Change7d.HasData {
		if s.Change7d.PctChange > 0 {
			score++
		} else if s.Change7d.PctChange < 0 {
			score--
		}
	}

	if s.Change30d.HasData {
		if s.Change30d.PctChange > 0 {
			score++
		} else if s.Change30d.PctChange < 0 {
			score--
		}
	}

	switch {
	case score >= 2:
		return "📈"
	case score <= -2:
		return "📉"
	default:
		return "➡️"
	}
}

// Ensure ReadmeBuilder implements ReadmeGenerator
var _ domain.ReadmeGenerator = (*ReadmeBuilder)(nil)
