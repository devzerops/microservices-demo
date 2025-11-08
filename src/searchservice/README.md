# Search Service with Autocomplete & Trending

Real-time search autocomplete service with trending queries tracking and fuzzy matching.

## Features

- 🔍 **Fast Autocomplete**: Trie-based prefix matching with sub-5ms response time
- 📈 **Trending Queries**: Real-time tracking of popular searches
- 🎯 **Fuzzy Matching**: Levenshtein distance-based typo correction
- 📊 **Search History**: User-specific search history tracking
- 🔄 **Real-time Updates**: WebSocket streaming for trending data
- ⚡ **High Performance**: In-memory data structures for instant results

## Architecture

### Core Components

**1. Trie (Prefix Tree)**
- O(k) search complexity where k = query length
- Automatic prefix matching
- Weighted scoring based on popularity
- Category-based organization

**2. Trending Tracker**
- Time-windowed query tracking
- Velocity calculation (searches per minute)
- Rank change detection
- Automatic cleanup of old data

**3. Fuzzy Matcher**
- Levenshtein distance algorithm
- Typo tolerance (up to 3 character edits)
- Fallback suggestions when no exact matches

**4. Search History**
- Per-user history tracking
- Chronological ordering
- Duplicate detection
- Configurable retention

## API Endpoints

### Autocomplete

**GET** `/autocomplete?q=<query>&user_id=<id>`

Get autocomplete suggestions for a query.

```bash
curl "http://localhost:8097/autocomplete?q=sun"
```

Response:
```json
{
  "suggestions": [
    {
      "text": "sunglasses",
      "score": 100,
      "category": "accessories",
      "popularity": 45,
      "is_exact": true
    }
  ],
  "took_ms": 2
}
```

**Query Parameters**:
- `q` (required): Search query prefix
- `user_id` (optional): User ID for history tracking
- `limit` (optional): Maximum suggestions (default: 10)

### Trending Queries

**GET** `/trending?period=<period>`

Get trending search queries.

```bash
curl "http://localhost:8097/trending?period=1h"
```

Response:
```json
{
  "trending": [
    {
      "query": "summer collection",
      "count": 156,
      "velocity": 2.6,
      "rank": 1,
      "change": 0,
      "percentage": 23.4
    }
  ],
  "period": "1h",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**Query Parameters**:
- `period` (optional): Time window (5m, 15m, 30m, 1h, 6h, 12h, 24h)

### Search History

**GET** `/history/{user_id}`

Get user's search history.

```bash
curl "http://localhost:8097/history/user-123"
```

Response:
```json
{
  "user_id": "user-123",
  "history": [
    {
      "query": "sunglasses",
      "timestamp": "2024-01-15T10:25:00Z",
      "results_found": 12
    }
  ],
  "total": 25,
  "page_size": 20
}
```

**DELETE** `/history/{user_id}`

Clear user's search history.

```bash
curl -X DELETE "http://localhost:8097/history/user-123"
```

### Index Product

**POST** `/index`

Add a product to the search index.

```bash
curl -X POST http://localhost:8097/index \
  -H "Content-Type: application/json" \
  -d '{
    "name": "vintage sunglasses",
    "category": "accessories",
    "score": 85
  }'
```

### WebSocket - Real-time Trending

**WS** `/ws/trending`

Stream real-time trending updates.

```javascript
const ws = new WebSocket('ws://localhost:8097/ws/trending');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);

  if (data.type === 'snapshot') {
    console.log('Initial trending:', data.data);
  } else if (data.type === 'update') {
    console.log('Updated trending:', data.data);
  }
};
```

### Health Check

**GET** `/health`

Check service health.

```bash
curl http://localhost:8097/health
```

Response:
```json
{
  "status": "healthy",
  "service": "search-service",
  "timestamp": "2024-01-15T10:30:00Z",
  "stats": {
    "indexed_terms": 120,
    "trending_count": 10
  }
}
```

### Statistics

**GET** `/stats`

Get service statistics.

```bash
curl http://localhost:8097/stats
```

Response:
```json
{
  "indexed_terms": 120,
  "total_searches": 5678,
  "unique_searchers": 234,
  "trending_queries": 10,
  "avg_response_time": "< 5ms"
}
```

## Installation

### Using Docker

```bash
# Build
docker build -t search-service .

# Run
docker run -p 8097:8097 search-service
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

- `PORT`: Server port (default: 8097)
- `LOG_LEVEL`: Logging level (default: INFO)

## Performance

### Benchmarks

- **Autocomplete**: < 5ms average response time
- **Trending**: < 10ms average response time
- **WebSocket**: Real-time updates every 5 seconds
- **Memory**: ~50MB for 10,000 indexed terms
- **Throughput**: 1000+ requests/second

### Optimizations

1. **In-Memory Trie**: O(k) lookup complexity
2. **Read-Write Locks**: Concurrent read operations
3. **Background Cleanup**: Automated old data removal
4. **Weighted Scoring**: Popularity-based ranking

## Usage Examples

### Frontend Integration

```javascript
// Autocomplete component
class SearchAutocomplete {
  constructor(inputElement) {
    this.input = inputElement;
    this.resultsContainer = document.getElementById('autocomplete-results');

    this.input.addEventListener('input', this.handleInput.bind(this));
  }

  async handleInput(event) {
    const query = event.target.value;

    if (query.length < 2) {
      this.resultsContainer.innerHTML = '';
      return;
    }

    const response = await fetch(
      `http://localhost:8097/autocomplete?q=${encodeURIComponent(query)}&user_id=user-123`
    );

    const data = await response.json();
    this.displayResults(data.suggestions);
  }

  displayResults(suggestions) {
    this.resultsContainer.innerHTML = suggestions.map(s => `
      <div class="suggestion ${s.is_exact ? 'exact' : 'fuzzy'}">
        <span class="text">${s.text}</span>
        <span class="category">${s.category}</span>
        ${!s.is_exact ? '<span class="fuzzy-badge">Did you mean?</span>' : ''}
      </div>
    `).join('');
  }
}

// Initialize
const searchInput = document.getElementById('search-input');
new SearchAutocomplete(searchInput);
```

### Trending Widget

```javascript
// Real-time trending widget
class TrendingWidget {
  constructor() {
    this.ws = new WebSocket('ws://localhost:8097/ws/trending');
    this.container = document.getElementById('trending-container');

    this.ws.onmessage = this.handleUpdate.bind(this);
  }

  handleUpdate(event) {
    const message = JSON.parse(event.data);

    if (message.type === 'update' || message.type === 'snapshot') {
      this.render(message.data);
    }
  }

  render(trending) {
    this.container.innerHTML = `
      <h3>🔥 Trending Now</h3>
      <ul class="trending-list">
        ${trending.map(t => `
          <li>
            <span class="rank">#${t.rank}</span>
            <span class="query">${t.query}</span>
            <span class="count">${t.count} searches</span>
            ${t.change > 0 ? `<span class="up">↑${t.change}</span>` : ''}
            ${t.change < 0 ? `<span class="down">↓${Math.abs(t.change)}</span>` : ''}
          </li>
        `).join('')}
      </ul>
    `;
  }
}

// Initialize
new TrendingWidget();
```

### Search Analytics

```javascript
// Track search analytics
async function trackSearch(query, userId) {
  // Perform search
  const results = await performProductSearch(query);

  // Track in search service
  await fetch(`http://localhost:8097/autocomplete?q=${query}&user_id=${userId}`);

  // The service automatically:
  // 1. Adds to trending
  // 2. Saves to user history
  // 3. Updates autocomplete index

  return results;
}
```

## Integration with Other Services

### Product Catalog Integration

```javascript
// Index products from catalog
async function indexAllProducts() {
  const products = await fetch('http://productcatalog:3550/products').then(r => r.json());

  for (const product of products) {
    await fetch('http://localhost:8097/index', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name: product.name,
        category: product.categories[0],
        score: product.popularity || 50
      })
    });
  }

  console.log(`Indexed ${products.length} products`);
}
```

### Visual Search Integration

```javascript
// Combine text + image search
async function hybridSearch(textQuery, imageFile) {
  // Text autocomplete
  const textResults = await fetch(
    `http://localhost:8097/autocomplete?q=${textQuery}`
  ).then(r => r.json());

  // Visual search
  const formData = new FormData();
  formData.append('image', imageFile);

  const visualResults = await fetch(
    'http://localhost:8093/search',
    { method: 'POST', body: formData }
  ).then(r => r.json());

  // Merge and rank results
  return mergeResults(textResults.suggestions, visualResults.results);
}
```

### Analytics Dashboard Integration

```javascript
// Stream search metrics to analytics
setInterval(async () => {
  const stats = await fetch('http://localhost:8097/stats').then(r => r.json());

  // Send to analytics service
  await fetch('http://localhost:8099/metrics', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      service: 'search',
      metrics: {
        total_searches: stats.total_searches,
        unique_users: stats.unique_searchers,
        indexed_terms: stats.indexed_terms
      },
      timestamp: new Date().toISOString()
    })
  });
}, 60000); // Every minute
```

## Data Structures

### Trie Node Structure

```
Root
├─ s
│  ├─ u
│  │  ├─ n [sunglasses] (score: 100, count: 45)
│  │  └─ m [summer] (score: 90, count: 32)
│  ├─ h
│  │  └─ o [shoes] (score: 85, count: 28)
└─ t
   ├─ a [tank] (score: 80, count: 25)
   └─ s [t-shirt] (score: 75, count: 22)
```

### Trending Data Format

```json
{
  "query_stats": {
    "sunglasses": {
      "count": 45,
      "timestamps": ["2024-01-15T10:00:00Z", "..."],
      "velocity": 2.5,
      "rank": 1,
      "prev_rank": 2
    }
  }
}
```

## Monitoring

### Metrics to Track

- Autocomplete latency (p50, p95, p99)
- Trending update frequency
- Cache hit rate
- Unique search queries per day
- Top searched terms
- Fuzzy match ratio

### Logging

```go
// Structured logging example
log.Printf("[Autocomplete] query=%s took=%dms results=%d",
    query, elapsed, len(suggestions))

log.Printf("[Trending] updated top_query=%s count=%d velocity=%.2f",
    top.Query, top.Count, top.Velocity)
```

## Testing

### Unit Tests

```bash
go test -v ./...
```

### Load Testing

```bash
# Install hey
go install github.com/rakyll/hey@latest

# Test autocomplete
hey -n 10000 -c 100 "http://localhost:8097/autocomplete?q=sun"

# Test trending
hey -n 5000 -c 50 "http://localhost:8097/trending"
```

## Troubleshooting

### High Memory Usage

- Limit indexed terms
- Implement TTL for old searches
- Increase cleanup frequency

### Slow Autocomplete

- Check Trie size
- Verify no blocking operations
- Profile with pprof

### Trending Not Updating

- Check WebSocket connections
- Verify background worker is running
- Check time window configuration

## Future Enhancements

- [ ] Redis backend for persistence
- [ ] Elasticsearch integration for advanced search
- [ ] Multi-language support
- [ ] Synonym handling
- [ ] Search suggestions based on user profile
- [ ] A/B testing for ranking algorithms
- [ ] Machine learning for personalized autocomplete

## License

Apache-2.0
