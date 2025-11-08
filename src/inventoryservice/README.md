# Real-time Inventory Service

Real-time inventory tracking and synchronization across multiple warehouses.

## Features

- 📊 **Real-time Updates**: WebSocket for instant inventory changes
- 🏢 **Multi-warehouse**: Track stock across multiple locations
- 🔒 **Stock Reservation**: Reserve items during checkout
- 📈 **Live Dashboard**: Real-time inventory monitoring
- ⚡ **Fast Sync**: Instant updates to all connected clients

## API Endpoints

### Get All Inventory
```bash
GET /inventory
```

### Get Product Inventory
```bash
GET /inventory/{product_id}
```

Response:
```json
{
  "product_id": "OLJCESPC7Z",
  "name": "Sunglasses",
  "total_stock": 350,
  "reserved_stock": 20,
  "available_stock": 330,
  "warehouses": {
    "US-WEST": 150,
    "US-EAST": 200
  },
  "last_updated": "2024-01-15T10:30:00Z"
}
```

### Update Inventory
```bash
POST /inventory/{product_id}/update
```

Request:
```json
{
  "warehouse": "US-WEST",
  "change": -5,
  "update_type": "sale"
}
```

### Reserve Stock
```bash
POST /inventory/{product_id}/reserve
```

Request:
```json
{
  "quantity": 2
}
```

## WebSocket API

Connect to `/ws` for real-time updates.

### Initial Snapshot
Upon connection, receive current inventory:
```json
{
  "type": "snapshot",
  "data": [/* all products */]
}
```

### Real-time Updates
Receive updates when inventory changes:
```json
{
  "type": "update",
  "data": {
    "product_id": "OLJCESPC7Z",
    "warehouse": "US-WEST",
    "quantity": 145,
    "change": -5,
    "timestamp": "2024-01-15T10:30:00Z",
    "update_type": "sale"
  }
}
```

## Usage Examples

### JavaScript Client

```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8092/ws');

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);

  if (message.type === 'snapshot') {
    // Initialize inventory display
    displayInventory(message.data);
  } else if (message.type === 'update') {
    // Update specific product
    updateProductDisplay(message.data);

    // Show notification
    if (message.data.available_stock < 10) {
      showLowStockAlert(message.data.product_id);
    }
  }
};

// Update inventory via REST API
async function updateInventory(productId, warehouse, change, type) {
  await fetch(`http://localhost:8092/inventory/${productId}/update`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      warehouse: warehouse,
      change: change,
      update_type: type
    })
  });
}

// Reserve stock during checkout
async function reserveStock(productId, quantity) {
  const response = await fetch(`http://localhost:8092/inventory/${productId}/reserve`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ quantity })
  });

  return response.json();
}
```

### Real-time Dashboard

```html
<div id="inventory-dashboard">
  <div id="products"></div>
</div>

<script>
const ws = new WebSocket('ws://localhost:8092/ws');
const inventory = {};

ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);

  if (msg.type === 'snapshot') {
    msg.data.forEach(product => {
      inventory[product.product_id] = product;
      renderProduct(product);
    });
  } else if (msg.type === 'update') {
    const product = inventory[msg.data.product_id];
    if (product) {
      // Update display with animation
      updateProductDisplay(msg.data);
    }
  }
};

function renderProduct(product) {
  const html = `
    <div class="product" id="product-${product.product_id}">
      <h3>${product.name}</h3>
      <div class="stock ${product.available_stock < 50 ? 'low' : ''}">
        Available: <span class="count">${product.available_stock}</span>
      </div>
      <div class="warehouses">
        ${Object.entries(product.warehouses).map(([wh, qty]) => `
          <div>${wh}: ${qty}</div>
        `).join('')}
      </div>
    </div>
  `;

  document.getElementById('products').innerHTML += html;
}

function updateProductDisplay(update) {
  const el = document.querySelector(`#product-${update.product_id} .count`);
  el.textContent = update.quantity;

  // Flash animation
  el.classList.add('updated');
  setTimeout(() => el.classList.remove('updated'), 1000);
}
</script>

<style>
.stock.low { color: red; font-weight: bold; }
.updated { animation: flash 1s; }
@keyframes flash {
  0%, 100% { background: transparent; }
  50% { background: yellow; }
}
</style>
```

## Running

```bash
# Local
go run *.go

# Docker
docker build -t inventoryservice .
docker run -p 8092:8092 inventoryservice
```

## Update Types

- `sale`: Product sold
- `restock`: New stock arrived
- `adjustment`: Manual inventory adjustment
- `return`: Product returned
- `damage`: Stock damaged/lost

## Testing

```bash
# Get inventory
curl http://localhost:8092/inventory

# Update stock (sale)
curl -X POST http://localhost:8092/inventory/OLJCESPC7Z/update \
  -H "Content-Type: application/json" \
  -d '{"warehouse": "US-WEST", "change": -5, "update_type": "sale"}'

# Reserve stock
curl -X POST http://localhost:8092/inventory/OLJCESPC7Z/reserve \
  -H "Content-Type: application/json" \
  -d '{"quantity": 2}'

# WebSocket test (using websocat)
websocat ws://localhost:8092/ws
```

## Integration

### On Checkout

```javascript
// Before checkout, reserve stock
const reservation = await reserveStock(productId, quantity);

if (!reservation.success) {
  alert('Sorry, this item is out of stock!');
  return;
}

// Proceed with checkout
processCheckout();
```

### Auto-Reorder

```javascript
ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);

  if (msg.type === 'update') {
    const product = inventory[msg.data.product_id];

    // Auto-reorder if stock is low
    if (product.available_stock < product.reorder_point) {
      createPurchaseOrder(msg.data.product_id);
    }
  }
};
```
