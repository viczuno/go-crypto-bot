<div align="center">

# 🚀 Crypto Market Tracker

[![Update Status](https://img.shields.io/badge/auto--update-every%2012h-brightgreen)]()
[![Data Source](https://img.shields.io/badge/data-CoinGecko-orange)](https://coingecko.com)
[![Built with Go](https://img.shields.io/badge/built%20with-Go-00ADD8?logo=go)](https://golang.org)

**Real-time cryptocurrency tracking powered by GitHub Actions**

🕐 *Last updated: Thursday, March 12, 2026 at 01:03 UTC*

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
<td align="right"><code>$70167.00</code></td>
<td align="center">🟢 +0.18%</td>
<td align="center">🔴 -0.66%</td>
<td align="center">🟢 +0.10%</td>
</tr>
<tr>
<td><b>Ethereum ETH</b><br/></td>
<td align="right"><code>$2054.22</code></td>
<td align="center">🟢 +0.80%</td>
<td align="center">🟢 +0.77%</td>
<td align="center">🔴 -2.39%</td>
</tr>
<tr>
<td><b>Solana SOL</b><br/></td>
<td align="right"><code>$86.66</code></td>
<td align="center">🟢 +0.71%</td>
<td align="center">🔴 -2.60%</td>
<td align="center">🔴 -0.20%</td>
</tr>
<tr>
<td><b>Cardano ADA</b><br/></td>
<td align="right"><code>$0.2620</code></td>
<td align="center">🔴 -0.06%</td>
<td align="center">🔴 -1.68%</td>
<td align="center">🔴 -2.93%</td>
</tr>
<tr>
<td><b>Polkadot DOT</b><br/></td>
<td align="right"><code>$1.52</code></td>
<td align="center">🟢 +2.38%</td>
<td align="center">⚪ 0.00%</td>
<td align="center">🟢 +15.10%</td>
</tr>
</tbody>
</table>

## 24-Hour Performance

<div align="center">

![24h Performance Chart](https://quickchart.io/chart?w=700&h=350&c=%7B%0A++type%3A+%27bar%27%2C%0A++data%3A+%7B%0A++++labels%3A+%5B%27BTC%27%2C+%27ETH%27%2C+%27SOL%27%2C+%27ADA%27%2C+%27DOT%27%5D%2C%0A++++datasets%3A+%5B%7B%0A++++++label%3A+%2724h+Change%27%2C%0A++++++data%3A+%5B0.18%2C+0.80%2C+0.71%2C+-0.06%2C+2.38%5D%2C%0A++++++backgroundColor%3A+%5B%27rgba%2834%2C+197%2C+94%2C+0.8%29%27%2C+%27rgba%2834%2C+197%2C+94%2C+0.8%29%27%2C+%27rgba%2834%2C+197%2C+94%2C+0.8%29%27%2C+%27rgba%28239%2C+68%2C+68%2C+0.8%29%27%2C+%27rgba%2834%2C+197%2C+94%2C+0.8%29%27%5D%2C%0A++++++borderRadius%3A+5%0A++++%7D%5D%0A++%7D%2C%0A++options%3A+%7B%0A++++plugins%3A+%7B%0A++++++title%3A+%7Bdisplay%3A+true%2C+text%3A+%2724-Hour+Performance+%28%25%29%27%2C+font%3A+%7Bsize%3A+16%7D%7D%2C%0A++++++legend%3A+%7Bdisplay%3A+false%7D%0A++++%7D%2C%0A++++scales%3A+%7B%0A++++++y%3A+%7B%0A++++++++beginAtZero%3A+true%2C%0A++++++++grid%3A+%7Bcolor%3A+%27rgba%280%2C0%2C0%2C0.1%29%27%7D%0A++++++%7D%0A++++%7D%0A++%7D%0A%7D)

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
