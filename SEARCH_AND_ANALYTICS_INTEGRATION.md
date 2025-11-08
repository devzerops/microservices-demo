# Search & Analytics Services Integration Guide

Integration guide for the Search Service (autocomplete & trending) and Analytics Dashboard Service.

## Services Overview

| Service | Port | Technology | Purpose |
|---------|------|------------|---------|
| Search | 8097 | Go + Trie + Redis | Autocomplete, trending queries, fuzzy matching |
| Analytics | 8099 | Go + In-Memory | Real-time metrics, event tracking, dashboard |

## Quick Start

```bash
# Start both services
make start-experimental

# Or individually
docker-compose -f docker-compose-experimental.yml up -d search analytics

# Check health
curl http://localhost:8097/health
curl http://localhost:8099/health
```

## Integration Examples

### 1. Track Search Queries in Analytics

```javascript
// When user searches
async function performSearch(query, userId) {
  // Get autocomplete suggestions
  const response = await fetch(
    `http://localhost:8097/autocomplete?q=${encodeURIComponent(query)}&user_id=${userId}`
  );
  const { suggestions } = await response.json();

  // Track search event in analytics
  await fetch('http://localhost:8099/events', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      type: 'search',
      service: 'search',
      user_id: userId,
      data: {
        query,
        results_count: suggestions.length,
        took_ms: response.headers.get('X-Response-Time')
      }
    })
  });

  return suggestions;
}
```

### 2. Monitor Search Service Performance

```javascript
// Send heartbeat from Search service to Analytics
setInterval(async () => {
  const stats = await fetch('http://localhost:8097/stats').then(r => r.json());

  await fetch('http://localhost:8099/heartbeat', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      service: 'search',
      status: 'healthy',
      metrics: {
        request_count: stats.total_searches,
        indexed_terms: stats.indexed_terms,
        trending_count: stats.trending_queries,
        avg_latency: 3.5 // Search is typically < 5ms
      }
    })
  });
}, 60000); // Every minute
```

### 3. Real-time Search Trending Dashboard

```javascript
class SearchTrendingDashboard {
  constructor() {
    // Connect to Search Service trending
    this.searchWS = new WebSocket('ws://localhost:8097/ws/trending');

    // Connect to Analytics dashboard
    this.analyticsWS = new WebSocket('ws://localhost:8099/ws/dashboard');

    this.setupHandlers();
  }

  setupHandlers() {
    // Search trending updates
    this.searchWS.onmessage = (event) => {
      const message = JSON.parse(event.data);

      if (message.type === 'update') {
        this.renderTrending(message.data);
      }
    };

    // Analytics dashboard updates
    this.analyticsWS.onmessage = (event) => {
      const message = JSON.parse(event.data);

      if (message.type === 'dashboard_update') {
        this.renderMetrics(message.data);
      }
    };
  }

  renderTrending(trending) {
    document.getElementById('trending').innerHTML = `
      <h3>🔥 Trending Searches</h3>
      <ul>
        ${trending.map(t => `
          <li>
            #${t.rank} ${t.query}
            (${t.count} searches, ${t.velocity.toFixed(1)}/min)
          </li>
        `).join('')}
      </ul>
    `;
  }

  renderMetrics(data) {
    const searchMetrics = data.services.search || {};

    document.getElementById('metrics').innerHTML = `
      <h3>📊 Search Metrics</h3>
      <div>Requests: ${searchMetrics.request_count || 0}</div>
      <div>Avg Latency: ${(searchMetrics.avg_latency_ms || 0).toFixed(1)}ms</div>
      <div>Status: ${searchMetrics.status || 'unknown'}</div>
    `;
  }
}

// Initialize
new SearchTrendingDashboard();
```

### 4. Search Analytics Dashboard

```javascript
// Complete dashboard combining both services
async function createSearchAnalyticsDashboard() {
  // Get trending from Search
  const trending = await fetch('http://localhost:8097/trending?period=1h')
    .then(r => r.json());

  // Get search events from Analytics
  const searchEvents = await fetch('http://localhost:8099/events?type=search')
    .then(r => r.json());

  // Get overall stats from Analytics
  const dashboard = await fetch('http://localhost:8099/dashboard')
    .then(r => r.json());

  return {
    trending: trending.trending,
    recent_searches: searchEvents.events,
    total_searches: dashboard.services.search?.request_count || 0,
    search_latency: dashboard.services.search?.avg_latency_ms || 0,
    active_users: dashboard.overview.active_users
  };
}

// Render dashboard
async function renderDashboard() {
  const data = await createSearchAnalyticsDashboard();

  document.getElementById('dashboard').innerHTML = `
    <div class="search-analytics">
      <div class="overview">
        <h2>Search Overview</h2>
        <div class="stat">
          <span class="value">${data.total_searches.toLocaleString()}</span>
          <span class="label">Total Searches</span>
        </div>
        <div class="stat">
          <span class="value">${data.search_latency.toFixed(1)}ms</span>
          <span class="label">Avg Response Time</span>
        </div>
        <div class="stat">
          <span class="value">${data.active_users}</span>
          <span class="label">Active Users</span>
        </div>
      </div>

      <div class="trending">
        <h3>🔥 Trending Now</h3>
        ${data.trending.map(t => `
          <div class="trend-item">
            <span class="rank">#${t.rank}</span>
            <span class="query">${t.query}</span>
            <span class="stats">${t.count} searches</span>
            <span class="velocity">${t.velocity.toFixed(1)}/min</span>
          </div>
        `).join('')}
      </div>

      <div class="recent">
        <h3>Recent Searches</h3>
        ${data.recent_searches.slice(0, 10).map(e => `
          <div class="search-event">
            <span class="query">${e.data.query}</span>
            <span class="time">${new Date(e.timestamp).toLocaleTimeString()}</span>
          </div>
        `).join('')}
      </div>
    </div>
  `;
}

// Update every 5 seconds
setInterval(renderDashboard, 5000);
renderDashboard();
```

### 5. Search Performance Metrics

```go
// Go middleware to track search performance
func trackSearchMetrics(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Call the handler
        next(w, r)

        // Calculate latency
        latency := time.Since(start).Milliseconds()

        // Send to Analytics
        go func() {
            metric := map[string]interface{}{
                "name":  "latency",
                "value": float64(latency),
                "unit":  "ms",
                "tags": map[string]string{
                    "service":  "search",
                    "endpoint": r.URL.Path,
                },
            }

            body, _ := json.Marshal(metric)
            http.Post(
                "http://analytics:8099/metrics",
                "application/json",
                bytes.NewBuffer(body),
            )
        }()
    }
}

// Use in Search service
router.HandleFunc("/autocomplete", trackSearchMetrics(autocompleteHandler))
router.HandleFunc("/trending", trackSearchMetrics(trendingHandler))
```

### 6. User Search History Analytics

```javascript
// Get user's search history and analytics
async function getUserSearchAnalytics(userId) {
  // Get search history from Search service
  const history = await fetch(`http://localhost:8097/history/${userId}`)
    .then(r => r.json());

  // Get user events from Analytics
  const events = await fetch(`http://localhost:8099/events?service=search`)
    .then(r => r.json())
    .then(data => data.events.filter(e => e.user_id === userId));

  // Analyze patterns
  const queryCounts = {};
  history.history.forEach(item => {
    queryCounts[item.query] = (queryCounts[item.query] || 0) + 1;
  });

  const topQueries = Object.entries(queryCounts)
    .sort((a, b) => b[1] - a[1])
    .slice(0, 5);

  return {
    total_searches: history.total,
    recent_searches: history.history,
    top_queries: topQueries,
    search_events: events
  };
}
```

### 7. Alerts Based on Search Trends

```javascript
// Monitor for sudden spikes in specific searches
async function monitorSearchSpikes() {
  const trending = await fetch('http://localhost:8097/trending?period=5m')
    .then(r => r.json());

  for (const query of trending.trending) {
    // If velocity is very high, it's trending fast
    if (query.velocity > 10) { // More than 10 searches per minute
      // Track alert event
      await fetch('http://localhost:8099/events', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          type: 'alert',
          service: 'search',
          data: {
            alert_type: 'trending_spike',
            query: query.query,
            velocity: query.velocity,
            count: query.count,
            message: `"${query.query}" is trending rapidly!`
          }
        })
      });

      console.warn(`Alert: "${query.query}" trending at ${query.velocity}/min`);
    }
  }
}

// Check every minute
setInterval(monitorSearchSpikes, 60000);
```

## Combined Features

### 1. Search Quality Monitoring

Track search quality metrics:

- Click-through rate (searches → product views)
- Zero-result searches
- Fuzzy match ratio
- Autocomplete effectiveness

```javascript
// Track search effectiveness
async function trackSearchQuality(query, resultClicked) {
  await fetch('http://localhost:8099/events', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      type: 'search_quality',
      service: 'search',
      data: {
        query,
        result_clicked: resultClicked,
        conversion: resultClicked ? 1 : 0
      }
    })
  });
}
```

### 2. A/B Testing Search Algorithms

```javascript
// Test different search configurations
async function abTestSearch(query, userId) {
  // Randomly assign to control or variant
  const variant = Math.random() < 0.5 ? 'control' : 'variant';

  // Get results (could use different endpoints/params)
  const results = await fetch(
    `http://localhost:8097/autocomplete?q=${query}&variant=${variant}`
  ).then(r => r.json());

  // Track experiment
  await fetch('http://localhost:8099/events', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      type: 'ab_test',
      service: 'search',
      user_id: userId,
      data: {
        experiment: 'autocomplete_v2',
        variant,
        query,
        results_count: results.suggestions.length
      }
    })
  });

  return results;
}
```

### 3. Real-time Search Status Page

```html
<!DOCTYPE html>
<html>
<head>
  <title>Search Service Status</title>
  <style>
    body { font-family: Arial, sans-serif; margin: 20px; }
    .status { display: flex; gap: 20px; margin-bottom: 30px; }
    .card { flex: 1; padding: 20px; border-radius: 8px; background: #f5f5f5; }
    .healthy { border-left: 4px solid #4caf50; }
    .unhealthy { border-left: 4px solid #f44336; }
    .trending { margin-top: 20px; }
    .trend-item { padding: 10px; margin: 5px 0; background: white; border-radius: 4px; }
  </style>
</head>
<body>
  <h1>🔍 Search Service Status</h1>
  <div id="status" class="status"></div>
  <div id="trending" class="trending"></div>
  <div id="metrics" style="margin-top: 20px;"></div>

  <script>
    // Combined status page
    async function updateStatus() {
      // Get Search service stats
      const searchStats = await fetch('http://localhost:8097/stats').then(r => r.json());

      // Get Analytics dashboard
      const dashboard = await fetch('http://localhost:8099/dashboard').then(r => r.json());

      // Get trending
      const trending = await fetch('http://localhost:8097/trending?period=1h').then(r => r.json());

      // Render status cards
      document.getElementById('status').innerHTML = `
        <div class="card healthy">
          <h3>Search Service</h3>
          <p>Status: ✓ Healthy</p>
          <p>Indexed Terms: ${searchStats.indexed_terms}</p>
          <p>Total Searches: ${searchStats.total_searches}</p>
        </div>
        <div class="card ${dashboard.services.search?.status === 'healthy' ? 'healthy' : 'unhealthy'}">
          <h3>Metrics</h3>
          <p>Requests/sec: ${dashboard.overview.requests_per_second.toFixed(1)}</p>
          <p>Avg Latency: ${dashboard.overview.average_latency_ms.toFixed(1)}ms</p>
          <p>Error Rate: ${dashboard.overview.error_rate.toFixed(2)}%</p>
        </div>
      `;

      // Render trending
      document.getElementById('trending').innerHTML = `
        <h2>🔥 Trending Searches</h2>
        ${trending.trending.map(t => `
          <div class="trend-item">
            <strong>#${t.rank} ${t.query}</strong> -
            ${t.count} searches (${t.velocity.toFixed(1)}/min)
          </div>
        `).join('')}
      `;
    }

    // Update every 5 seconds
    setInterval(updateStatus, 5000);
    updateStatus();
  </script>
</body>
</html>
```

## Testing

### Test Search + Analytics Integration

```bash
# 1. Start services
docker-compose -f docker-compose-experimental.yml up -d search analytics

# 2. Perform some searches
for i in {1..10}; do
  curl "http://localhost:8097/autocomplete?q=sun&user_id=test-user-$i"
  sleep 1
done

# 3. Check trending
curl "http://localhost:8097/trending?period=1h"

# 4. Check analytics dashboard
curl "http://localhost:8099/dashboard" | jq .

# 5. Check search events
curl "http://localhost:8099/events?type=search" | jq .
```

## Performance Considerations

**Search Service**:
- Autocomplete: < 5ms
- Trending updates: Every 5 seconds
- Memory: ~100MB for 10,000 terms

**Analytics Service**:
- Event tracking: < 5ms
- Dashboard updates: Every 1 second
- Memory: ~150MB for 100,000 events

**Combined**:
- Total memory: ~250MB
- Network overhead: Minimal (<1KB per request)

## Monitoring Dashboard Setup

```javascript
// Complete monitoring setup
class CombinedMonitoring {
  constructor() {
    this.setupWebSockets();
    this.setupPolling();
  }

  setupWebSockets() {
    // Search trending
    this.searchWS = new WebSocket('ws://localhost:8097/ws/trending');
    this.searchWS.onmessage = (e) => this.handleSearchUpdate(JSON.parse(e.data));

    // Analytics dashboard
    this.analyticsWS = new WebSocket('ws://localhost:8099/ws/dashboard');
    this.analyticsWS.onmessage = (e) => this.handleAnalyticsUpdate(JSON.parse(e.data));
  }

  setupPolling() {
    // Poll for additional data every 30 seconds
    setInterval(() => this.fetchAdditionalData(), 30000);
  }

  handleSearchUpdate(data) {
    if (data.type === 'update') {
      this.updateTrending(data.data);
    }
  }

  handleAnalyticsUpdate(data) {
    if (data.type === 'dashboard_update') {
      this.updateMetrics(data.data);
    }
  }

  async fetchAdditionalData() {
    // Fetch supplementary data
    const [searchStats, analyticsStats] = await Promise.all([
      fetch('http://localhost:8097/stats').then(r => r.json()),
      fetch('http://localhost:8099/stats').then(r => r.json())
    ]);

    this.updateStats({ search: searchStats, analytics: analyticsStats });
  }
}
```

## License

Apache-2.0
