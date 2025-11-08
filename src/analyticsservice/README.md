# Analytics Dashboard Service

Real-time analytics and monitoring service for tracking metrics, events, and service health across the microservices ecosystem.

## Features

- 📊 **Real-time Dashboard**: Live metrics via WebSocket
- 📈 **Event Tracking**: Custom event collection and analysis
- 💓 **Health Monitoring**: Service heartbeat and status tracking
- 📉 **Metrics Collection**: Time-series metrics storage
- 🔄 **Aggregation**: Hourly and daily statistics
- ⚡ **High Performance**: In-memory processing with <10ms latency

## Architecture

### Core Components

**1. Metrics Collector**
- Records service metrics (requests, errors, latency)
- Tracks service health and uptime
- Calculates percentiles (p50, p95, p99)
- Stores time-series data

**2. Event Tracker**
- Collects custom events from all services
- Real-time event streaming
- Event categorization by type and service
- Automatic cleanup of old events

**3. Aggregator**
- Hourly and daily aggregations
- Peak hour detection
- Unique user tracking
- Historical data retention

**4. WebSocket Broadcaster**
- Real-time dashboard updates
- Multiple client support
- Automatic reconnection handling
- Snapshot + incremental updates

## API Endpoints

### Track Event

**POST** `/events`

Track a custom event.

```bash
curl -X POST http://localhost:8099/events \
  -H "Content-Type: application/json" \
  -d '{
    "type": "product_view",
    "service": "frontend",
    "user_id": "user-123",
    "data": {
      "product_id": "OLJCESPC7Z",
      "category": "accessories"
    }
  }'
```

**Event Types**:
- `request` - HTTP request
- `error` - Error occurred
- `purchase` - Order completed
- `product_view` - Product viewed
- `search` - Search performed
- `cart_add` - Item added to cart
- Custom types...

### Get Events

**GET** `/events?type=<type>&service=<service>`

Retrieve tracked events.

```bash
# Get all events
curl "http://localhost:8099/events"

# Get events by type
curl "http://localhost:8099/events?type=purchase"

# Get events by service
curl "http://localhost:8099/events?service=checkout"
```

Response:
```json
{
  "events": [
    {
      "type": "purchase",
      "service": "checkout",
      "user_id": "user-123",
      "timestamp": "2024-01-15T10:30:00Z",
      "data": {
        "order_id": "ORDER-123",
        "total": 99.99
      }
    }
  ],
  "count": 25,
  "limit": 100
}
```

### Record Metric

**POST** `/metrics`

Record a metric value.

```bash
curl -X POST http://localhost:8099/metrics \
  -H "Content-Type: application/json" \
  -d '{
    "name": "latency",
    "value": 45.2,
    "unit": "ms",
    "tags": {
      "service": "productcatalog",
      "endpoint": "/products"
    }
  }'
```

**Metric Names**:
- `requests` - Request count
- `errors` - Error count
- `latency` - Response time (ms)
- `memory` - Memory usage (MB)
- `cpu` - CPU usage (%)
- Custom metrics...

### Get Dashboard Data

**GET** `/dashboard`

Get complete dashboard snapshot.

```bash
curl http://localhost:8099/dashboard
```

Response:
```json
{
  "overview": {
    "total_requests": 123456,
    "active_users": 45,
    "average_latency_ms": 25.3,
    "error_rate": 0.5,
    "requests_per_second": 42.5
  },
  "services": {
    "productcatalog": {
      "name": "productcatalog",
      "status": "healthy",
      "uptime_hours": 72.5,
      "request_count": 15234,
      "error_count": 12,
      "avg_latency_ms": 18.2,
      "last_heartbeat": "2024-01-15T10:30:00Z"
    }
  },
  "realtime_stats": {
    "requests_last_1min": 125,
    "errors_last_1min": 2,
    "active_connections": 8,
    "top_endpoints": [
      {
        "endpoint": "/products",
        "count": 450,
        "avg_time_ms": 22.1
      }
    ],
    "recent_events": [...]
  },
  "updated_at": "2024-01-15T10:30:15Z"
}
```

### Service Heartbeat

**POST** `/heartbeat`

Report service health.

```bash
curl -X POST http://localhost:8099/heartbeat \
  -H "Content-Type: application/json" \
  -d '{
    "service": "productcatalog",
    "status": "healthy",
    "metrics": {
      "request_count": 1000,
      "error_count": 5,
      "avg_latency": 23.5,
      "uptime": 72.5
    }
  }'
```

### Get Service Metrics

**GET** `/services/{service}/metrics`

Get metrics for specific service.

```bash
curl http://localhost:8099/services/productcatalog/metrics
```

### WebSocket - Real-time Dashboard

**WS** `/ws/dashboard`

Stream real-time dashboard updates.

```javascript
const ws = new WebSocket('ws://localhost:8099/ws/dashboard');

ws.onopen = () => {
  console.log('Connected to analytics dashboard');
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);

  if (message.type === 'snapshot') {
    // Initial dashboard data
    console.log('Dashboard snapshot:', message.data);
  } else if (message.type === 'dashboard_update') {
    // Real-time update
    updateDashboard(message.data);
  } else if (message.type === 'event') {
    // New event occurred
    showNotification(message.data);
  }
};
```

### Health Check

**GET** `/health`

Check service health.

```bash
curl http://localhost:8099/health
```

### Statistics

**GET** `/stats`

Get service statistics.

```bash
curl http://localhost:8099/stats
```

## Installation

### Using Docker

```bash
# Build
docker build -t analytics-service .

# Run
docker run -p 8099:8099 analytics-service
```

### Local Development

```bash
# Install dependencies
go mod download

# Run
go run *.go
```

## Configuration

Environment variables:

- `PORT`: Server port (default: 8099)
- `LOG_LEVEL`: Logging level (default: INFO)

## Performance

### Benchmarks

- **Event Tracking**: < 5ms average
- **Metric Recording**: < 3ms average
- **Dashboard Query**: < 10ms average
- **WebSocket Update**: Real-time (1s interval)
- **Memory**: ~100MB for 100,000 events
- **Throughput**: 10,000+ events/second

## Usage Examples

### Frontend Integration

```javascript
// Analytics client
class AnalyticsClient {
  constructor() {
    this.baseUrl = 'http://localhost:8099';
  }

  async trackEvent(type, data, userId = null) {
    return fetch(`${this.baseUrl}/events`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        type,
        service: 'frontend',
        user_id: userId,
        data
      })
    });
  }

  async trackPageView(page, userId) {
    return this.trackEvent('page_view', { page }, userId);
  }

  async trackPurchase(orderId, total, userId) {
    return this.trackEvent('purchase', {
      order_id: orderId,
      total,
      currency: 'USD'
    }, userId);
  }

  async recordMetric(name, value, tags = {}) {
    return fetch(`${this.baseUrl}/metrics`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name,
        value,
        unit: 'ms',
        tags: { service: 'frontend', ...tags }
      })
    });
  }
}

// Usage
const analytics = new AnalyticsClient();

// Track page view
await analytics.trackPageView('/products', 'user-123');

// Track purchase
await analytics.trackPurchase('ORDER-456', 99.99, 'user-123');

// Record latency
await analytics.recordMetric('page_load_time', 1250, {
  page: '/products'
});
```

### Real-time Dashboard

```javascript
class RealtimeDashboard {
  constructor(containerElement) {
    this.container = containerElement;
    this.ws = new WebSocket('ws://localhost:8099/ws/dashboard');
    this.data = null;

    this.ws.onmessage = this.handleMessage.bind(this);
  }

  handleMessage(event) {
    const message = JSON.parse(event.data);

    if (message.type === 'snapshot' || message.type === 'dashboard_update') {
      this.data = message.data;
      this.render();
    } else if (message.type === 'event') {
      this.showEventNotification(message.data);
    }
  }

  render() {
    const { overview, services, realtime_stats } = this.data;

    this.container.innerHTML = `
      <div class="dashboard">
        <div class="overview">
          <h2>Overview</h2>
          <div class="metrics">
            <div class="metric">
              <span class="value">${overview.total_requests.toLocaleString()}</span>
              <span class="label">Total Requests</span>
            </div>
            <div class="metric">
              <span class="value">${overview.active_users}</span>
              <span class="label">Active Users</span>
            </div>
            <div class="metric">
              <span class="value">${overview.average_latency_ms.toFixed(1)}ms</span>
              <span class="label">Avg Latency</span>
            </div>
            <div class="metric">
              <span class="value">${overview.error_rate.toFixed(2)}%</span>
              <span class="label">Error Rate</span>
            </div>
          </div>
        </div>

        <div class="services">
          <h2>Services</h2>
          ${Object.values(services).map(s => this.renderService(s)).join('')}
        </div>

        <div class="realtime">
          <h2>Real-time Activity</h2>
          <p>Requests (1min): ${realtime_stats.requests_last_1min}</p>
          <p>Errors (1min): ${realtime_stats.errors_last_1min}</p>

          <h3>Top Endpoints</h3>
          <ul>
            ${realtime_stats.top_endpoints.map(e => `
              <li>${e.endpoint}: ${e.count} requests (${e.avg_time_ms.toFixed(1)}ms avg)</li>
            `).join('')}
          </ul>
        </div>
      </div>
    `;
  }

  renderService(service) {
    const statusClass = service.status === 'healthy' ? 'healthy' : 'unhealthy';

    return `
      <div class="service ${statusClass}">
        <h3>${service.name}</h3>
        <span class="status">${service.status}</span>
        <p>Requests: ${service.request_count}</p>
        <p>Errors: ${service.error_count}</p>
        <p>Latency: ${service.avg_latency_ms.toFixed(1)}ms</p>
        <p>Uptime: ${service.uptime_hours.toFixed(1)}h</p>
      </div>
    `;
  }

  showEventNotification(event) {
    console.log('New event:', event);
    // Show toast notification, etc.
  }
}

// Initialize dashboard
const dashboard = new RealtimeDashboard(
  document.getElementById('dashboard-container')
);
```

### Backend Integration

```go
// Send heartbeat from Go service
func sendHeartbeat() {
    heartbeat := map[string]interface{}{
        "service": "productcatalog",
        "status":  "healthy",
        "metrics": map[string]interface{}{
            "request_count": requestCount,
            "error_count":   errorCount,
            "avg_latency":   avgLatency,
            "uptime":        time.Since(startTime).Hours(),
        },
    }

    body, _ := json.Marshal(heartbeat)

    http.Post(
        "http://analytics:8099/heartbeat",
        "application/json",
        bytes.NewBuffer(body),
    )
}

// Send heartbeat every minute
go func() {
    ticker := time.NewTicker(1 * time.Minute)
    for range ticker.C {
        sendHeartbeat()
    }
}()
```

```python
# Send metrics from Python service
import requests
import time

def track_request_metric(endpoint, latency_ms):
    requests.post('http://analytics:8099/metrics', json={
        'name': 'latency',
        'value': latency_ms,
        'unit': 'ms',
        'tags': {
            'service': 'visualsearch',
            'endpoint': endpoint
        }
    })

# Decorator for tracking
def track_performance(func):
    def wrapper(*args, **kwargs):
        start = time.time()
        result = func(*args, **kwargs)
        duration = (time.time() - start) * 1000

        track_request_metric(func.__name__, duration)

        return result
    return wrapper

@track_performance
def search_products(query):
    # Your search logic
    pass
```

## Integration with Other Services

### Visual Search Integration

```javascript
// Track visual searches
await analytics.trackEvent('visual_search', {
  image_size: imageFile.size,
  results_count: results.length,
  top_similarity: results[0].similarity_score
}, userId);
```

### Gamification Integration

```javascript
// Track point awards
await analytics.trackEvent('points_awarded', {
  points: reward.total_points,
  action: 'purchase',
  level_up: reward.leveled_up
}, userId);
```

### Search Service Integration

```javascript
// Track search queries
await analytics.trackEvent('search', {
  query: searchQuery,
  results_count: results.length,
  autocomplete_used: true
}, userId);
```

### Inventory Integration

```javascript
// Track inventory changes
await analytics.trackEvent('inventory_update', {
  product_id: productId,
  warehouse: warehouse,
  change: change,
  new_quantity: newQuantity
});
```

## Monitoring & Alerts

### Setting Up Alerts

```javascript
// Monitor error rate
setInterval(async () => {
  const dashboard = await fetch('http://localhost:8099/dashboard').then(r => r.json());

  if (dashboard.overview.error_rate > 5.0) {
    // Send alert
    sendAlert('High error rate detected', {
      error_rate: dashboard.overview.error_rate,
      timestamp: new Date()
    });
  }
}, 60000); // Check every minute
```

### Service Health Monitoring

```javascript
// Check service health
setInterval(async () => {
  const dashboard = await fetch('http://localhost:8099/dashboard').then(r => r.json());

  for (const [name, service] of Object.entries(dashboard.services)) {
    if (service.status !== 'healthy') {
      console.warn(`Service ${name} is ${service.status}`);
    }

    // Check if heartbeat is stale (> 2 minutes)
    const lastHeartbeat = new Date(service.last_heartbeat);
    const now = new Date();
    if ((now - lastHeartbeat) > 2 * 60 * 1000) {
      console.error(`Service ${name} heartbeat is stale`);
    }
  }
}, 30000); // Check every 30 seconds
```

## Data Retention

- **Events**: Last 10,000 events in memory
- **Metrics**: Last 10,000 metrics in memory
- **Hourly Stats**: Last 7 days
- **Daily Stats**: Last 90 days
- **Real-time Data**: Last 1 minute

## Future Enhancements

- [ ] InfluxDB backend for long-term storage
- [ ] Grafana integration for visualization
- [ ] Alert rules and notifications
- [ ] Custom dashboard builder
- [ ] Export to CSV/JSON
- [ ] Anomaly detection with ML
- [ ] Distributed tracing integration
- [ ] Log aggregation

## License

Apache-2.0
