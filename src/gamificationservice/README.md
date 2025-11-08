# Gamification Service

User engagement through points, badges, missions, and leaderboards.

## Features

- 🎯 **Points & Levels**: Earn XP, level up, get multipliers
- 🏆 **Badges**: Achievement system with multiple rarities
- 📋 **Daily Missions**: Complete tasks for rewards
- 🎰 **Lucky Wheel**: Spin for random rewards
- 📊 **Leaderboards**: Compete with other users
- 🔥 **Streak System**: Login daily for bonuses

## API Endpoints

### User Progress

#### Get User Points
```bash
GET /users/{user_id}/points
```

Response:
```json
{
  "user_id": "user123",
  "level": 5,
  "xp": 250,
  "points": 1500,
  "streak": 7,
  "last_login": "2024-01-15T10:30:00Z"
}
```

#### Award Points
```bash
POST /users/{user_id}/points
```

Request:
```json
{
  "points": 100,
  "action": "purchase",
  "reason": "Completed purchase"
}
```

Response:
```json
{
  "user_id": "user123",
  "points": 100,
  "multiplier": 1.35,
  "total_points": 135,
  "reason": "Completed purchase",
  "leveled_up": true,
  "new_level": 6,
  "new_badges": [
    {
      "badge_id": "level_10",
      "name": "Rising Star",
      "description": "Reached level 10",
      "rarity": "rare"
    }
  ]
}
```

### Badges

#### Get User Badges
```bash
GET /users/{user_id}/badges
```

#### Award Badge
```bash
POST /users/{user_id}/badges
```

Request:
```json
{
  "badge_id": "first_purchase",
  "reason": "Made first purchase"
}
```

### Daily Missions

#### Get Daily Missions
```bash
GET /users/{user_id}/missions
```

Response:
```json
{
  "user_id": "user123",
  "missions": [
    {
      "mission_id": "daily_purchase",
      "title": "Make a Purchase",
      "description": "Buy any product today",
      "reward": 100,
      "progress": 0,
      "target": 1,
      "completed": false
    }
  ],
  "date": "2024-01-15"
}
```

#### Complete Mission
```bash
POST /users/{user_id}/missions/{mission_id}/complete
```

### Leaderboard

```bash
GET /leaderboard/{type}?period=weekly
```

Types: `points`, `purchases`, `reviews`
Periods: `daily`, `weekly`, `monthly`, `all_time`

Response:
```json
{
  "type": "points",
  "period": "weekly",
  "leaderboard": [
    {
      "user_id": "user1",
      "username": "Alice",
      "score": 5000,
      "rank": 1
    }
  ]
}
```

### Lucky Wheel

```bash
POST /users/{user_id}/spin
```

Cost: 100 points

Response:
```json
{
  "type": "points",
  "value": 50,
  "description": "50 Points"
}
```

Possible rewards:
- 10 Points (40%)
- 50 Points (30%)
- 100 Points (20%)
- 10% Discount (7%)
- 20% Discount (2%)
- Free Shipping (1%)

## Integration Examples

### Award Points on Purchase

```javascript
// When user completes a purchase
await fetch(`http://localhost:8091/users/${userId}/points`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    points: calculatePurchasePoints(orderTotal),
    action: 'purchase',
    reason: `Purchase #${orderId}`
  })
});
```

### Show User Progress

```javascript
const response = await fetch(`http://localhost:8091/users/${userId}/points`);
const progress = await response.json();

document.getElementById('user-level').textContent = `Level ${progress.level}`;
document.getElementById('user-points').textContent = progress.points;
document.getElementById('user-streak').textContent = `${progress.streak} day streak!`;
```

### Display Daily Missions

```javascript
const response = await fetch(`http://localhost:8091/users/${userId}/missions`);
const data = await response.json();

const missionsHTML = data.missions.map(mission => `
  <div class="mission ${mission.completed ? 'completed' : ''}">
    <h4>${mission.title}</h4>
    <p>${mission.description}</p>
    <div class="progress">
      <div class="bar" style="width: ${(mission.progress / mission.target) * 100}%"></div>
    </div>
    <span class="reward">+${mission.reward} points</span>
  </div>
`).join('');
```

## Point Calculation

### Base Points

| Action | Points |
|--------|--------|
| Purchase | $1 = 10 points |
| Review | 50 points |
| Share product | 10 points |
| Daily login | 20 points |
| Complete mission | Varies |

### Multipliers

**Level Bonus:** +5% per level
- Level 1: 1.05x
- Level 5: 1.25x
- Level 10: 1.50x

**Streak Bonus:** +2% per day (max 50%)
- 7 days: +14%
- 14 days: +28%
- 25 days: +50% (max)

**Example:**
```
Base: 100 points
Level 5 bonus: +25%
Streak 7 days: +14%
Total: 100 × 1.39 = 139 points
```

## Level System

### XP Requirements

```
Level 1 → 2: 150 XP
Level 2 → 3: 300 XP
Level 3 → 4: 450 XP
...
Level N → N+1: N × 150 XP
```

### Level Rewards

| Level | Reward |
|-------|--------|
| 5 | +5% discount on all purchases |
| 10 | "Rising Star" badge |
| 15 | Free shipping for 1 month |
| 20 | VIP customer support |
| 25 | "Legendary Shopper" badge |

## Badge System

### Badge Rarities

- **Common** (white): Easy to obtain
- **Rare** (blue): Moderate challenge
- **Epic** (purple): Significant achievement
- **Legendary** (gold): Extraordinary accomplishment

### Available Badges

| Badge ID | Name | Description | Rarity |
|----------|------|-------------|--------|
| first_purchase | First Purchase | Made your first purchase | Common |
| level_10 | Rising Star | Reached level 10 | Rare |
| streak_7 | Week Warrior | 7 day login streak | Rare |
| reviewer | Reviewer | Written 10 reviews | Common |
| big_spender | Big Spender | Spent over $1000 | Epic |
| early_bird | Early Bird | Make purchase before 9 AM | Common |
| night_owl | Night Owl | Make purchase after 11 PM | Common |

## Running the Service

### Local Development

```bash
# Install dependencies
go mod download

# Run service
go run *.go
```

### With Docker

```bash
# Build
docker build -t gamificationservice .

# Run
docker run -p 8091:8091 gamificationservice
```

### Environment Variables

```bash
PORT=8091                  # Service port
REDIS_ADDR=localhost:6379  # Redis for persistence (optional)
```

## Future Enhancements

- [ ] Redis integration for persistence
- [ ] PostgreSQL for leaderboards
- [ ] Real-time notifications (WebSocket)
- [ ] Social features (friend challenges)
- [ ] Seasonal events
- [ ] Achievement progression tracking
- [ ] Point expiration
- [ ] Badge showcase on profile

## Testing

```bash
# Health check
curl http://localhost:8091/health

# Get user progress
curl http://localhost:8091/users/test-user/points

# Award points
curl -X POST http://localhost:8091/users/test-user/points \
  -H "Content-Type: application/json" \
  -d '{"points": 100, "action": "purchase", "reason": "Test"}'

# Get missions
curl http://localhost:8091/users/test-user/missions

# Spin wheel
curl -X POST http://localhost:8091/users/test-user/spin

# Get leaderboard
curl http://localhost:8091/leaderboard/points?period=weekly
```
