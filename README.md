<div align="center">

# 🚀 Crypto Market Tracker

[![Update Status](https://img.shields.io/badge/auto--update-every%2012h-brightgreen)]()
[![Data Source](https://img.shields.io/badge/data-CoinGecko-orange)](https://coingecko.com)
[![Built with Go](https://img.shields.io/badge/built%20with-Go-00ADD8?logo=go)](https://golang.org)

**Real-time cryptocurrency tracking powered by GitHub Actions**

[![View Dashboard](https://img.shields.io/badge/View-Live%20Dashboard-blue?style=for-the-badge)](https://viczuno.github.io/go-crypto-bot/)

🕐 *Last updated: Friday, May 22, 2026 at 02:11 UTC*

</div>

---

## 💰 Live Prices & Trends

<table>
<thead>
<tr>
<th align="left">Asset</th>
<th align="right">Price (USD)</th>
<th align="center">24h</th>
<th align="center">7 Days</th>
<th align="center">30 Days</th>
</tr>
</thead>
<tbody>
<tr>
<td><b>Bitcoin BTC</b><br/></td>
<td align="right"><code>$77589.00</code></td>
<td align="center">🔴 -0.25%</td>
<td align="center">🔴 -2.87%</td>
<td align="center">🟢 +2.11%</td>
</tr>
<tr>
<td><b>Ethereum ETH</b><br/></td>
<td align="right"><code>$2135.55</code></td>
<td align="center">🔴 -0.16%</td>
<td align="center">🔴 -5.53%</td>
<td align="center">🔴 -7.46%</td>
</tr>
<tr>
<td><b>Solana SOL</b><br/></td>
<td align="right"><code>$87.04</code></td>
<td align="center">🟢 +0.73%</td>
<td align="center">🔴 -4.65%</td>
<td align="center">🟢 +1.81%</td>
</tr>
<tr>
<td><b>Cardano ADA</b><br/></td>
<td align="right"><code>$0.2507</code></td>
<td align="center">🟢 +0.05%</td>
<td align="center">🔴 -6.19%</td>
<td align="center">🟢 +0.81%</td>
</tr>
<tr>
<td><b>Polkadot DOT</b><br/></td>
<td align="right"><code>$1.30</code></td>
<td align="center">🟢 +3.18%</td>
<td align="center">🔴 -2.99%</td>
<td align="center">🟢 +1.56%</td>
</tr>
</tbody>
</table>

## 📈 Technical Analysis

<table>
<thead>
<tr>
<th align="left">Asset</th>
<th align="center">Signal</th>
<th align="center">RSI</th>
<th align="center">MACD</th>
<th align="center">MA</th>
<th align="center">Bollinger</th>
</tr>
</thead>
<tbody>
<tr>
<td><b>BTC</b></td>
<td align="center">⚪ <b>Hold</b><br/><sub>56% conf</sub></td>
<td align="center">🟠 76.0</td>
<td align="center">⚪ ↓</td>
<td align="center">🔵 ↑ Bullish</td>
<td align="center">⚪ Middle</td>
</tr>
<tr>
<td><b>ETH</b></td>
<td align="center">⚪ <b>Hold</b><br/><sub>64% conf</sub></td>
<td align="center">🟠 77.0</td>
<td align="center">⚪ ↓</td>
<td align="center">🔵 ↑ Bullish</td>
<td align="center">🟠 Upper</td>
</tr>
<tr>
<td><b>SOL</b></td>
<td align="center">⚪ <b>Hold</b><br/><sub>61% conf</sub></td>
<td align="center">🟠 73.4</td>
<td align="center">⚪ ↓</td>
<td align="center">🔵 ↑ Bullish</td>
<td align="center">🟠 Upper</td>
</tr>
<tr>
<td><b>ADA</b></td>
<td align="center">⚪ <b>Hold</b><br/><sub>56% conf</sub></td>
<td align="center">⚪ 67.8</td>
<td align="center">⚪ ↓</td>
<td align="center">🔵 ↑ Bullish</td>
<td align="center">🟠 Upper</td>
</tr>
<tr>
<td><b>DOT</b></td>
<td align="center">⚪ <b>Hold</b><br/><sub>63% conf</sub></td>
<td align="center">🟠 78.1</td>
<td align="center">⚪ ↓</td>
<td align="center">🔵 ↑ Bullish</td>
<td align="center">🟠 Upper</td>
</tr>
</tbody>
</table>

<details>
<summary><b>📖 Signal Legend</b></summary>

| Signal | Meaning |
|--------|--------|
| 🟢 Strong Buy | Multiple indicators suggest strong upward momentum |
| 🔵 Buy | Indicators lean bullish |
| ⚪ Hold | Mixed or neutral signals |
| 🟠 Sell | Indicators lean bearish |
| 🔴 Strong Sell | Multiple indicators suggest strong downward momentum |

**Indicators used:** RSI (14), MACD (12/26/9), Moving Averages (50/200), Bollinger Bands (20, 2σ)

</details>

## 24-Hour Performance

<div align="center">

![24h Performance Chart](https://quickchart.io/chart?w=700&h=350&c=%7B%0A++type%3A+%27bar%27%2C%0A++data%3A+%7B%0A++++labels%3A+%5B%27BTC%27%2C+%27ETH%27%2C+%27SOL%27%2C+%27ADA%27%2C+%27DOT%27%5D%2C%0A++++datasets%3A+%5B%7B%0A++++++label%3A+%2724h+Change%27%2C%0A++++++data%3A+%5B-0.25%2C+-0.16%2C+0.73%2C+0.05%2C+3.18%5D%2C%0A++++++backgroundColor%3A+%5B%27rgba%28239%2C+68%2C+68%2C+0.8%29%27%2C+%27rgba%28239%2C+68%2C+68%2C+0.8%29%27%2C+%27rgba%2834%2C+197%2C+94%2C+0.8%29%27%2C+%27rgba%2834%2C+197%2C+94%2C+0.8%29%27%2C+%27rgba%2834%2C+197%2C+94%2C+0.8%29%27%5D%2C%0A++++++borderRadius%3A+5%0A++++%7D%5D%0A++%7D%2C%0A++options%3A+%7B%0A++++plugins%3A+%7B%0A++++++title%3A+%7Bdisplay%3A+true%2C+text%3A+%2724-Hour+Performance+%28%25%29%27%2C+font%3A+%7Bsize%3A+16%7D%7D%2C%0A++++++legend%3A+%7Bdisplay%3A+false%7D%0A++++%7D%2C%0A++++scales%3A+%7B%0A++++++y%3A+%7B%0A++++++++beginAtZero%3A+true%2C%0A++++++++grid%3A+%7Bcolor%3A+%27rgba%280%2C0%2C0%2C0.1%29%27%7D%0A++++++%7D%0A++++%7D%0A++%7D%0A%7D)

</div>

---

<details>
<summary><b>ℹ️ About This Project</b></summary>

This automated tracker runs every 12 hours via GitHub Actions.

**Features:**
- Auto-updates twice daily
- Historical trend tracking using SQLite
- Dynamic chart generation
- No external server required

**Tech Stack:** Go • SQLite • GitHub Actions • CoinGecko API

</details>

<div align="center">

*Data provided by [CoinGecko](https://coingecko.com)*

</div>
