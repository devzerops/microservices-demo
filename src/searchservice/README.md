# Search Service with Autocomplete & Trending

Real-time search autocomplete service with trending queries tracking, fuzzy matching, and advanced product search.

## Features

- 🔍 **Fast Autocomplete**: Trie-based prefix matching with sub-5ms response time
- 🔎 **Advanced Product Search**: Multi-criteria search with filters and sorting
- 💰 **Price Range Filtering**: Search within specific price ranges
- ⭐ **Rating Filtering**: Filter by minimum rating threshold
- 🏷️ **Category Filtering**: Filter by product categories
- 📦 **Stock Filtering**: Filter in-stock products only
- 🔢 **Pagination**: Efficient handling of large result sets
- 📈 **Trending Queries**: Real-time tracking of popular searches
- 🎯 **Fuzzy Matching**: Levenshtein distance-based typo correction
- 📊 **Search History**: User-specific search history tracking
- 🔄 **Real-time Updates**: WebSocket streaming for trending data
- ⚡ **High Performance**: In-memory data structures for instant results

## Architecture

### Core Components

**1. Product Index**
- Advanced product search engine
- Multi-field relevance scoring (name, description, category)
- Multiple match types (exact, prefix, partial, fuzzy)
- Comprehensive filtering (price, rating, category, stock)
- Multiple sorting options (relevance, price, rating, popularity, newest)
- Efficient pagination support
- Thread-safe concurrent operations with read-write locks

**2. Trie (Prefix Tree)**
- O(k) search complexity where k = query length
- Automatic prefix matching
- Weighted scoring based on popularity
- Category-based organization

**3. Trending Tracker**
- Time-windowed query tracking
- Velocity calculation (searches per minute)
- Rank change detection
- Automatic cleanup of old data

**4. Fuzzy Matcher**
- Levenshtein distance algorithm
- Typo tolerance (up to 3 character edits)
- Fallback suggestions when no exact matches

**5. Search History**
- Per-user history tracking
- Chronological ordering
- Duplicate detection
- Configurable retention

## API Endpoints

### Advanced Product Search

**GET** `/search?q=<query>&filters...`

Comprehensive product search with filtering, sorting, and pagination.

```bash
# Basic search
curl "http://localhost:8097/search?q=sunglasses"

# Search with price range
curl "http://localhost:8097/search?q=watch&min_price=20&max_price=100"

# Search with category and rating filter
curl "http://localhost:8097/search?q=&categories=accessories&min_rating=4.5"

# Search with sorting and pagination
curl "http://localhost:8097/search?q=mug&sort_by=price_asc&page=1&page_size=10"

# Combined filters
curl "http://localhost:8097/search?q=&categories=home&min_price=10&max_price=30&in_stock_only=true&sort_by=rating"
```

Response:
```json
{
  "query": "sunglasses",
  "results": [
    {
      "product": {
        "id": "OLJCESPC7Z",
        "name": "Sunglasses",
        "description": "Vintage sunglasses with UV protection",
        "category": "accessories",
        "price": 19.99,
        "rating": 4.5,
        "review_count": 156,
        "in_stock": true,
        "created_at": "2024-08-09T12:00:00Z",
        "popularity": 250
      },
      "score": 100.0,
      "relevance": 100.0,
      "match_type": "exact"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20,
  "total_pages": 1,
  "filters": {
    "sort_by": "relevance",
    "page": 1,
    "page_size": 20
  },
  "took_ms": 5
}
```

**Query Parameters**:
- `q` (optional): Search query (returns all products if empty)
- `categories` (optional): Filter by category (e.g., "accessories", "clothing", "home", "shoes")
- `min_price` (optional): Minimum price filter
- `max_price` (optional): Maximum price filter
- `min_rating` (optional): Minimum rating filter (1.0-5.0)
- `in_stock_only` (optional): Filter only in-stock products (true/false)
- `sort_by` (optional): Sort order - `relevance` (default), `price_asc`, `price_desc`, `rating`, `popularity`, `newest`
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Results per page (default: 20, max: 100)
- `user_id` (optional): User ID for analytics tracking

**Match Types**:
- `exact`: Exact name match (score: 100)
- `prefix`: Name starts with query (score: 90)
- `partial`: Name contains query (score: 80)
- `category`: Category matches query (score: 70)
- `description`: Description contains query (score: 60)
- `word`: Any word in name starts with query (score: 50)
- `fuzzy`: Fuzzy match using Levenshtein distance (score: calculated)
- `all`: Empty query returns all products (score: 50)

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

### Advanced Search Integration

```javascript
// Advanced product search with filters
class ProductSearch {
  constructor() {
    this.searchInput = document.getElementById('search-input');
    this.resultsContainer = document.getElementById('results');
    this.filtersForm = document.getElementById('filters-form');

    this.searchInput.addEventListener('input', this.handleSearch.bind(this));
    this.filtersForm.addEventListener('change', this.handleSearch.bind(this));
  }

  async handleSearch() {
    const query = this.searchInput.value;
    const filters = this.getFilters();

    const url = new URL('http://localhost:8097/search');
    url.searchParams.append('q', query);

    if (filters.category) url.searchParams.append('categories', filters.category);
    if (filters.minPrice) url.searchParams.append('min_price', filters.minPrice);
    if (filters.maxPrice) url.searchParams.append('max_price', filters.maxPrice);
    if (filters.minRating) url.searchParams.append('min_rating', filters.minRating);
    if (filters.inStockOnly) url.searchParams.append('in_stock_only', 'true');
    if (filters.sortBy) url.searchParams.append('sort_by', filters.sortBy);
    url.searchParams.append('page', filters.page || 1);

    const response = await fetch(url);
    const data = await response.json();

    this.displayResults(data);
  }

  getFilters() {
    return {
      category: document.getElementById('category-filter').value,
      minPrice: document.getElementById('min-price').value,
      maxPrice: document.getElementById('max-price').value,
      minRating: document.getElementById('min-rating').value,
      inStockOnly: document.getElementById('in-stock-only').checked,
      sortBy: document.getElementById('sort-by').value,
      page: 1
    };
  }

  displayResults(data) {
    const html = `
      <div class="search-stats">
        Found ${data.total} results in ${data.took_ms}ms
      </div>
      <div class="results-grid">
        ${data.results.map(result => `
          <div class="product-card">
            <h3>${result.product.name}</h3>
            <p class="category">${result.product.category}</p>
            <p class="price">$${result.product.price.toFixed(2)}</p>
            <div class="rating">⭐ ${result.product.rating}</div>
            <div class="match-info">
              Match: ${result.match_type} (${result.score.toFixed(0)})
            </div>
            ${!result.product.in_stock ? '<span class="out-of-stock">Out of Stock</span>' : ''}
          </div>
        `).join('')}
      </div>
      <div class="pagination">
        Page ${data.page} of ${data.total_pages}
      </div>
    `;

    this.resultsContainer.innerHTML = html;
  }
}

// Initialize
new ProductSearch();
```

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
