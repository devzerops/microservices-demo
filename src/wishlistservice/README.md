# Wishlist Service

A comprehensive wishlist microservice for the Online Boutique e-commerce platform with price tracking, alerts, and sharing capabilities.

## Features

- **Wishlist Management**: Add, update, remove, and view favorite products
- **Price Tracking**: Automatic price history tracking for all wishlist items
- **Smart Alerts**:
  - Price drop notifications
  - Target price alerts
  - Restock notifications for out-of-stock items
- **Priority Levels**: Organize items by priority (high, medium, low)
- **Personal Notes**: Add notes/reminders for each item
- **Wishlist Sharing**: Share wishlists with friends or make them public
- **Statistics**: View total value, average savings, and item breakdown
- **Advanced Sorting**: Sort by price, priority, discount, or date added

## API Endpoints

### Wishlist Operations

#### Get User's Wishlist
```http
GET /users/{user_id}/wishlist?sort_by=priority

Query Parameters:
- sort_by: price_asc, price_desc, priority, discount, recent, oldest
```

#### Add Item to Wishlist
```http
POST /users/{user_id}/wishlist/items
Content-Type: application/json

{
  "product_id": "OLJCESPC7Z",
  "product_name": "Sunglasses",
  "current_price": 19.99,
  "target_price": 15.00,
  "priority": "high",
  "notes": "Wait for summer sale",
  "notify_on_price_drop": true,
  "notify_on_restock": false,
  "in_stock": true
}
```

#### Update Wishlist Item
```http
PUT /users/{user_id}/wishlist/items/{item_id}
Content-Type: application/json

{
  "target_price": 14.99,
  "priority": "medium",
  "notes": "Updated note",
  "notify_on_price_drop": true
}
```

#### Remove Item from Wishlist
```http
DELETE /users/{user_id}/wishlist/items/{item_id}
```

#### Get Specific Item
```http
GET /users/{user_id}/wishlist/items/{item_id}
```

### Statistics

#### Get Wishlist Statistics
```http
GET /users/{user_id}/wishlist/stats

Response:
{
  "user_id": "user123",
  "total_items": 15,
  "total_value": 457.85,
  "average_price_drop": 8.50,
  "items_on_sale": 4,
  "out_of_stock_items": 2,
  "high_priority_items": 5,
  "medium_priority_items": 7,
  "low_priority_items": 3
}
```

### Alerts & Notifications

#### Get Alerts
```http
GET /users/{user_id}/alerts?unread_only=true

Response:
{
  "user_id": "user123",
  "alerts": [
    {
      "alert_id": "uuid",
      "product_id": "OLJCESPC7Z",
      "product_name": "Sunglasses",
      "alert_type": "price_drop",
      "price_update": {
        "old_price": 19.99,
        "new_price": 15.99,
        "price_change": -4.00,
        "percent_change": -20.01
      },
      "target_price": 15.00,
      "is_read": false,
      "created_at": "2024-11-09T10:30:00Z"
    }
  ],
  "total": 1
}
```

#### Mark Alert as Read
```http
POST /users/{user_id}/alerts/{alert_id}/read
```

### Sharing

#### Share Wishlist
```http
POST /users/{user_id}/wishlist/share
Content-Type: application/json

{
  "share_with_user_id": "user456"
}
```

#### Unshare Wishlist
```http
DELETE /users/{user_id}/wishlist/share/{unshare_user_id}
```

#### Set Public/Private
```http
PUT /users/{user_id}/wishlist/public
Content-Type: application/json

{
  "is_public": true
}
```

### System Endpoints

#### Update Product Price (Admin/System)
```http
PUT /products/{product_id}/price
Content-Type: application/json

{
  "new_price": 15.99,
  "in_stock": true
}

Response:
{
  "success": true,
  "alerts_created": 3,
  "alerts": [...]
}
```

### Health Check
```http
GET /health
```

## Data Models

### WishlistItem
```go
{
  "item_id": "uuid",
  "user_id": "string",
  "product_id": "string",
  "product_name": "string",
  "current_price": 19.99,
  "original_price": 24.99,
  "target_price": 15.00,
  "priority": "high|medium|low",
  "notes": "string",
  "notify_on_price_drop": true,
  "notify_on_restock": false,
  "in_stock": true,
  "added_at": "timestamp",
  "updated_at": "timestamp"
}
```

### Wishlist
```go
{
  "user_id": "string",
  "items": [WishlistItem],
  "total_items": 15,
  "shared_with": ["user456", "user789"],
  "is_public": false,
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

### PriceAlert
```go
{
  "alert_id": "uuid",
  "user_id": "string",
  "item_id": "uuid",
  "product_id": "string",
  "product_name": "string",
  "alert_type": "price_drop|target_reached|restock",
  "price_update": {
    "old_price": 24.99,
    "new_price": 19.99,
    "price_change": -5.00,
    "percent_change": -20.01
  },
  "target_price": 15.00,
  "is_read": false,
  "created_at": "timestamp"
}
```

## Running the Service

### Locally
```bash
cd src/wishlistservice
go run .
```

### With Docker
```bash
docker build -t wishlistservice .
docker run -p 8098:8098 wishlistservice
```

### Environment Variables
- `PORT`: Service port (default: 8098)

## Testing

Run unit tests:
```bash
go test -v
```

Run tests with coverage:
```bash
go test -v -cover
```

## Business Rules

1. **Item Management**:
   - Each product can only be added once per user
   - Removing an item cleans up all associated alerts
   - Original price is set when item is first added

2. **Price Alerts**:
   - Price drop alerts trigger when notify_on_price_drop is true AND price decreases
   - Target price alerts trigger when price drops to or below target_price
   - Restock alerts trigger when notify_on_restock is true AND item goes from out-of-stock to in-stock

3. **Priority Levels**:
   - Valid values: "high", "medium", "low"
   - Default: "medium" if not specified
   - Used for sorting and organization

4. **Sharing**:
   - Wishlist can be shared with specific users
   - Can be set to public (viewable by anyone)
   - Share list is independent of public status

5. **Statistics**:
   - Items on sale: current_price < original_price
   - Average price drop: calculated only for items currently on sale
   - Total value: sum of all current prices

## Integration with Other Services

This service integrates with:
- **Product Catalog**: References product IDs
- **User Service**: References user IDs for wishlists and sharing
- **Inventory Service**: Could pull real-time stock status
- **API Gateway**: Routes requests to this service
- **Frontend**: Displays wishlists and notifications

## Alert Types

1. **price_drop**: Price decreased from previous value
2. **target_reached**: Price dropped to or below user's target price
3. **restock**: Item is back in stock

## Sorting Options

- `price_asc`: Lowest price first
- `price_desc`: Highest price first
- `priority`: High → Medium → Low
- `discount`: Largest discount first
- `recent`: Most recently added first
- `oldest`: Oldest items first

## Use Cases

### Use Case 1: Price Monitoring
```
1. User adds expensive item to wishlist
2. Sets target price to $50
3. Enables price drop notifications
4. System monitors price changes
5. When price drops to $50, user receives alert
```

### Use Case 2: Wishlist Sharing
```
1. User creates wishlist for birthday
2. Shares with family members
3. Family can view items
4. User receives what they want
```

### Use Case 3: Sale Tracking
```
1. User adds multiple items during browsing
2. Sets all to high priority
3. Checks stats regularly for sale items
4. Buys items when good discounts appear
```

## Future Enhancements

- [ ] Persistent storage (PostgreSQL/MongoDB)
- [ ] Email/SMS notifications for alerts
- [ ] Collaborative wishlists (multiple owners)
- [ ] Price prediction using ML
- [ ] Browser extension for easy adding
- [ ] Gift registry features
- [ ] Auto-purchase at target price
- [ ] Integration with external price comparison sites
- [ ] Wishlist analytics dashboard
- [ ] Social features (popular items, trending)

## Architecture

The service is built with:
- **Language**: Go 1.21
- **HTTP Framework**: Gorilla Mux
- **Storage**: In-memory (easily replaceable with DB)
- **API Style**: REST
- **Port**: 8098

## Performance Considerations

- In-memory storage for demo; use Redis/DB for production
- Price updates trigger alert generation in real-time
- Statistics calculated on-demand (could be cached)
- Thread-safe operations using sync.RWMutex

## License

Apache License 2.0
