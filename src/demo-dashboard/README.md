# Demo Dashboard

Interactive web dashboard for testing and monitoring all experimental microservices.

## Features

- 📊 **Real-time Service Status**: Live health monitoring for all 6 services
- 📈 **Metrics Dashboard**: Real-time metrics from Analytics service
- 🔍 **Search Demo**: Interactive autocomplete and trending queries
- 📷 **Visual Search**: Upload images for product search
- 🎮 **Gamification**: Test points, badges, missions, and lucky wheel
- 📦 **Inventory**: Real-time stock updates via WebSocket
- 📝 **Activity Log**: Console-style logging of all actions

## Quick Start

### Using Node.js

```bash
# Install dependencies
npm install

# Start server
npm start

# Or with auto-reload
npm run dev
```

Open http://localhost:3000 in your browser.

### Using Docker

```bash
# Build
docker build -t demo-dashboard .

# Run
docker run -p 3000:3000 demo-dashboard
```

### Using Python (Alternative)

```bash
# No dependencies needed
python3 -m http.server 3000
```

## Prerequisites

All experimental services must be running:

```bash
# Start all services
make start-experimental

# Or with Docker Compose
docker-compose -f docker-compose-experimental.yml up -d
```

## Service Tabs

### 🔍 Search Tab

- **Autocomplete**: Type in the search box to see live suggestions
- **Trending**: View real-time trending search queries
- **Features**:
  - Fuzzy matching for typos
  - Category-based suggestions
  - Search history tracking

### 📷 Visual Search Tab

- **Upload**: Drag & drop or click to upload product images
- **Results**: See similar products with similarity scores
- **Features**:
  - MobileNetV2 feature extraction
  - FAISS similarity search
  - Top-K results with threshold

### 🎮 Gamification Tab

- **User Profile**: Load user stats (level, points, XP, streak)
- **Award Points**: Give 100 points to test user
- **Lucky Wheel**: Spin for random rewards
- **Daily Missions**: View available missions
- **Features**:
  - Level progression
  - Multipliers and streaks
  - Badge system
  - Mission tracking

### 📦 Inventory Tab

- **Check Stock**: Query product inventory levels
- **WebSocket**: Connect for real-time inventory updates
- **Features**:
  - Multi-warehouse tracking
  - Stock reservations
  - Live updates on changes

### 📊 Analytics Tab

- **Service Health**: Status of all services
- **Recent Events**: Latest tracked events
- **Top Endpoints**: Most-used API endpoints
- **Real-time Activity**: Requests and errors in last minute

## API Gateway Integration

All requests go through the API Gateway (port 8080):

```javascript
// Instead of:
fetch('http://localhost:8097/autocomplete?q=sun')

// Dashboard uses:
fetch('http://localhost:8080/search/autocomplete?q=sun')
```

Benefits:
- Single entry point
- Rate limiting (100 req/min per IP)
- Request logging
- CORS handling

## WebSocket Connections

The dashboard maintains WebSocket connections for real-time updates:

- **Inventory**: `ws://localhost:8092/ws`
- **Search Trending**: `ws://localhost:8097/ws/trending`
- **Analytics Dashboard**: `ws://localhost:8099/ws/dashboard`

## Configuration

Edit `dashboard.js` to customize endpoints:

```javascript
const API_BASE = 'http://localhost:8080'; // API Gateway
const WS_BASE = 'ws://localhost'; // WebSocket base
```

## Usage Examples

### Test Complete Workflow

1. **Search**: Type "sunglasses" in Search tab
2. **Visual**: Upload a sunglasses image in Visual Search tab
3. **Gamification**: Load profile and award points
4. **Inventory**: Check stock for product ID
5. **Analytics**: View all tracked events

### Monitor Services

1. Check service status cards at the top
2. Watch metrics update every 5 seconds
3. Connect to inventory WebSocket for live updates
4. Monitor activity log for all actions

### Test Rate Limiting

```javascript
// In browser console:
for (let i = 0; i < 150; i++) {
  fetch('http://localhost:8080/search/autocomplete?q=test');
}
// After 100 requests, you'll see 429 errors
```

## Screenshots

### Service Status
Shows all 6 services with health indicators (green = healthy, red = unhealthy).

### Metrics Dashboard
Real-time counters for:
- Total Requests
- Active Users
- Average Latency
- Error Rate

### Activity Log
Console-style log showing:
- ✅ Success messages (green)
- ℹ️ Info messages (blue)
- ❌ Error messages (red)

## Troubleshooting

### Services show as unhealthy

Check if services are running:
```bash
make health-check
```

Start services:
```bash
make start-experimental
```

### WebSocket won't connect

Check if inventory service is running:
```bash
curl http://localhost:8092/health
```

### Autocomplete not working

Check if search service is running:
```bash
curl http://localhost:8097/autocomplete?q=test
```

### CORS errors

Make sure API Gateway is running (provides CORS headers):
```bash
curl http://localhost:8080/health
```

## Development

### File Structure

```
demo-dashboard/
├── index.html       # Main HTML page
├── styles.css       # Stylesheet
├── dashboard.js     # Dashboard logic
├── server.js        # Simple HTTP server
├── package.json     # Node.js config
├── Dockerfile       # Container config
└── README.md        # This file
```

### Adding New Features

1. Add tab button in HTML:
```html
<button class="tab-button" data-tab="newfeature">🆕 New Feature</button>
```

2. Add tab content:
```html
<div class="tab-content" id="tab-newfeature">
  <!-- Your content -->
</div>
```

3. Add logic in `dashboard.js`:
```javascript
function initNewFeature() {
  // Your logic
}

// Call in DOMContentLoaded
initNewFeature();
```

## Performance

- **Load Time**: < 1 second
- **Update Frequency**: 5-30 seconds depending on component
- **WebSocket Overhead**: Minimal (<1KB per message)
- **Browser Support**: Modern browsers (Chrome, Firefox, Safari, Edge)

## Security Notes

- Dashboard is for **demo/development only**
- No authentication implemented
- CORS allows all origins
- Suitable for local testing, not production

## Future Enhancements

- [ ] Authentication (login system)
- [ ] Dark mode toggle
- [ ] Customizable metrics
- [ ] Export data to CSV
- [ ] Chart visualizations (Chart.js)
- [ ] Service comparison view
- [ ] Alert notifications
- [ ] Mobile responsive improvements

## License

Apache-2.0
