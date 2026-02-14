package markdown

import (
	"fmt"
	"strings"
	"time"

	"github.com/viczuno/go-crypto-bot.git/internal/db"
)

type CoinStats struct {
	Name      string
	Price     float64
	Change24h float64
	Change7d  db.PriceChange
	Change30d db.PriceChange
}

func BuildReadme(stats []CoinStats) string {
	var sb strings.Builder
	now := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 UTC")

	sb.WriteString("<div align=\"center\">\n\n")
	sb.WriteString("# 📈 Automated Crypto Market Tracker\n\n")
	sb.WriteString(fmt.Sprintf("*Last updated: %s*\n\n", now))
	sb.WriteString("This repository runs a Go binary via GitHub Actions every 12 hours. It uses **Git as a Database** (SQLite) to track long-term historical trends without requiring a dedicated server.\n\n")
	sb.WriteString("</div>\n\n")

	sb.WriteString("### 📊 Live Prices & Historical Trends\n\n")
	sb.WriteString("| Asset | Price (USD) | 24h Change | 7-Day Change | 30-Day Change |\n")
	sb.WriteString("| :--- | :--- | :--- | :--- | :--- |\n")

	var labels []string
	var changes []float64

	for _, s := range stats {
		c24 := formatChange(s.Change24h, s.Change24h)

		c7 := "*(Gathering Data...)*"
		if s.Change7d.HasData {
			c7 = formatChange(s.Change7d.AbsChange, s.Change7d.PctChange)
		}

		c30 := "*(Gathering Data...)*"
		if s.Change30d.HasData {
			c30 = formatChange(s.Change30d.AbsChange, s.Change30d.PctChange)
		}

		sb.WriteString(fmt.Sprintf("| **%s** | $%.2f | %s | %s | %s |\n", strings.Title(s.Name), s.Price, c24, c7, c30))

		labels = append(labels, fmt.Sprintf("'%s'", strings.Title(s.Name)))
		changes = append(changes, s.Change24h)
	}

	chartURL := generateChartURL(labels, changes)
	sb.WriteString("\n### 📉 24-Hour Performance Visualization\n")
	sb.WriteString(fmt.Sprintf("![Crypto Chart](%s)\n", chartURL))

	return sb.String()
}

func formatChange(absolute, percent float64) string {
	emoji := "🔴"
	if percent > 0 {
		emoji = "🟢"
	}
	return fmt.Sprintf("%s %.2f%%", emoji, percent)
}

func generateChartURL(labels []string, data []float64) string {
	chartConfig := fmt.Sprintf(`{
		type: 'bar',
		data: {
			labels: [%s],
			datasets: [{
				label: '24h Change (%%)',
				data: [%s],
				backgroundColor: 'rgba(54, 162, 235, 0.6)',
				borderColor: 'rgba(54, 162, 235, 1)',
				borderWidth: 1
			}]
		},
		options: { title: { display: true, text: '24h Market Fluctuations' }, legend: { display: false } }
	}`, strings.Join(labels, ","), strings.Trim(strings.Join(strings.Fields(fmt.Sprint(data)), ","), "[]"))

	encoded := strings.ReplaceAll(chartConfig, " ", "%20")
	encoded = strings.ReplaceAll(encoded, "\n", "")
	encoded = strings.ReplaceAll(encoded, "\t", "")
	return "https://quickchart.io/chart?w=600&h=300&c=" + encoded
}
