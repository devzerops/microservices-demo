# Experimental Services Integration Guide

This guide explains how to integrate the four new experimental services into the Online Boutique microservices demo.

## Services Overview

| Service | Port | Technology | Purpose |
|---------|------|------------|---------|
| Visual Search | 8093 | Python/FastAPI | Image-based product search using ML |
| Gamification | 8094 | Go | Points, badges, and rewards system |
| Real-time Inventory | 8092 | Go + WebSocket | Live stock updates across warehouses |
| PWA Service | 8095 | Node.js | Offline-first web app experience |

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         Frontend (8080)                          │
│  - Product browsing                                              │
│  - Shopping cart                                                 │
│  - Checkout flow                                                 │
└────────────┬────────────┬────────────┬────────────┬─────────────┘
             │            │            │            │
    ┌────────┴────┐  ┌────┴─────┐  ┌──┴───┐  ┌────┴─────┐
    │   Visual    │  │ Gamifi-  │  │ Real │  │   PWA    │
    │   Search    │  │ cation   │  │ time │  │ Service  │
    │  (8093)     │  │ (8094)   │  │ Inv. │  │ (8095)   │
    │             │  │          │  │(8092)│  │          │
    └─────────────┘  └──────────┘  └──────┘  └──────────┘
         │                │           │            │
         ├────────────────┴───────────┴────────────┤
         │         Existing Services               │
         │  - Product Catalog (ProductCatalog)     │
         │  - Shopping Cart (Cart)                 │
         │  - Checkout (Checkout)                  │
         │  - Recommendations (Recommendation)     │
         └─────────────────────────────────────────┘
```

## Quick Start

### 1. Using Docker Compose

```bash
# Build all services
docker-compose -f docker-compose-experimental.yml build

# Start all services
docker-compose -f docker-compose-experimental.yml up -d

# Check status
docker-compose -f docker-compose-experimental.yml ps

# View logs
docker-compose -f docker-compose-experimental.yml logs -f
```

### 2. Manual Setup

```bash
# Visual Search Service
cd src/visualsearchservice
pip install -r requirements.txt
uvicorn app.main:app --host 0.0.0.0 --port 8093

# Gamification Service
cd src/gamificationservice
go run *.go

# Real-time Inventory Service
cd src/inventoryservice
go run *.go

# PWA Service
cd src/pwa-service
npm install
npm start
```

## Integration Examples

### 1. Visual Search Integration

#### Frontend Integration

Add image upload to product search page:

```html
<!-- Add to frontend templates -->
<div class="visual-search">
  <h3>Search by Image</h3>
  <input type="file" id="image-upload" accept="image/*">
  <button onclick="visualSearch()">Search Similar Products</button>
  <div id="search-results"></div>
</div>

<script>
async function visualSearch() {
  const fileInput = document.getElementById('image-upload');
  const file = fileInput.files[0];

  if (!file) {
    alert('Please select an image');
    return;
  }

  const formData = new FormData();
  formData.append('image', file);
  formData.append('top_k', '10');

  try {
    const response = await fetch('http://localhost:8093/search', {
      method: 'POST',
      body: formData
    });

    const data = await response.json();
    displayResults(data.results);
  } catch (error) {
    console.error('Visual search failed:', error);
  }
}

function displayResults(results) {
  const container = document.getElementById('search-results');
  container.innerHTML = results.map(product => `
    <div class="product-card">
      <h4>${product.product_name}</h4>
      <p>Similarity: ${(product.similarity_score * 100).toFixed(1)}%</p>
      <p>Price: ${product.price}</p>
      <a href="/product/${product.product_id}">View Product</a>
    </div>
  `).join('');
}
</script>
```

#### Index Products on Startup

```python
# Add to productcatalogservice initialization
import requests

def index_products_for_visual_search():
    """Index all products in visual search service"""
    products = get_all_products()  # Your existing function

    product_data = [
        {
            "product_id": p.id,
            "name": p.name,
            "price": float(p.price_usd.units + p.price_usd.nanos / 1e9),
            "image_url": f"http://productcatalog:3550/images/{p.id}.jpg"
        }
        for p in products
    ]

    response = requests.post(
        'http://visualsearch:8093/index',
        json={"products": product_data}
    )

    print(f"Indexed {len(product_data)} products for visual search")

# Call on service startup
index_products_for_visual_search()
```

### 2. Gamification Integration

#### Award Points on Purchase

```go
// Add to checkoutservice after successful order
func awardPointsForOrder(userID string, orderTotal float64) {
    points := int(orderTotal * 10) // 10 points per dollar

    rewardReq := map[string]interface{}{
        "points": points,
        "action": "purchase",
        "reason": "Order completed",
    }

    body, _ := json.Marshal(rewardReq)

    resp, err := http.Post(
        fmt.Sprintf("http://gamification:8094/users/%s/points", userID),
        "application/json",
        bytes.NewBuffer(body),
    )

    if err != nil {
        log.Printf("Failed to award points: %v", err)
        return
    }

    var reward PointsReward
    json.NewDecoder(resp.Body).Decode(&reward)

    log.Printf("Awarded %d points to user %s", reward.TotalPoints, userID)

    // Show notification if user leveled up
    if reward.LeveledUp {
        notifyUser(userID, "Level Up!", fmt.Sprintf("You're now level %d!", reward.NewLevel))
    }
}
```

#### Display User Progress in Frontend

```html
<!-- Add to user profile page -->
<div class="user-gamification">
  <div id="user-stats"></div>
  <div id="user-badges"></div>
  <div id="daily-missions"></div>
</div>

<script>
async function loadUserGamification(userId) {
  // Get user progress
  const progressRes = await fetch(`http://localhost:8094/users/${userId}/progress`);
  const progress = await progressRes.json();

  // Get badges
  const badgesRes = await fetch(`http://localhost:8094/users/${userId}/badges`);
  const badges = await badgesRes.json();

  // Get missions
  const missionsRes = await fetch(`http://localhost:8094/missions/daily`);
  const missions = await missionsRes.json();

  displayUserStats(progress);
  displayBadges(badges);
  displayMissions(missions);
}

function displayUserStats(progress) {
  const html = `
    <div class="stats-card">
      <h3>Level ${progress.level}</h3>
      <div class="progress-bar">
        <div class="progress" style="width: ${(progress.xp / calculateNextLevelXP(progress.level)) * 100}%"></div>
      </div>
      <p>${progress.xp} / ${calculateNextLevelXP(progress.level)} XP</p>
      <p>Total Points: ${progress.points}</p>
      <p>🔥 Streak: ${progress.login_streak} days</p>
    </div>
  `;
  document.getElementById('user-stats').innerHTML = html;
}

function displayBadges(badges) {
  const html = badges.map(badge => `
    <div class="badge ${badge.rarity}">
      <span class="badge-icon">${badge.icon}</span>
      <h4>${badge.name}</h4>
      <p>${badge.description}</p>
      <small>Earned: ${new Date(badge.earned_at).toLocaleDateString()}</small>
    </div>
  `).join('');
  document.getElementById('user-badges').innerHTML = html;
}

function displayMissions(missions) {
  const html = missions.map(mission => `
    <div class="mission">
      <h4>${mission.name}</h4>
      <p>${mission.description}</p>
      <div class="progress-bar">
        <div class="progress" style="width: ${(mission.progress / mission.target) * 100}%"></div>
      </div>
      <p>${mission.progress} / ${mission.target}</p>
      <p>Reward: ${mission.reward_points} points</p>
    </div>
  `).join('');
  document.getElementById('daily-missions').innerHTML = html;
}
</script>
```

### 3. Real-time Inventory Integration

#### Connect to Inventory WebSocket

```javascript
// Add to frontend for real-time stock updates
class InventoryMonitor {
  constructor() {
    this.ws = null;
    this.inventory = {};
    this.connect();
  }

  connect() {
    this.ws = new WebSocket('ws://localhost:8092/ws');

    this.ws.onopen = () => {
      console.log('[Inventory] Connected to real-time updates');
    };

    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);

      if (message.type === 'snapshot') {
        // Initial inventory snapshot
        message.data.forEach(product => {
          this.inventory[product.product_id] = product;
        });
        this.renderInventory();
      } else if (message.type === 'update') {
        // Real-time update
        this.handleInventoryUpdate(message.data);
      }
    };

    this.ws.onerror = (error) => {
      console.error('[Inventory] WebSocket error:', error);
    };

    this.ws.onclose = () => {
      console.log('[Inventory] Connection closed, reconnecting...');
      setTimeout(() => this.connect(), 3000);
    };
  }

  handleInventoryUpdate(update) {
    const product = this.inventory[update.product_id];

    if (product) {
      // Update stock levels
      product.available_stock = update.available_stock;
      product.total_stock = update.total_stock;
      product.warehouses[update.warehouse] = update.quantity;

      // Update UI
      this.updateProductDisplay(update);

      // Show low stock alert
      if (update.available_stock < 10 && update.available_stock > 0) {
        this.showLowStockAlert(product);
      } else if (update.available_stock === 0) {
        this.showOutOfStockAlert(product);
      }
    }
  }

  updateProductDisplay(update) {
    const stockElement = document.querySelector(`[data-product-id="${update.product_id}"] .stock`);
    if (stockElement) {
      stockElement.textContent = `${update.available_stock} in stock`;
      stockElement.classList.add('updated');

      // Flash animation
      setTimeout(() => stockElement.classList.remove('updated'), 1000);

      // Change color based on stock level
      if (update.available_stock < 10) {
        stockElement.classList.add('low-stock');
      } else {
        stockElement.classList.remove('low-stock');
      }
    }
  }

  showLowStockAlert(product) {
    if (Notification.permission === 'granted') {
      new Notification('Low Stock Alert', {
        body: `Only ${product.available_stock} ${product.name} left!`,
        icon: '/images/icons/icon-192x192.png',
        tag: `low-stock-${product.product_id}`
      });
    }
  }

  showOutOfStockAlert(product) {
    const productCard = document.querySelector(`[data-product-id="${product.product_id}"]`);
    if (productCard) {
      const button = productCard.querySelector('.add-to-cart-button');
      button.disabled = true;
      button.textContent = 'Out of Stock';
    }
  }
}

// Initialize on page load
const inventoryMonitor = new InventoryMonitor();
```

#### Reserve Stock During Checkout

```go
// Add to checkoutservice before placing order
func reserveStock(items []*pb.CartItem) (bool, error) {
    for _, item := range items {
        reqBody := map[string]interface{}{
            "quantity": item.Quantity,
        }

        body, _ := json.Marshal(reqBody)

        resp, err := http.Post(
            fmt.Sprintf("http://inventory:8092/inventory/%s/reserve", item.ProductId),
            "application/json",
            bytes.NewBuffer(body),
        )

        if err != nil || resp.StatusCode != http.StatusOK {
            log.Printf("Failed to reserve stock for %s", item.ProductId)
            return false, fmt.Errorf("stock reservation failed")
        }

        var result map[string]interface{}
        json.NewDecoder(resp.Body).Decode(&result)

        if !result["reserved"].(bool) {
            return false, fmt.Errorf("insufficient stock for %s", item.ProductId)
        }
    }

    return true, nil
}

// Use in PlaceOrder handler
func (cs *checkoutService) PlaceOrder(ctx context.Context, req *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {
    // ... existing code ...

    // Reserve stock before proceeding
    reserved, err := reserveStock(req.Items)
    if !reserved {
        return nil, status.Errorf(codes.FailedPrecondition, "Unable to reserve stock: %v", err)
    }

    // ... continue with checkout ...
}
```

### 4. PWA Service Integration

#### Update Frontend to Be PWA-Ready

```html
<!-- Add to main HTML template head -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta name="description" content="Browse and shop products even when offline">
  <meta name="theme-color" content="#326ce5">

  <!-- PWA Manifest -->
  <link rel="manifest" href="http://localhost:8095/manifest.json">

  <!-- iOS Meta Tags -->
  <meta name="apple-mobile-web-app-capable" content="yes">
  <meta name="apple-mobile-web-app-status-bar-style" content="default">
  <meta name="apple-mobile-web-app-title" content="Boutique">
  <link rel="apple-touch-icon" href="http://localhost:8095/images/icons/icon-192x192.png">

  <!-- PWA Styles -->
  <link rel="stylesheet" href="http://localhost:8095/css/styles.css">

  <title>Online Boutique</title>
</head>
<body>
  <!-- Your existing content -->

  <!-- PWA UI Elements -->
  <div id="connection-status"></div>
  <button id="install-button">Install App</button>
  <div id="update-notification">
    <div class="content">
      <p>A new version is available!</p>
      <button id="reload-app">Update Now</button>
    </div>
  </div>

  <!-- Load PWA Manager -->
  <script src="http://localhost:8095/app.js"></script>

  <!-- Register Service Worker -->
  <script>
    if ('serviceWorker' in navigator) {
      navigator.serviceWorker.register('http://localhost:8095/service-worker.js')
        .then(registration => {
          console.log('[PWA] Service Worker registered:', registration);
        })
        .catch(error => {
          console.error('[PWA] Service Worker registration failed:', error);
        });
    }
  </script>
</body>
</html>
```

#### Offline Cart Synchronization

```javascript
// Wrap cart operations with offline support
async function addToCart(productId, quantity) {
  const item = { productId, quantity };

  if (navigator.onLine) {
    // Online - use normal API
    try {
      const response = await fetch('/cart', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(item)
      });

      if (!response.ok) throw new Error('Cart API failed');

      return await response.json();
    } catch (error) {
      // Fallback to offline storage
      return addToOfflineCart(item);
    }
  } else {
    // Offline - use PWA offline cart
    return addToOfflineCart(item);
  }
}

async function addToOfflineCart(item) {
  if (window.pwaManager) {
    await window.pwaManager.addToOfflineCart(item);
    showNotification('Added to cart (will sync when online)', 'info');
    return { success: true, offline: true };
  }

  throw new Error('PWA Manager not available');
}

// Sync offline cart when connection is restored
window.addEventListener('online', async () => {
  if (window.pwaManager) {
    const offlineCart = await window.pwaManager.getOfflineCart();

    if (offlineCart.length > 0) {
      showNotification('Syncing your cart...', 'info');
      await window.pwaManager.triggerBackgroundSync();
    }
  }
});
```

## Complete User Flow Example

### Scenario: User browses, adds to cart, and checks out

```javascript
// 1. User uploads image to find similar products
async function visualSearchFlow() {
  // Upload image
  const formData = new FormData();
  formData.append('image', imageFile);

  const searchResponse = await fetch('http://localhost:8093/search', {
    method: 'POST',
    body: formData
  });

  const { results } = await searchResponse.json();

  // 2. Display products with real-time inventory
  results.forEach(product => {
    displayProduct(product);

    // Check real-time inventory
    const inventoryData = inventoryMonitor.inventory[product.product_id];
    updateStockDisplay(product.product_id, inventoryData);
  });
}

// 3. User adds product to cart
async function addProductToCart(productId, quantity) {
  // Check inventory first
  const inventoryResponse = await fetch(`http://localhost:8092/inventory/${productId}`);
  const inventory = await inventoryResponse.json();

  if (inventory.available_stock < quantity) {
    alert('Not enough stock available!');
    return;
  }

  // Add to cart (works offline via PWA)
  await addToCart(productId, quantity);

  // Award points for adding to cart
  await fetch(`http://localhost:8094/users/${userId}/points`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      points: 5,
      action: 'add_to_cart',
      reason: 'Added item to cart'
    })
  });
}

// 4. User completes checkout
async function checkout(userId, items) {
  // Reserve inventory
  for (const item of items) {
    const reserveResponse = await fetch(
      `http://localhost:8092/inventory/${item.productId}/reserve`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ quantity: item.quantity })
      }
    );

    if (!reserveResponse.ok) {
      alert('Unable to reserve stock');
      return;
    }
  }

  // Process checkout
  const orderResponse = await fetch('/checkout', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ userId, items })
  });

  const order = await orderResponse.json();

  // Award points for purchase
  const total = calculateOrderTotal(items);
  const pointsResponse = await fetch(`http://localhost:8094/users/${userId}/points`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      points: Math.floor(total * 10),
      action: 'purchase',
      reason: `Order #${order.orderId}`
    })
  });

  const reward = await pointsResponse.json();

  // Check for level up or new badges
  if (reward.leveled_up) {
    showLevelUpAnimation(reward.new_level);
  }

  if (reward.new_badges && reward.new_badges.length > 0) {
    showNewBadges(reward.new_badges);
  }

  // Send push notification
  if (Notification.permission === 'granted') {
    new Notification('Order Confirmed!', {
      body: `Your order #${order.orderId} has been placed successfully`,
      icon: '/images/icons/icon-192x192.png',
      tag: `order-${order.orderId}`
    });
  }

  return order;
}
```

## Testing the Integration

### 1. Test Visual Search

```bash
# Index test products
curl -X POST http://localhost:8093/index \
  -H "Content-Type: application/json" \
  -d '{
    "products": [
      {
        "product_id": "OLJCESPC7Z",
        "name": "Sunglasses",
        "price": 19.99,
        "image_url": "http://localhost:3550/images/sunglasses.jpg"
      }
    ]
  }'

# Search by image
curl -X POST http://localhost:8093/search \
  -F "image=@test_image.jpg" \
  -F "top_k=5"
```

### 2. Test Gamification

```bash
# Award points
curl -X POST http://localhost:8094/users/user-123/points \
  -H "Content-Type: application/json" \
  -d '{
    "points": 100,
    "action": "purchase",
    "reason": "Completed order"
  }'

# Check user progress
curl http://localhost:8094/users/user-123/progress

# Get user badges
curl http://localhost:8094/users/user-123/badges
```

### 3. Test Real-time Inventory

```bash
# Get inventory
curl http://localhost:8092/inventory/OLJCESPC7Z

# Update stock
curl -X POST http://localhost:8092/inventory/OLJCESPC7Z/update \
  -H "Content-Type: application/json" \
  -d '{
    "warehouse": "US-WEST",
    "change": -5,
    "update_type": "sale"
  }'

# Reserve stock
curl -X POST http://localhost:8092/inventory/OLJCESPC7Z/reserve \
  -H "Content-Type: application/json" \
  -d '{"quantity": 2}'

# Connect to WebSocket
websocat ws://localhost:8092/ws
```

### 4. Test PWA Features

```bash
# Open in browser
open http://localhost:8095

# Test offline mode in DevTools:
# 1. Open DevTools (F12)
# 2. Go to Application > Service Workers
# 3. Check "Offline"
# 4. Reload page - should show offline page

# Test install prompt:
# 1. Open in Chrome
# 2. Look for install icon in address bar
# 3. Click to install

# Test push notifications:
# 1. Allow notifications when prompted
# 2. Open browser console
# 3. Run: new Notification('Test', { body: 'Testing push notifications' })
```

## Monitoring and Debugging

### Health Checks

```bash
# Check all services
curl http://localhost:8093/health  # Visual Search
curl http://localhost:8094/health  # Gamification
curl http://localhost:8092/health  # Inventory
curl http://localhost:8095/health  # PWA
```

### View Logs

```bash
# Docker logs
docker-compose -f docker-compose-experimental.yml logs -f visualsearch
docker-compose -f docker-compose-experimental.yml logs -f gamification
docker-compose -f docker-compose-experimental.yml logs -f inventory
docker-compose -f docker-compose-experimental.yml logs -f pwa

# Or all together
docker-compose -f docker-compose-experimental.yml logs -f
```

### Metrics

Each service exposes metrics endpoints:

```bash
# Service metrics
curl http://localhost:8093/metrics
curl http://localhost:8094/metrics
curl http://localhost:8092/metrics
curl http://localhost:8095/metrics
```

## Production Considerations

### 1. Security

- Enable HTTPS for all services
- Implement proper authentication
- Add rate limiting
- Configure CORS policies
- Sanitize user inputs

### 2. Scalability

- Add load balancers
- Use Redis for gamification storage
- Implement database for inventory
- Add CDN for PWA static assets
- Scale horizontally based on load

### 3. Performance

- Enable response caching
- Optimize ML model size
- Use connection pooling
- Implement request batching
- Add proper database indices

### 4. Monitoring

- Set up Prometheus metrics
- Configure alerting rules
- Monitor service health
- Track user engagement
- Log errors and warnings

## Troubleshooting

### Visual Search not finding products

- Verify products are indexed: `curl http://localhost:8093/index/status`
- Check image format (JPEG, PNG supported)
- Ensure images are accessible
- Check feature extraction logs

### Gamification points not awarded

- Verify user ID is correct
- Check service connectivity
- Review action/reason values
- Check logs for errors

### Real-time Inventory not updating

- Verify WebSocket connection
- Check network connectivity
- Review WebSocket logs
- Ensure broadcasts are enabled

### PWA offline mode not working

- Check service worker registration
- Verify HTTPS (required for SW)
- Clear browser cache
- Check caching strategy

## Next Steps

1. **Frontend Integration**: Update the main frontend service to include all four experimental services
2. **Database Migration**: Move from in-memory to persistent storage
3. **Authentication**: Add proper user authentication across services
4. **CI/CD**: Set up automated testing and deployment
5. **Monitoring**: Implement comprehensive monitoring and alerting

## Support

For issues or questions:
- Check service README files
- Review Docker logs
- Test individual endpoints
- Check browser console for PWA issues
