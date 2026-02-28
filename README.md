<div align="center">

# 🚀 Crypto Market Tracker

[![Update Status](https://img.shields.io/badge/auto--update-every%2012h-brightgreen)]()
[![Data Source](https://img.shields.io/badge/data-CoinGecko-orange)](https://coingecko.com)
[![Built with Go](https://img.shields.io/badge/built%20with-Go-00ADD8?logo=go)](https://golang.org)

**Real-time cryptocurrency tracking powered by GitHub Actions**

🕐 *Last updated: Saturday, February 28, 2026 at 12:24 UTC*

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
<td align="right"><code>$64025.00</code></td>
<td align="center">🔴 -2.95%</td>
<td align="center">🔴 -5.64%</td>
<td align="center">🔴 -28.19%</td>
</tr>
<tr>
<td><b>Ethereum ETH</b><br/></td>
<td align="right"><code>$1871.17</code></td>
<td align="center">🔴 -4.79%</td>
<td align="center">🔴 -4.73%</td>
<td align="center">🔴 -37.77%</td>
</tr>
<tr>
<td><b>Solana SOL</b><br/></td>
<td align="right"><code>$79.10</code></td>
<td align="center">🔴 -5.18%</td>
<td align="center">🔴 -6.32%</td>
<td align="center">🔴 -36.74%</td>
</tr>
<tr>
<td><b>Cardano ADA</b><br/></td>
<td align="right"><code>$0.2635</code></td>
<td align="center">🔴 -6.74%</td>
<td align="center">🔴 -6.76%</td>
<td align="center">🔴 -26.38%</td>
</tr>
<tr>
<td><b>Polkadot DOT</b><br/></td>
<td align="right"><code>$1.49</code></td>
<td align="center">🔴 -4.57%</td>
<td align="center">🟢 +12.03%</td>
<td align="center">🔴 -20.31%</td>
</tr>
</tbody>
</table>

## 24-Hour Performance

<div align="center">

![24h Performance Chart](https://quickchart.io/chart?w=700&h=350&c=%7B%0A++type%3A+%27bar%27%2C%0A++data%3A+%7B%0A++++labels%3A+%5B%27BTC%27%2C+%27ETH%27%2C+%27SOL%27%2C+%27ADA%27%2C+%27DOT%27%5D%2C%0A++++datasets%3A+%5B%7B%0A++++++label%3A+%2724h+Change%27%2C%0A++++++data%3A+%5B-2.95%2C+-4.79%2C+-5.18%2C+-6.74%2C+-4.57%5D%2C%0A++++++backgroundColor%3A+%5B%27rgba%28239%2C+68%2C+68%2C+0.8%29%27%2C+%27rgba%28239%2C+68%2C+68%2C+0.8%29%27%2C+%27rgba%28239%2C+68%2C+68%2C+0.8%29%27%2C+%27rgba%28239%2C+68%2C+68%2C+0.8%29%27%2C+%27rgba%28239%2C+68%2C+68%2C+0.8%29%27%5D%2C%0A++++++borderRadius%3A+5%0A++++%7D%5D%0A++%7D%2C%0A++options%3A+%7B%0A++++plugins%3A+%7B%0A++++++title%3A+%7Bdisplay%3A+true%2C+text%3A+%2724-Hour+Performance+%28%25%29%27%2C+font%3A+%7Bsize%3A+16%7D%7D%2C%0A++++++legend%3A+%7Bdisplay%3A+false%7D%0A++++%7D%2C%0A++++scales%3A+%7B%0A++++++y%3A+%7B%0A++++++++beginAtZero%3A+true%2C%0A++++++++grid%3A+%7Bcolor%3A+%27rgba%280%2C0%2C0%2C0.1%29%27%7D%0A++++++%7D%0A++++%7D%0A++%7D%0A%7D)

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
