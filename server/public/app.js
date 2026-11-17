// ==========================================================================
// OME // High-Frequency Order Matching Engine & Terminal Controller
// ==========================================================================

const SYMBOLS = ['AAPL', 'TSLA', 'BTC-USD'];

function emptyStats() {
  return { high: null, low: null, volume: 0, tradesCount: 0, openPrice: null, lastPrice: null };
}

const state = {
  activeSymbol: 'AAPL',
  matchMode: 'audit', // 'audit' | 'tape'
  side: 0, // 0 = Buy, 1 = Sell
  type: 0, // 0 = Limit, 1 = Market
  myOrders: [],
  ws: null,
  botRunning: false,
  statsBySymbol: Object.fromEntries(SYMBOLS.map(s => [s, emptyStats()])),
  recentTrades: Object.fromEntries(SYMBOLS.map(s => [s, []])),
  earlyMakerFills: new Map(),
  lastBookBySymbol: Object.fromEntries(SYMBOLS.map(s => [s, { bids: [], asks: [] }])),
  hoverPriceChart: null,
};

const REFERENCE_PRICES = {
  'AAPL': '150.00',
  'TSLA': '240.00',
  'BTC-USD': '64000.00',
};

// DOM Elements
const symbolTabs = document.getElementById('symbolTabs');
const activeSymbolTag = document.getElementById('activeSymbolTag');
const btnSideBuy = document.getElementById('btnSideBuy');
const btnSideSell = document.getElementById('btnSideSell');
const btnTypeLimit = document.getElementById('btnTypeLimit');
const btnTypeMarket = document.getElementById('btnTypeMarket');
const priceGroup = document.getElementById('priceGroup');
const inputPrice = document.getElementById('inputPrice');
const inputAmount = document.getElementById('inputAmount');
const orderTotal = document.getElementById('orderTotal');
const btnSubmitOrder = document.getElementById('btnSubmitOrder');
const asksContainer = document.getElementById('asksContainer');
const bidsContainer = document.getElementById('bidsContainer');
const spreadValue = document.getElementById('spreadValue');
const spreadBps = document.getElementById('spreadBps');
const lastTradedPrice = document.getElementById('lastTradedPrice');
const tradesStream = document.getElementById('tradesStream');
const btnModeAudit = document.getElementById('btnModeAudit');
const btnModeTape = document.getElementById('btnModeTape');
const tradesTableHeader = document.getElementById('tradesTableHeader');

// Modal Elements
const matchModalBackdrop = document.getElementById('matchModalBackdrop');
const modalCloseBtn = document.getElementById('modalCloseBtn');
const modalTradeTitle = document.getElementById('modalTradeTitle');
const modalTakerId = document.getElementById('modalTakerId');
const modalTakerSideBadge = document.getElementById('modalTakerSideBadge');
const modalConnectorPrice = document.getElementById('modalConnectorPrice');
const modalMakerId = document.getElementById('modalMakerId');
const modalSettlementVal = document.getElementById('modalSettlementVal');
const modalSharesVal = document.getElementById('modalSharesVal');
const modalTicksVal = document.getElementById('modalTicksVal');
const ordersTableBody = document.getElementById('ordersTableBody');
const openOrdersCount = document.getElementById('openOrdersCount');
const wsStatus = document.getElementById('wsStatus');

// Chart Elements
const priceCanvas = document.getElementById('priceChartCanvas');
const ohlcOpen = document.getElementById('ohlcOpen');
const ohlcHigh = document.getElementById('ohlcHigh');
const ohlcLow = document.getElementById('ohlcLow');
const ohlcClose = document.getElementById('ohlcClose');
const ohlcVol = document.getElementById('ohlcVol');

// Simulator & Ticker Elements
const btnResetMarket = document.getElementById('btnResetMarket');
const btnToggleBot = document.getElementById('btnToggleBot');
const botLabel = document.getElementById('botLabel');
const tickerLastPrice = document.getElementById('tickerLastPrice');
const tickerChange = document.getElementById('tickerChange');
const tickerHigh = document.getElementById('tickerHigh');
const tickerLow = document.getElementById('tickerLow');
const tickerVolume = document.getElementById('tickerVolume');
const tickerTradesCount = document.getElementById('tickerTradesCount');

// --------------------------------------------------------------------------
// 1. Candlestick Data Store & Live Time-Bucketed Updates
// --------------------------------------------------------------------------
const CANDLE_PERIOD_MS = 15000; // 15-second candles
const MAX_CANDLES = 80;
const candles = Object.fromEntries(SYMBOLS.map(s => [s, []]));

function updateCandleOnTrade(symbol, price, amount, tsMs) {
  const symCandles = candles[symbol];
  const bucketTime = Math.floor(tsMs / CANDLE_PERIOD_MS) * CANDLE_PERIOD_MS;
  const current = symCandles.length > 0 ? symCandles[symCandles.length - 1] : null;

  if (current && current.time === bucketTime) {
    current.high = Math.max(current.high, price);
    current.low = Math.min(current.low, price);
    current.close = price;
    current.volume += amount;
  } else {
    symCandles.push({
      time: bucketTime,
      open: price,
      high: price,
      low: price,
      close: price,
      volume: amount,
    });
    if (symCandles.length > MAX_CANDLES) {
      symCandles.shift();
    }
  }

  if (symbol === state.activeSymbol) {
    updateOHLCHeader();
    renderPriceChart();
  }
}

function updateOHLCHeader(customCandle = null) {
  const symCandles = candles[state.activeSymbol];
  const c = customCandle || (symCandles.length > 0 ? symCandles[symCandles.length - 1] : null);

  if (!c) {
    ohlcOpen.textContent = '—';
    ohlcHigh.textContent = '—';
    ohlcLow.textContent = '—';
    ohlcClose.textContent = '—';
    ohlcVol.textContent = '—';
    return;
  }

  ohlcOpen.textContent = `$${c.open.toFixed(2)}`;
  ohlcHigh.textContent = `$${c.high.toFixed(2)}`;
  ohlcLow.textContent = `$${c.low.toFixed(2)}`;
  ohlcClose.textContent = `$${c.close.toFixed(2)}`;
  ohlcClose.style.color = c.close >= c.open ? 'var(--color-buy)' : 'var(--color-sell)';
  ohlcVol.textContent = c.volume.toLocaleString();
}

// --------------------------------------------------------------------------
// 2. High-DPI Canvas Candlestick & Volume Chart Engine
// --------------------------------------------------------------------------
function setupCanvasDPI(canvas) {
  const rect = canvas.getBoundingClientRect();
  const dpr = window.devicePixelRatio || 1;
  const width = Math.floor(rect.width);
  const height = Math.floor(rect.height);

  if (canvas.width !== width * dpr || canvas.height !== height * dpr) {
    canvas.width = width * dpr;
    canvas.height = height * dpr;
  }

  const ctx = canvas.getContext('2d');
  ctx.resetTransform();
  ctx.scale(dpr, dpr);
  return { ctx, width, height };
}

function renderPriceChart() {
  if (!priceCanvas) return;
  const { ctx, width, height } = setupCanvasDPI(priceCanvas);
  ctx.clearRect(0, 0, width, height);

  const symCandles = candles[state.activeSymbol] || [];
  if (symCandles.length === 0) {
    ctx.fillStyle = '#64748b';
    ctx.font = '12px "Plus Jakarta Sans", sans-serif';
    ctx.textAlign = 'center';
    ctx.fillText('Awaiting live market trade flow...', width / 2, height / 2);
    return;
  }

  const paddingRight = 68;
  const paddingBottom = 22;
  const chartW = width - paddingRight;
  const chartH = height - paddingBottom;
  const volH = Math.max(36, Math.floor(chartH * 0.22));
  const priceH = chartH - volH;

  let minPrice = Infinity;
  let maxPrice = -Infinity;
  let maxVolume = 1;

  for (const c of symCandles) {
    if (c.low < minPrice) minPrice = c.low;
    if (c.high > maxPrice) maxPrice = c.high;
    if (c.volume > maxVolume) maxVolume = c.volume;
  }

  const priceRange = maxPrice - minPrice;
  const pad = priceRange > 0 ? priceRange * 0.08 : (minPrice * 0.01 || 1);
  minPrice -= pad;
  maxPrice += pad;

  const priceToY = p => priceH - ((p - minPrice) / (maxPrice - minPrice)) * priceH;
  const yToPrice = y => maxPrice - (y / priceH) * (maxPrice - minPrice);

  // Horizontal Grid Lines & Price Ticks
  const gridSteps = 4;
  ctx.strokeStyle = 'rgba(255, 255, 255, 0.04)';
  ctx.lineWidth = 1;
  ctx.font = '10px "JetBrains Mono", monospace';
  ctx.fillStyle = '#64748b';
  ctx.textAlign = 'left';

  for (let i = 0; i <= gridSteps; i++) {
    const y = (priceH / gridSteps) * i;
    const priceVal = yToPrice(y);

    ctx.beginPath();
    ctx.moveTo(0, y);
    ctx.lineTo(chartW, y);
    ctx.stroke();

    ctx.fillText(`$${priceVal.toFixed(2)}`, chartW + 8, y + 3);
  }

  // Volume separator line
  ctx.strokeStyle = 'rgba(255, 255, 255, 0.06)';
  ctx.beginPath();
  ctx.moveTo(0, priceH);
  ctx.lineTo(chartW, priceH);
  ctx.stroke();

  // Draw Candlesticks & Volume Bars
  const totalSlots = Math.max(symCandles.length, 30);
  const slotW = chartW / totalSlots;
  const candleW = Math.max(2, slotW * 0.72);
  const startX = chartW - symCandles.length * slotW;

  symCandles.forEach((c, idx) => {
    const x = startX + idx * slotW + slotW / 2;
    const isUp = c.close >= c.open;
    const color = isUp ? '#05d58c' : '#ff4766';
    const volColor = isUp ? 'rgba(5, 213, 140, 0.28)' : 'rgba(255, 71, 102, 0.28)';

    // Volume Bar
    const vHeight = (c.volume / maxVolume) * (volH - 6);
    const vY = height - paddingBottom - vHeight;
    ctx.fillStyle = volColor;
    ctx.fillRect(x - candleW / 2, vY, candleW, vHeight);

    // Candle Wick
    const highY = priceToY(c.high);
    const lowY = priceToY(c.low);
    ctx.strokeStyle = color;
    ctx.lineWidth = 1.2;
    ctx.beginPath();
    ctx.moveTo(x, highY);
    ctx.lineTo(x, lowY);
    ctx.stroke();

    // Candle Body
    const openY = priceToY(c.open);
    const closeY = priceToY(c.close);
    const bodyTop = Math.min(openY, closeY);
    const bodyHeight = Math.max(1.5, Math.abs(closeY - openY));

    ctx.fillStyle = color;
    ctx.fillRect(x - candleW / 2, bodyTop, candleW, bodyHeight);
  });

  // Current Price Dashed Reference Line
  const lastC = symCandles[symCandles.length - 1];
  if (lastC) {
    const lastY = priceToY(lastC.close);
    const lineCol = lastC.close >= lastC.open ? '#05d58c' : '#ff4766';

    ctx.strokeStyle = lineCol;
    ctx.setLineDash([4, 4]);
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(0, lastY);
    ctx.lineTo(chartW, lastY);
    ctx.stroke();
    ctx.setLineDash([]);

    // Price badge on right axis
    ctx.fillStyle = lineCol;
    ctx.fillRect(chartW + 2, lastY - 9, paddingRight - 6, 18);
    ctx.fillStyle = '#08090c';
    ctx.font = '600 10px "JetBrains Mono", monospace';
    ctx.textAlign = 'center';
    ctx.fillText(`$${lastC.close.toFixed(2)}`, chartW + (paddingRight - 4) / 2 + 2, lastY + 3.5);
  }

  // Crosshair & Tooltip Hover Interaction
  if (state.hoverPriceChart) {
    const { x: hX, y: hY } = state.hoverPriceChart;
    if (hX >= 0 && hX <= chartW && hY >= 0 && hY <= chartH) {
      // Crosshair lines
      ctx.strokeStyle = 'rgba(255, 255, 255, 0.22)';
      ctx.setLineDash([2, 2]);
      ctx.lineWidth = 1;

      // Vertical line
      ctx.beginPath();
      ctx.moveTo(hX, 0);
      ctx.lineTo(hX, chartH);
      ctx.stroke();

      // Horizontal line
      ctx.beginPath();
      ctx.moveTo(0, hY);
      ctx.lineTo(chartW, hY);
      ctx.stroke();
      ctx.setLineDash([]);

      // Tooltip price pill on axis
      const hoverPrice = yToPrice(hY);
      ctx.fillStyle = '#1e2430';
      ctx.strokeStyle = '#38bdf8';
      ctx.lineWidth = 1;
      ctx.fillRect(chartW + 2, hY - 9, paddingRight - 6, 18);
      ctx.strokeRect(chartW + 2, hY - 9, paddingRight - 6, 18);

      ctx.fillStyle = '#f1f5f9';
      ctx.font = '500 10px "JetBrains Mono", monospace';
      ctx.textAlign = 'center';
      ctx.fillText(`$${hoverPrice.toFixed(2)}`, chartW + (paddingRight - 4) / 2 + 2, hY + 3.5);

      // Find hovered candle
      const relativeX = hX - startX;
      if (relativeX >= 0) {
        const cIdx = Math.floor(relativeX / slotW);
        if (cIdx >= 0 && cIdx < symCandles.length) {
          updateOHLCHeader(symCandles[cIdx]);
        }
      }
    }
  }
}

// Chart Mouse Event Listeners
if (priceCanvas) {
  priceCanvas.addEventListener('mousemove', e => {
    const rect = priceCanvas.getBoundingClientRect();
    state.hoverPriceChart = {
      x: e.clientX - rect.left,
      y: e.clientY - rect.top,
    };
    renderPriceChart();
  });

  priceCanvas.addEventListener('mouseleave', () => {
    state.hoverPriceChart = null;
    updateOHLCHeader();
    renderPriceChart();
  });
}

// --------------------------------------------------------------------------
// 3. WebSocket Real-Time Client & Event Dispatcher
// --------------------------------------------------------------------------
function connectWebSocket() {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const wsUrl = `${protocol}//${window.location.host}/ws`;

  wsStatus.className = 'telemetry-pill connecting';
  wsStatus.querySelector('.telemetry-val').textContent = 'CONNECTING';

  state.ws = new WebSocket(wsUrl);

  state.ws.onopen = () => {
    wsStatus.className = 'telemetry-pill online';
    wsStatus.querySelector('.telemetry-val').textContent = 'LIVE FEED';
    fetchOrderBook();
  };

  state.ws.onmessage = evt => {
    try {
      const msg = JSON.parse(evt.data);
      handleServerEvent(msg);
    } catch (e) {
      console.error('WS parse error:', e);
    }
  };

  state.ws.onclose = () => {
    wsStatus.className = 'telemetry-pill offline';
    wsStatus.querySelector('.telemetry-val').textContent = 'RECONNECTING';
    setTimeout(connectWebSocket, 1500);
  };

  state.ws.onerror = () => {
    state.ws.close();
  };
}

function handleServerEvent(msg) {
  if (msg.type === 'trades') {
    const trades = msg.data || [];
    trades.forEach(t => {
      recordTrade(msg.symbol, t);
      updateCandleOnTrade(msg.symbol, t.price / 100, t.amount, t.timestamp || Date.now());

      // Visually pulse matched price level in L2 Order Book
      if (msg.symbol === state.activeSymbol) {
        const priceFormatted = (t.price / 100).toFixed(2);
        const row = document.querySelector(`.book-row[data-price="${priceFormatted}"]`);
        if (row) {
          row.classList.add('book-row-matched');
          setTimeout(() => row.classList.remove('book-row-matched'), 650);
        }
      }
    });
    updateOpenOrdersAfterMatch(trades);
    if (msg.symbol === state.activeSymbol) {
      renderTradesStream();
      renderTickerForActiveSymbol();
    }
  } else if (msg.type === 'book_update') {
    if (msg.symbol === state.activeSymbol) {
      scheduleBookRefresh();
    }
  } else if (msg.type === 'order_cancelled') {
    const idx = state.myOrders.findIndex(o => o.id === msg.order_id);
    if (idx !== -1) {
      state.myOrders.splice(idx, 1);
      renderOpenOrders();
    }
  } else if (msg.type === 'market_reset') {
    resetLocalMarketState();
  }
}

let bookRefreshTimer = null;
function scheduleBookRefresh() {
  if (bookRefreshTimer) return;
  bookRefreshTimer = setTimeout(() => {
    bookRefreshTimer = null;
    fetchOrderBook();
  }, 40);
}

// --------------------------------------------------------------------------
// 4. L2 Order Book Depth Fetch & Rendering
// --------------------------------------------------------------------------
async function fetchOrderBook() {
  try {
    const res = await fetch(`/orderbook?symbol=${state.activeSymbol}`);
    if (!res.ok) return;
    const data = await res.json();
    state.lastBookBySymbol[state.activeSymbol] = data;
    renderOrderBook(data);
  } catch (e) {
    console.error('Failed to fetch order book:', e);
  }
}

function renderOrderBook(data) {
  const bids = data.bids || [];
  const asks = data.asks || [];

  // Limit to top 14 levels per side for maximum readability
  const topAsks = asks.slice(0, 14);
  const topBids = bids.slice(0, 14);

  let cumAskVol = 0;
  const asksWithCum = topAsks.map(a => {
    cumAskVol += a.volume;
    return { ...a, cum: cumAskVol };
  });

  let cumBidVol = 0;
  const bidsWithCum = topBids.map(b => {
    cumBidVol += b.volume;
    return { ...b, cum: cumBidVol };
  });

  const maxCum = Math.max(cumAskVol, cumBidVol, 1);

  // Render Asks (Reversed so lowest ask is closest to spread in center)
  if (asksWithCum.length === 0) {
    asksContainer.innerHTML = '<div class="empty-state">No resting asks</div>';
  } else {
    const reversedAsks = [...asksWithCum].reverse();
    asksContainer.innerHTML = reversedAsks.map(a => {
      const pct = Math.min(100, (a.cum / maxCum) * 100);
      const priceFormatted = (a.price / 100).toFixed(2);
      return `
        <div class="book-row ask-row" data-price="${priceFormatted}">
          <div class="book-depth-bar ask-bar" style="width: ${pct}%"></div>
          <span class="col-price price-ask">$${priceFormatted}</span>
          <span class="col-size">${a.volume.toLocaleString()}</span>
          <span class="col-total">${a.cum.toLocaleString()}</span>
        </div>
      `;
    }).join('');
  }

  // Update Inside Spread
  const bestBid = topBids.length > 0 ? topBids[0].price : null;
  const bestAsk = topAsks.length > 0 ? topAsks[0].price : null;

  if (bestBid && bestAsk && bestAsk > bestBid) {
    const diff = (bestAsk - bestBid) / 100;
    const mid = (bestAsk + bestBid) / 2;
    const bps = ((bestAsk - bestBid) / mid) * 10000;
    spreadValue.textContent = `$${diff.toFixed(2)}`;
    spreadBps.textContent = `(${bps.toFixed(1)} bps)`;
  } else {
    spreadValue.textContent = '—';
    spreadBps.textContent = '';
  }

  // Render Bids (Highest bid at top, closest to spread)
  if (bidsWithCum.length === 0) {
    bidsContainer.innerHTML = '<div class="empty-state">No resting bids</div>';
  } else {
    bidsContainer.innerHTML = bidsWithCum.map(b => {
      const pct = Math.min(100, (b.cum / maxCum) * 100);
      const priceFormatted = (b.price / 100).toFixed(2);
      return `
        <div class="book-row bid-row" data-price="${priceFormatted}">
          <div class="book-depth-bar bid-bar" style="width: ${pct}%"></div>
          <span class="col-price price-bid">$${priceFormatted}</span>
          <span class="col-size">${b.volume.toLocaleString()}</span>
          <span class="col-total">${b.cum.toLocaleString()}</span>
        </div>
      `;
    }).join('');
  }

  // Attach click listener to book rows to auto-populate limit price
  document.querySelectorAll('.book-row').forEach(row => {
    row.addEventListener('click', () => {
      const p = row.getAttribute('data-price');
      if (p) {
        inputPrice.value = p;
        if (state.type === 1) {
          // Switch to limit mode if user clicked a specific price
          setType(0);
        }
        updateTotal();
      }
    });
  });
}

// --------------------------------------------------------------------------
// 5. Execution Tape & Session Telemetry
// --------------------------------------------------------------------------
function formatUSD(val) {
  if (val === null || val === undefined) return '—';
  return `$${val.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
}

function formatPct(pct) {
  if (pct === null || pct === undefined) return '—';
  const sign = pct >= 0 ? '+' : '';
  return `${sign}${pct.toFixed(2)}%`;
}

function getTradeKey(symbol, trade) {
  return `${symbol}:${trade.maker_order_id}:${trade.taker_order_id}:${trade.price}:${trade.amount}:${trade.timestamp}`;
}

function recordTrade(symbol, trade) {
  const list = state.recentTrades[symbol];
  const key = getTradeKey(symbol, trade);

  if (list.some(t => getTradeKey(symbol, t) === key)) {
    return;
  }

  list.unshift(trade);
  if (list.length > 50) list.pop();

  const stats = state.statsBySymbol[symbol];
  const p = trade.price / 100;
  if (stats.openPrice === null) stats.openPrice = p;
  stats.lastPrice = p;
  stats.high = stats.high === null ? p : Math.max(stats.high, p);
  stats.low = stats.low === null ? p : Math.min(stats.low, p);
  stats.volume += trade.amount;
  stats.tradesCount += 1;

  // Update header tab price
  const tabPriceEl = document.getElementById(`tabPrice-${symbol}`);
  if (tabPriceEl) {
    tabPriceEl.textContent = `$${p.toFixed(2)}`;
  }
}

function renderTradesStream() {
  const trades = state.recentTrades[state.activeSymbol] || [];
  if (trades.length === 0) {
    tradesStream.innerHTML = '<div class="empty-state">Waiting for executions...</div>';
    return;
  }

  if (state.matchMode === 'audit') {
    tradesStream.innerHTML = trades.map((t, idx) => {
      const isBuy = t.side === 0;
      const priceFormatted = (t.price / 100).toFixed(2);
      const notionalFormatted = ((t.price * t.amount) / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
      const timeStr = new Date(t.timestamp).toTimeString().split(' ')[0];
      const isNew = idx === 0 ? 'new-match' : '';

      return `
        <div class="match-card ${isNew}" onclick="inspectTrade(${idx})">
          <div class="match-card-header">
            <div class="match-linkage">
              <span class="id-pill ${isBuy ? 'taker-pill-buy' : 'taker-pill-sell'}">
                TAKER #${t.taker_order_id} ${isBuy ? 'BUY' : 'SELL'}
              </span>
              <span class="match-arrow">➔</span>
              <span class="id-pill maker-pill">
                MAKER #${t.maker_order_id}
              </span>
            </div>
            <span class="fifo-badge">FIFO 1ST</span>
          </div>

          <div class="match-card-body">
            <span class="match-price ${isBuy ? 'text-buy' : 'text-sell'}">$${priceFormatted}</span>
            <div class="match-details">
              <span>${t.amount.toLocaleString()} UNITS</span>
              <span>$${notionalFormatted}</span>
              <span class="match-time">${timeStr}</span>
            </div>
          </div>

          <div class="match-card-footer">
            <span>⚡ Matched from Head of PriceLevel queue</span>
            <span class="match-action-hint">Audit Invariants ↗</span>
          </div>
        </div>
      `;
    }).join('');
  } else {
    tradesStream.innerHTML = trades.map((t, idx) => {
      const isBuy = t.side === 0;
      const priceFormatted = (t.price / 100).toFixed(2);
      const timeStr = new Date(t.timestamp).toTimeString().split(' ')[0];

      return `
        <div class="trade-row-compact" onclick="inspectTrade(${idx})">
          <span class="col-link-text">#${t.taker_order_id} ➔ #${t.maker_order_id}</span>
          <span class="${isBuy ? 'text-buy' : 'text-sell'}">$${priceFormatted}</span>
          <span>${t.amount.toLocaleString()}</span>
          <span class="col-time">${timeStr}</span>
        </div>
      `;
    }).join('');
  }
}

window.inspectTrade = function(idx) {
  const trades = state.recentTrades[state.activeSymbol] || [];
  const t = trades[idx];
  if (!t) return;

  const isBuy = t.side === 0;
  const priceFormatted = (t.price / 100).toFixed(2);
  const notionalFormatted = ((t.price * t.amount) / 100).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });

  modalTradeTitle.textContent = `Taker #${t.taker_order_id} (${isBuy ? 'BUY' : 'SELL'}) ➔ Maker #${t.maker_order_id}`;
  modalTakerId.textContent = `Order #${t.taker_order_id}`;
  modalTakerSideBadge.textContent = isBuy ? 'BUY (TAKER)' : 'SELL (TAKER)';
  modalTakerSideBadge.className = `flow-badge ${isBuy ? 'buy-badge' : 'sell-badge'}`;
  modalConnectorPrice.textContent = `$${priceFormatted}`;
  modalMakerId.textContent = `Order #${t.maker_order_id}`;
  modalSettlementVal.textContent = `$${notionalFormatted} USD`;
  modalSharesVal.textContent = `${t.amount.toLocaleString()} shares @ $${priceFormatted}`;
  modalTicksVal.textContent = `${t.price.toLocaleString()} ticks × ${t.amount} units = ${(t.price * t.amount).toLocaleString()} cents`;

  matchModalBackdrop.style.display = 'flex';
};

if (modalCloseBtn) {
  modalCloseBtn.addEventListener('click', () => {
    matchModalBackdrop.style.display = 'none';
  });
}
if (matchModalBackdrop) {
  matchModalBackdrop.addEventListener('click', e => {
    if (e.target === matchModalBackdrop) {
      matchModalBackdrop.style.display = 'none';
    }
  });
}

if (btnModeAudit && btnModeTape) {
  btnModeAudit.addEventListener('click', () => {
    state.matchMode = 'audit';
    btnModeAudit.classList.add('active');
    btnModeTape.classList.remove('active');
    tradesStream.classList.add('audit-mode');
    tradesTableHeader.style.display = 'none';
    renderTradesStream();
  });

  btnModeTape.addEventListener('click', () => {
    state.matchMode = 'tape';
    btnModeTape.classList.add('active');
    btnModeAudit.classList.remove('active');
    tradesStream.classList.remove('audit-mode');
    tradesTableHeader.style.display = 'grid';
    renderTradesStream();
  });
}

function renderTickerForActiveSymbol() {
  const stats = state.statsBySymbol[state.activeSymbol];
  tickerLastPrice.textContent = formatUSD(stats.lastPrice);
  lastTradedPrice.textContent = stats.lastPrice !== null ? `$${stats.lastPrice.toFixed(2)}` : '—';

  if (stats.openPrice && stats.lastPrice !== null) {
    const chg = ((stats.lastPrice - stats.openPrice) / stats.openPrice) * 100;
    tickerChange.textContent = formatPct(chg);
    tickerChange.className = `ticker-val ${chg >= 0 ? 'text-buy' : 'text-sell'}`;
    lastTradedPrice.style.color = chg >= 0 ? 'var(--color-buy)' : 'var(--color-sell)';
  } else {
    tickerChange.textContent = '0.00%';
    tickerChange.className = 'ticker-val';
    lastTradedPrice.style.color = 'var(--text-primary)';
  }

  tickerHigh.textContent = formatUSD(stats.high);
  tickerLow.textContent = formatUSD(stats.low);
  tickerVolume.textContent = stats.volume.toLocaleString();
  tickerTradesCount.textContent = stats.tradesCount.toLocaleString();
}

// --------------------------------------------------------------------------
// 6. Order Placement & Resting Order Management
// --------------------------------------------------------------------------
async function submitOrder() {
  const priceVal = parseFloat(inputPrice.value);
  const amountVal = parseInt(inputAmount.value, 10);

  if (isNaN(amountVal) || amountVal <= 0) {
    alert('Please enter a valid order quantity (positive integer).');
    return;
  }

  if (state.type === 0 && (isNaN(priceVal) || priceVal <= 0)) {
    alert('Please enter a valid limit price.');
    return;
  }

  const payload = {
    symbol: state.activeSymbol,
    side: state.side,
    type: state.type,
    price: state.type === 0 ? Math.round(priceVal * 100) : 0,
    amount: amountVal,
  };

  btnSubmitOrder.disabled = true;
  btnSubmitOrder.textContent = 'ROUTING TO ENGINE...';

  try {
    const res = await fetch('/order', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });

    if (!res.ok) {
      const errText = await res.text();
      alert(`Order rejected: ${errText}`);
      return;
    }

    const data = await res.json();

    // If order rested or partially rested, track in active orders table
    if (data.status === 'RESTING' || data.status === 'PARTIALLY_FILLED_RESTING') {
      const earlyFilled = takeEarlyFills(data.order.id);
      const remaining = data.remaining_amount - earlyFilled;
      if (remaining > 0) {
        state.myOrders.unshift({
          id: data.order.id,
          symbol: data.order.symbol,
          side: data.order.side,
          price: data.order.price,
          amount: remaining,
        });
        renderOpenOrders();
      }
    }

    // Refresh book and trade tape immediately
    scheduleBookRefresh();
  } catch (e) {
    console.error('Order submission failed:', e);
    alert('Failed to connect to matching engine.');
  } finally {
    btnSubmitOrder.disabled = false;
    updateSubmitButton();
  }
}

function renderOpenOrders() {
  const activeOrders = state.myOrders.filter(o => o.symbol === state.activeSymbol);
  openOrdersCount.textContent = activeOrders.length;

  if (activeOrders.length === 0) {
    ordersTableBody.innerHTML = '<tr class="empty-row"><td colspan="5">No active resting orders</td></tr>';
    return;
  }

  ordersTableBody.innerHTML = activeOrders.map(o => {
    const isBuy = o.side === 0;
    const priceFormatted = (o.price / 100).toFixed(2);
    return `
      <tr>
        <td class="order-id">#${o.id}</td>
        <td class="${isBuy ? 'text-buy' : 'text-sell'}">${isBuy ? 'BUY' : 'SELL'}</td>
        <td class="mono">$${priceFormatted}</td>
        <td class="mono">${o.amount.toLocaleString()}</td>
        <td>
          <button class="btn-cancel" onclick="cancelOrder(${o.id})">CANCEL</button>
        </td>
      </tr>
    `;
  }).join('');
}

window.cancelOrder = async function(id) {
  try {
    const res = await fetch(`/order?symbol=${state.activeSymbol}&id=${id}`, {
      method: 'DELETE',
    });
    if (!res.ok) {
      const errText = await res.text();
      alert(`Cancel failed: ${errText}`);
      return;
    }
    state.myOrders = state.myOrders.filter(o => o.id !== id);
    renderOpenOrders();
    scheduleBookRefresh();
  } catch (e) {
    console.error('Cancel order error:', e);
  }
};

function updateOpenOrdersAfterMatch(trades) {
  trades.forEach(t => {
    const idx = state.myOrders.findIndex(o => o.id === t.maker_order_id);
    if (idx === -1) {
      const cur = state.earlyMakerFills.get(t.maker_order_id) || 0;
      state.earlyMakerFills.set(t.maker_order_id, cur + t.amount);
      return;
    }
    const o = state.myOrders[idx];
    const remaining = o.amount - t.amount;
    if (remaining > 0) {
      state.myOrders[idx] = { ...o, amount: remaining };
    } else {
      state.myOrders.splice(idx, 1);
    }
  });

  while (state.earlyMakerFills.size > 100) {
    state.earlyMakerFills.delete(state.earlyMakerFills.keys().next().value);
  }
  renderOpenOrders();
}

function takeEarlyFills(id) {
  const filled = state.earlyMakerFills.get(id) || 0;
  state.earlyMakerFills.delete(id);
  return filled;
}

// --------------------------------------------------------------------------
// 7. Form Controls & Event Handlers
// --------------------------------------------------------------------------
function updateTotal() {
  const price = parseFloat(inputPrice.value) || 0;
  const amount = parseInt(inputAmount.value, 10) || 0;
  if (state.type === 1) {
    orderTotal.textContent = 'At Best Market Price';
  } else {
    orderTotal.textContent = `$${(price * amount).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
  }
}

function updateSubmitButton() {
  const sideText = state.side === 0 ? 'BUY' : 'SELL';
  const typeText = state.type === 0 ? 'LIMIT' : 'MARKET';
  btnSubmitOrder.textContent = `SUBMIT ${typeText} ${sideText}`;
  btnSubmitOrder.className = `btn-submit ${state.side === 0 ? 'btn-submit-buy' : 'btn-submit-sell'}`;
}

function setSide(side) {
  state.side = side;
  btnSideBuy.classList.toggle('active', side === 0);
  btnSideSell.classList.toggle('active', side === 1);
  updateSubmitButton();
}

function setType(type) {
  state.type = type;
  btnTypeLimit.classList.toggle('active', type === 0);
  btnTypeMarket.classList.toggle('active', type === 1);
  priceGroup.style.display = type === 0 ? 'block' : 'none';
  updateTotal();
  updateSubmitButton();
}

btnSideBuy.addEventListener('click', () => setSide(0));
btnSideSell.addEventListener('click', () => setSide(1));
btnTypeLimit.addEventListener('click', () => setType(0));
btnTypeMarket.addEventListener('click', () => setType(1));

inputPrice.addEventListener('input', updateTotal);
inputAmount.addEventListener('input', updateTotal);

document.querySelectorAll('.quick-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    inputAmount.value = btn.getAttribute('data-qty');
    updateTotal();
  });
});

btnSubmitOrder.addEventListener('click', submitOrder);

// Symbol Navigation Tabs
symbolTabs.addEventListener('click', e => {
  const btn = e.target.closest('.tab-btn');
  if (!btn) return;
  const sym = btn.getAttribute('data-symbol');
  if (sym && sym !== state.activeSymbol) {
    switchSymbol(sym);
  }
});

function switchSymbol(symbol) {
  state.activeSymbol = symbol;
  activeSymbolTag.textContent = symbol;

  document.querySelectorAll('.tab-btn').forEach(btn => {
    btn.classList.toggle('active', btn.getAttribute('data-symbol') === symbol);
  });

  if (REFERENCE_PRICES[symbol]) {
    inputPrice.value = REFERENCE_PRICES[symbol];
  }

  updateTotal();
  renderTickerForActiveSymbol();
  renderTradesStream();
  renderOpenOrders();
  updateOHLCHeader();
  renderPriceChart();
  fetchOrderBook();
}

// --------------------------------------------------------------------------
// 8. Simulator Toggle & Initial State Sync
// --------------------------------------------------------------------------
function renderBotButton() {
  btnToggleBot.classList.toggle('active', state.botRunning);
  botLabel.textContent = state.botRunning ? 'SIMULATOR: LIVE' : 'SIMULATOR: PAUSED';
}

async function syncBotStatus() {
  try {
    const res = await fetch('/simulator/status');
    if (!res.ok) return;
    const data = await res.json();
    state.botRunning = !!data.running;
    renderBotButton();
  } catch (e) {
    console.error('Failed to sync bot status:', e);
  }
}

if (btnResetMarket) {
  btnResetMarket.addEventListener('click', async () => {
    btnResetMarket.disabled = true;
    btnResetMarket.textContent = 'RESETTING...';
    try {
      const res = await fetch('/simulator/reset', { method: 'POST' });
      if (res.ok) {
        resetLocalMarketState();
      }
    } catch (e) {
      console.error('Failed to reset market:', e);
    } finally {
      btnResetMarket.disabled = false;
      btnResetMarket.innerHTML = '<span class="reset-icon">↺</span> RESET MARKET';
    }
  });
}

function resetLocalMarketState() {
  for (const sym of SYMBOLS) {
    state.statsBySymbol[sym] = emptyStats();
    state.recentTrades[sym] = [];
    candles[sym] = [];
    const tabPriceEl = document.getElementById(`tabPrice-${sym}`);
    if (tabPriceEl) tabPriceEl.textContent = '—';
  }
  state.myOrders = [];
  state.earlyMakerFills.clear();

  renderTickerForActiveSymbol();
  renderTradesStream();
  renderOpenOrders();
  updateOHLCHeader();
  renderPriceChart();
  fetchOrderBook();
}

btnToggleBot.addEventListener('click', async () => {
  try {
    btnToggleBot.disabled = true;
    const res = await fetch('/simulator/toggle', { method: 'POST' });
    if (res.ok) {
      const data = await res.json();
      state.botRunning = !!data.running;
      renderBotButton();
    }
  } catch (e) {
    console.error('Failed to toggle bot:', e);
  } finally {
    btnToggleBot.disabled = false;
  }
});

async function hydrateTradeHistory(symbol) {
  try {
    const res = await fetch(`/trades?symbol=${symbol}&limit=40`);
    if (!res.ok) return;
    const data = await res.json();
    const trades = (data.trades || []).slice().reverse();
    trades.forEach(t => {
      recordTrade(symbol, t);
      updateCandleOnTrade(symbol, t.price / 100, t.amount, t.timestamp || Date.now());
    });
  } catch (e) {
    console.error('Failed to hydrate trade history:', e);
  }
}

// Window resize re-render
window.addEventListener('resize', () => {
  renderPriceChart();
});

// --------------------------------------------------------------------------
// 9. Application Bootstrap
// --------------------------------------------------------------------------
async function init() {
  updateTotal();
  updateSubmitButton();

  for (const sym of SYMBOLS) {
    await hydrateTradeHistory(sym);
  }

  renderTickerForActiveSymbol();
  renderTradesStream();
  renderOpenOrders();
  updateOHLCHeader();
  renderPriceChart();

  connectWebSocket();
  syncBotStatus();
}

document.addEventListener('DOMContentLoaded', init);
