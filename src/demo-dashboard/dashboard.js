// Configuration
const API_BASE = 'http://localhost:8080'; // API Gateway
const WS_BASE = 'ws://localhost'; // WebSocket base

// State
let inventoryWS = null;
let analyticsWS = null;
let searchWS = null;

// Utility Functions
function log(message, type = 'info') {
  const logDiv = document.getElementById('activity-log');
  const entry = document.createElement('div');
  entry.className = `log-entry ${type}`;
  entry.textContent = `[${new Date().toLocaleTimeString()}] ${message}`;
  logDiv.insertBefore(entry, logDiv.firstChild);

  // Keep only last 100 entries
  while (logDiv.children.length > 100) {
    logDiv.removeChild(logDiv.lastChild);
  }
}

async function fetchAPI(endpoint) {
  try {
    const response = await fetch(`${API_BASE}${endpoint}`);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    return await response.json();
  } catch (error) {
    log(`API Error: ${error.message}`, 'error');
    throw error;
  }
}

// Service Status Monitoring
async function updateServiceStatus() {
  try {
    const health = await fetchAPI('/health');

    const statusDiv = document.getElementById('service-status');
    statusDiv.innerHTML = '';

    health.services.forEach(service => {
      const card = document.createElement('div');
      card.className = `service-card ${service.available ? 'healthy' : 'unhealthy'}`;
      card.innerHTML = `
        <h3>${service.name}</h3>
        <span class="status ${service.status}">${service.status}</span>
        <div class="port">${service.url}</div>
      `;
      statusDiv.appendChild(card);
    });

    log('Service status updated', 'success');
  } catch (error) {
    log('Failed to update service status', 'error');
  }
}

// Metrics Monitoring
async function updateMetrics() {
  try {
    const dashboard = await fetchAPI('/analytics/dashboard');

    document.getElementById('total-requests').textContent =
      dashboard.overview.total_requests.toLocaleString();

    document.getElementById('active-users').textContent =
      dashboard.overview.active_users;

    document.getElementById('avg-latency').textContent =
      `${dashboard.overview.average_latency_ms.toFixed(1)}ms`;

    document.getElementById('error-rate').textContent =
      `${dashboard.overview.error_rate.toFixed(2)}%`;
  } catch (error) {
    log('Failed to update metrics', 'error');
  }
}

// Tab Management
function initTabs() {
  const tabButtons = document.querySelectorAll('.tab-button');
  const tabContents = document.querySelectorAll('.tab-content');

  tabButtons.forEach(button => {
    button.addEventListener('click', () => {
      const targetTab = button.getAttribute('data-tab');

      // Update active states
      tabButtons.forEach(btn => btn.classList.remove('active'));
      tabContents.forEach(content => content.classList.remove('active'));

      button.classList.add('active');
      document.getElementById(`tab-${targetTab}`).classList.add('active');

      log(`Switched to ${targetTab} tab`, 'info');
    });
  });
}

// Search Autocomplete
function initSearch() {
  const searchInput = document.getElementById('search-input');
  const resultsDiv = document.getElementById('autocomplete-results');
  let debounceTimer;

  searchInput.addEventListener('input', async (e) => {
    clearTimeout(debounceTimer);

    const query = e.target.value;
    if (query.length < 2) {
      resultsDiv.classList.remove('show');
      return;
    }

    debounceTimer = setTimeout(async () => {
      try {
        const data = await fetchAPI(`/search/autocomplete?q=${encodeURIComponent(query)}`);

        resultsDiv.innerHTML = '';
        data.suggestions.forEach(suggestion => {
          const item = document.createElement('div');
          item.className = 'autocomplete-item';
          item.innerHTML = `
            <strong>${suggestion.text}</strong>
            <span style="color: #666; margin-left: 10px;">${suggestion.category || ''}</span>
          `;
          item.addEventListener('click', () => {
            searchInput.value = suggestion.text;
            resultsDiv.classList.remove('show');
            log(`Selected: ${suggestion.text}`, 'info');
          });
          resultsDiv.appendChild(item);
        });

        resultsDiv.classList.add('show');
        log(`Found ${data.suggestions.length} suggestions for "${query}"`, 'success');
      } catch (error) {
        log(`Autocomplete error: ${error.message}`, 'error');
      }
    }, 300);
  });

  // Close dropdown when clicking outside
  document.addEventListener('click', (e) => {
    if (!searchInput.contains(e.target) && !resultsDiv.contains(e.target)) {
      resultsDiv.classList.remove('show');
    }
  });
}

// Trending Updates
async function updateTrending() {
  try {
    const data = await fetchAPI('/search/trending?period=1h');

    const trendingList = document.getElementById('trending-list');
    trendingList.innerHTML = '';

    data.trending.slice(0, 5).forEach(item => {
      const div = document.createElement('div');
      div.className = 'trending-item';
      div.innerHTML = `
        <span><span class="trending-rank">#${item.rank}</span> ${item.query}</span>
        <span style="color: #666;">${item.count} searches</span>
      `;
      trendingList.appendChild(div);
    });
  } catch (error) {
    log('Failed to update trending', 'error');
  }
}

// Visual Search
function initVisualSearch() {
  const uploadZone = document.getElementById('upload-zone');
  const fileInput = document.getElementById('image-upload');
  const resultsDiv = document.getElementById('visual-search-results');

  uploadZone.addEventListener('click', () => fileInput.click());

  uploadZone.addEventListener('dragover', (e) => {
    e.preventDefault();
    uploadZone.classList.add('dragover');
  });

  uploadZone.addEventListener('dragleave', () => {
    uploadZone.classList.remove('dragover');
  });

  uploadZone.addEventListener('drop', async (e) => {
    e.preventDefault();
    uploadZone.classList.remove('dragover');

    const file = e.dataTransfer.files[0];
    if (file && file.type.startsWith('image/')) {
      await performVisualSearch(file);
    }
  });

  fileInput.addEventListener('change', async (e) => {
    const file = e.target.files[0];
    if (file) {
      await performVisualSearch(file);
    }
  });

  async function performVisualSearch(file) {
    try {
      log('Uploading image for visual search...', 'info');

      const formData = new FormData();
      formData.append('image', file);

      const response = await fetch(`${API_BASE}/visualsearch/search`, {
        method: 'POST',
        body: formData
      });

      if (!response.ok) throw new Error(`HTTP ${response.status}`);

      const data = await response.json();

      resultsDiv.innerHTML = '<h4>Similar Products:</h4>';
      data.results.forEach(result => {
        const item = document.createElement('div');
        item.className = 'result-item';
        item.innerHTML = `
          <strong>${result.product_name || result.product_id}</strong>
          <div>Similarity: ${(result.similarity_score * 100).toFixed(1)}%</div>
          <div>Price: $${result.price || 'N/A'}</div>
        `;
        resultsDiv.appendChild(item);
      });

      log(`Found ${data.results.length} similar products`, 'success');
    } catch (error) {
      log(`Visual search failed: ${error.message}`, 'error');
      resultsDiv.innerHTML = `<p style="color: red;">Error: ${error.message}</p>`;
    }
  }
}

// Gamification
function initGamification() {
  const userIdInput = document.getElementById('user-id-input');
  const loadProfileBtn = document.getElementById('load-profile-btn');
  const awardPointsBtn = document.getElementById('award-points-btn');
  const spinWheelBtn = document.getElementById('spin-wheel-btn');
  const dailyMissionBtn = document.getElementById('daily-mission-btn');
  const statsDiv = document.getElementById('user-stats');
  const resultsDiv = document.getElementById('gamification-results');

  loadProfileBtn.addEventListener('click', async () => {
    const userId = userIdInput.value;
    if (!userId) return;

    try {
      log(`Loading profile for ${userId}...`, 'info');

      const progress = await fetchAPI(`/gamification/users/${userId}/progress`);

      statsDiv.innerHTML = `
        <div class="stat-box">
          <div class="stat-value">${progress.level}</div>
          <div class="stat-label">Level</div>
        </div>
        <div class="stat-box">
          <div class="stat-value">${progress.points}</div>
          <div class="stat-label">Points</div>
        </div>
        <div class="stat-box">
          <div class="stat-value">${progress.xp}</div>
          <div class="stat-label">XP</div>
        </div>
        <div class="stat-box">
          <div class="stat-value">🔥 ${progress.login_streak}</div>
          <div class="stat-label">Day Streak</div>
        </div>
      `;

      log(`Loaded profile for ${userId}`, 'success');
    } catch (error) {
      log(`Failed to load profile: ${error.message}`, 'error');
    }
  });

  awardPointsBtn.addEventListener('click', async () => {
    const userId = userIdInput.value;
    if (!userId) return;

    try {
      const response = await fetch(`${API_BASE}/gamification/users/${userId}/points`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          points: 100,
          action: 'demo',
          reason: 'Demo button click'
        })
      });

      const data = await response.json();

      resultsDiv.innerHTML = `
        <div class="result-item">
          <h4>🎁 Points Awarded!</h4>
          <p>Base Points: ${data.points}</p>
          <p>Multiplier: ${data.multiplier}x</p>
          <p>Total: ${data.total_points} points</p>
          ${data.leveled_up ? '<p style="color: green; font-weight: bold;">🎉 LEVEL UP!</p>' : ''}
        </div>
      `;

      log(`Awarded ${data.total_points} points to ${userId}`, 'success');

      // Reload profile
      loadProfileBtn.click();
    } catch (error) {
      log(`Failed to award points: ${error.message}`, 'error');
    }
  });

  spinWheelBtn.addEventListener('click', async () => {
    const userId = userIdInput.value;
    if (!userId) return;

    try {
      const data = await fetchAPI(`/gamification/users/${userId}/spin`);

      resultsDiv.innerHTML = `
        <div class="result-item">
          <h4>🎰 Lucky Spin Result!</h4>
          <p>Reward: ${data.reward.name}</p>
          <p>Value: ${data.reward.value}</p>
          <p>Rarity: ${data.reward.rarity}</p>
        </div>
      `;

      log(`Spin result: ${data.reward.name}`, 'success');
    } catch (error) {
      log(`Spin failed: ${error.message}`, 'error');
    }
  });

  dailyMissionBtn.addEventListener('click', async () => {
    try {
      const data = await fetchAPI('/gamification/missions/daily');

      resultsDiv.innerHTML = '<h4>📋 Daily Missions:</h4>';
      data.missions.forEach(mission => {
        const item = document.createElement('div');
        item.className = 'result-item';
        item.innerHTML = `
          <strong>${mission.name}</strong>
          <p>${mission.description}</p>
          <p>Progress: ${mission.progress}/${mission.target}</p>
          <p>Reward: ${mission.reward_points} points</p>
        `;
        resultsDiv.appendChild(item);
      });

      log('Loaded daily missions', 'success');
    } catch (error) {
      log(`Failed to load missions: ${error.message}`, 'error');
    }
  });
}

// Inventory WebSocket
function initInventory() {
  const productIdInput = document.getElementById('product-id-input');
  const checkBtn = document.getElementById('check-inventory-btn');
  const connectWsBtn = document.getElementById('connect-ws-btn');
  const wsStatus = document.getElementById('ws-status');
  const dataDiv = document.getElementById('inventory-data');
  const updatesDiv = document.getElementById('inventory-updates');

  checkBtn.addEventListener('click', async () => {
    const productId = productIdInput.value;
    if (!productId) return;

    try {
      log(`Checking inventory for ${productId}...`, 'info');

      const data = await fetchAPI(`/inventory/inventory/${productId}`);

      dataDiv.innerHTML = `
        <div class="data-item">
          <h4>${data.product_id}</h4>
          <p>Available Stock: ${data.available_stock}</p>
          <p>Total Stock: ${data.total_stock}</p>
          <p>Reserved: ${data.reserved_stock}</p>
          <h5>Warehouses:</h5>
          ${Object.entries(data.warehouses).map(([wh, qty]) =>
            `<div>${wh}: ${qty}</div>`
          ).join('')}
        </div>
      `;

      log(`Inventory loaded for ${productId}`, 'success');
    } catch (error) {
      log(`Failed to load inventory: ${error.message}`, 'error');
    }
  });

  connectWsBtn.addEventListener('click', () => {
    if (inventoryWS && inventoryWS.readyState === WebSocket.OPEN) {
      inventoryWS.close();
      return;
    }

    try {
      inventoryWS = new WebSocket(`${WS_BASE}:8092/ws`);

      inventoryWS.onopen = () => {
        wsStatus.textContent = 'WebSocket: Connected';
        wsStatus.classList.remove('offline');
        wsStatus.classList.add('online');
        connectWsBtn.textContent = 'Disconnect WebSocket';
        log('Connected to inventory WebSocket', 'success');
      };

      inventoryWS.onmessage = (event) => {
        const message = JSON.parse(event.data);

        if (message.type === 'snapshot') {
          updatesDiv.innerHTML = '<h5>Initial Snapshot Received</h5>';
        } else if (message.type === 'update') {
          const update = document.createElement('div');
          update.className = 'inventory-update new';
          update.innerHTML = `
            <strong>${message.data.product_id}</strong>
            <div>Warehouse: ${message.data.warehouse}</div>
            <div>Change: ${message.data.change}</div>
            <div>New Quantity: ${message.data.quantity}</div>
          `;
          updatesDiv.insertBefore(update, updatesDiv.firstChild);

          log(`Inventory update: ${message.data.product_id}`, 'info');
        }
      };

      inventoryWS.onclose = () => {
        wsStatus.textContent = 'WebSocket: Disconnected';
        wsStatus.classList.remove('online');
        wsStatus.classList.add('offline');
        connectWsBtn.textContent = 'Connect WebSocket';
        log('Disconnected from inventory WebSocket', 'info');
      };

      inventoryWS.onerror = (error) => {
        log('WebSocket error', 'error');
      };
    } catch (error) {
      log(`Failed to connect WebSocket: ${error.message}`, 'error');
    }
  });
}

// Analytics Dashboard
async function updateAnalyticsDashboard() {
  try {
    const dashboard = await fetchAPI('/analytics/dashboard');

    // Service Health
    const healthChart = document.getElementById('service-health-chart');
    healthChart.innerHTML = '';
    Object.values(dashboard.services).slice(0, 6).forEach(service => {
      const item = document.createElement('div');
      item.style.padding = '8px';
      item.style.margin = '4px 0';
      item.style.background = service.status === 'healthy' ? '#d4edda' : '#f8d7da';
      item.style.borderRadius = '4px';
      item.innerHTML = `<strong>${service.name}</strong>: ${service.status}`;
      healthChart.appendChild(item);
    });

    // Recent Events
    const events = await fetchAPI('/analytics/events');
    const eventsDiv = document.getElementById('recent-events');
    eventsDiv.innerHTML = '';
    events.events.slice(0, 5).forEach(event => {
      const item = document.createElement('div');
      item.style.padding = '8px';
      item.style.margin = '4px 0';
      item.style.background = '#f8f9fa';
      item.style.borderRadius = '4px';
      item.style.fontSize = '0.85rem';
      item.textContent = `${event.type} - ${event.service}`;
      eventsDiv.appendChild(item);
    });

    // Top Endpoints
    const endpointsDiv = document.getElementById('top-endpoints');
    endpointsDiv.innerHTML = '';
    const endpoints = dashboard.realtime_stats.top_endpoints || [];
    endpoints.slice(0, 5).forEach(endpoint => {
      const item = document.createElement('div');
      item.style.padding = '8px';
      item.style.margin = '4px 0';
      item.style.background = '#f8f9fa';
      item.style.borderRadius = '4px';
      item.innerHTML = `
        <strong>${endpoint.endpoint}</strong>
        <div style="font-size: 0.85rem; color: #666;">
          ${endpoint.count} requests - ${endpoint.avg_time_ms.toFixed(1)}ms avg
        </div>
      `;
      endpointsDiv.appendChild(item);
    });

    // Real-time Activity
    const activityDiv = document.getElementById('realtime-activity');
    activityDiv.innerHTML = `
      <div style="padding: 8px; margin: 4px 0; background: #f8f9fa; border-radius: 4px;">
        Requests (1min): ${dashboard.realtime_stats.requests_last_1min}
      </div>
      <div style="padding: 8px; margin: 4px 0; background: #f8f9fa; border-radius: 4px;">
        Errors (1min): ${dashboard.realtime_stats.errors_last_1min}
      </div>
    `;
  } catch (error) {
    log('Failed to update analytics dashboard', 'error');
  }
}

// Initialize
document.addEventListener('DOMContentLoaded', () => {
  log('Dashboard initializing...', 'info');

  // Init components
  initTabs();
  initSearch();
  initVisualSearch();
  initGamification();
  initInventory();

  // Initial updates
  updateServiceStatus();
  updateMetrics();
  updateTrending();
  updateAnalyticsDashboard();

  // Periodic updates
  setInterval(updateServiceStatus, 30000); // Every 30s
  setInterval(updateMetrics, 5000); // Every 5s
  setInterval(updateTrending, 10000); // Every 10s
  setInterval(updateAnalyticsDashboard, 10000); // Every 10s

  log('Dashboard initialized successfully!', 'success');
});

// Cleanup
window.addEventListener('beforeunload', () => {
  if (inventoryWS) inventoryWS.close();
  if (analyticsWS) analyticsWS.close();
  if (searchWS) searchWS.close();
});
