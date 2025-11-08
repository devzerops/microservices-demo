package main

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

// Data structures

type UserProgress struct {
	UserID    string    `json:"user_id"`
	Level     int       `json:"level"`
	XP        int       `json:"xp"`
	Points    int       `json:"points"`
	Streak    int       `json:"streak"`
	LastLogin time.Time `json:"last_login"`
}

type Badge struct {
	BadgeID     string    `json:"badge_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IconURL     string    `json:"icon_url"`
	AwardedAt   time.Time `json:"awarded_at"`
	Rarity      string    `json:"rarity"` // common, rare, epic, legendary
}

type Mission struct {
	MissionID   string `json:"mission_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Reward      int    `json:"reward"`
	Progress    int    `json:"progress"`
	Target      int    `json:"target"`
	Completed   bool   `json:"completed"`
}

type PointsReward struct {
	UserID       string   `json:"user_id"`
	Points       int      `json:"points"`
	Multiplier   float64  `json:"multiplier"`
	TotalPoints  int      `json:"total_points"`
	Reason       string   `json:"reason"`
	LeveledUp    bool     `json:"leveled_up"`
	NewLevel     int      `json:"new_level,omitempty"`
	NewBadges    []Badge  `json:"new_badges,omitempty"`
}

type WheelReward struct {
	Type        string `json:"type"`
	Value       interface{} `json:"value"`
	Description string `json:"description"`
}

type LeaderboardEntry struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Score    int    `json:"score"`
	Rank     int    `json:"rank"`
}

// In-memory storage (in production, use Redis or a database)
var (
	userProgressMap = make(map[string]*UserProgress)
	userBadgesMap   = make(map[string][]Badge)
	userMissionsMap = make(map[string]map[string]*Mission)
	leaderboardCache = make(map[string][]LeaderboardEntry)
	mu              sync.RWMutex
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GetUserProgress retrieves or creates user progress
func GetUserProgress(userID string) *UserProgress {
	mu.RLock()
	progress, exists := userProgressMap[userID]
	mu.RUnlock()

	if !exists {
		mu.Lock()
		progress = &UserProgress{
			UserID:    userID,
			Level:     1,
			XP:        0,
			Points:    0,
			Streak:    0,
			LastLogin: time.Now(),
		}
		userProgressMap[userID] = progress
		mu.Unlock()
	}

	// Update streak
	updateStreak(progress)

	return progress
}

// AwardPoints awards points to a user
func AwardPoints(userID, action string, basePoints int, reason string) *PointsReward {
	progress := GetUserProgress(userID)

	mu.Lock()
	defer mu.Unlock()

	// Calculate multiplier based on level and streak
	multiplier := calculateMultiplier(progress)

	totalPoints := int(float64(basePoints) * multiplier)

	// Add points
	progress.Points += totalPoints
	progress.XP += totalPoints

	// Check for level up
	leveledUp := false
	newLevel := progress.Level

	nextLevelXP := calculateNextLevelXP(progress.Level)
	if progress.XP >= nextLevelXP {
		progress.Level++
		progress.XP -= nextLevelXP
		leveledUp = true
		newLevel = progress.Level
	}

	// Check for new badges
	newBadges := checkForBadges(userID, action, progress)

	return &PointsReward{
		UserID:      userID,
		Points:      basePoints,
		Multiplier:  multiplier,
		TotalPoints: totalPoints,
		Reason:      reason,
		LeveledUp:   leveledUp,
		NewLevel:    newLevel,
		NewBadges:   newBadges,
	}
}

// GetUserBadges retrieves all badges for a user
func GetUserBadges(userID string) []Badge {
	mu.RLock()
	defer mu.RUnlock()

	badges, exists := userBadgesMap[userID]
	if !exists {
		return []Badge{}
	}

	return badges
}

// AwardBadge awards a badge to a user
func AwardBadge(userID, badgeID, reason string) *Badge {
	mu.Lock()
	defer mu.Unlock()

	// Check if user already has this badge
	if badges, exists := userBadgesMap[userID]; exists {
		for _, badge := range badges {
			if badge.BadgeID == badgeID {
				return &badge // Already has badge
			}
		}
	}

	// Get badge info
	badgeInfo := getBadgeInfo(badgeID)

	badge := Badge{
		BadgeID:     badgeID,
		Name:        badgeInfo.Name,
		Description: badgeInfo.Description,
		IconURL:     badgeInfo.IconURL,
		AwardedAt:   time.Now(),
		Rarity:      badgeInfo.Rarity,
	}

	if userBadgesMap[userID] == nil {
		userBadgesMap[userID] = []Badge{}
	}

	userBadgesMap[userID] = append(userBadgesMap[userID], badge)

	return &badge
}

// GetDailyMissions retrieves daily missions for a user
func GetDailyMissions(userID string) []Mission {
	mu.Lock()
	defer mu.Unlock()

	// Check if missions exist for today
	if userMissionsMap[userID] == nil {
		userMissionsMap[userID] = make(map[string]*Mission)
	}

	// Generate daily missions if not exists
	if len(userMissionsMap[userID]) == 0 {
		missions := generateDailyMissions(userID)
		for _, mission := range missions {
			userMissionsMap[userID][mission.MissionID] = &mission
		}
	}

	// Convert map to slice
	var missions []Mission
	for _, mission := range userMissionsMap[userID] {
		missions = append(missions, *mission)
	}

	return missions
}

// CompleteMission marks a mission as complete
func CompleteMission(userID, missionID string) *PointsReward {
	mu.Lock()
	defer mu.Unlock()

	if userMissionsMap[userID] == nil || userMissionsMap[userID][missionID] == nil {
		return nil
	}

	mission := userMissionsMap[userID][missionID]

	if mission.Completed {
		return nil
	}

	mission.Completed = true
	mission.Progress = mission.Target

	mu.Unlock()
	reward := AwardPoints(userID, "mission_complete", mission.Reward, "Completed mission: "+mission.Title)
	mu.Lock()

	return reward
}

// GetLeaderboard retrieves leaderboard
func GetLeaderboard(leaderboardType, period string) []LeaderboardEntry {
	mu.RLock()
	defer mu.RUnlock()

	key := leaderboardType + "_" + period

	// If cached, return cache
	if cached, exists := leaderboardCache[key]; exists {
		return cached
	}

	// Generate leaderboard
	leaderboard := generateLeaderboard(leaderboardType, period)

	mu.RUnlock()
	mu.Lock()
	leaderboardCache[key] = leaderboard
	mu.Unlock()
	mu.RLock()

	return leaderboard
}

// SpinWheel spins the lucky wheel
func SpinWheel(userID string) (*WheelReward, error) {
	const spinCost = 100

	progress := GetUserProgress(userID)

	mu.Lock()
	defer mu.Unlock()

	if progress.Points < spinCost {
		return nil, errors.New("insufficient points")
	}

	// Deduct spin cost
	progress.Points -= spinCost

	// Generate random reward
	reward := generateWheelReward()

	// Apply reward
	applyWheelReward(userID, reward)

	return reward, nil
}

// Helper functions

func updateStreak(progress *UserProgress) {
	now := time.Now()
	lastLogin := progress.LastLogin

	// If last login was yesterday, increment streak
	if now.Format("2006-01-02") != lastLogin.Format("2006-01-02") {
		daysDiff := int(now.Sub(lastLogin).Hours() / 24)
		if daysDiff == 1 {
			progress.Streak++
		} else if daysDiff > 1 {
			progress.Streak = 1
		}
		progress.LastLogin = now
	}
}

func calculateMultiplier(progress *UserProgress) float64 {
	multiplier := 1.0

	// Level bonus (5% per level)
	multiplier += float64(progress.Level) * 0.05

	// Streak bonus (2% per day, max 50%)
	streakBonus := float64(progress.Streak) * 0.02
	if streakBonus > 0.5 {
		streakBonus = 0.5
	}
	multiplier += streakBonus

	return multiplier
}

func checkForBadges(userID, action string, progress *UserProgress) []Badge {
	var newBadges []Badge

	// First purchase
	if action == "purchase" {
		badges := GetUserBadges(userID)
		hasFirstPurchase := false
		for _, badge := range badges {
			if badge.BadgeID == "first_purchase" {
				hasFirstPurchase = true
				break
			}
		}
		if !hasFirstPurchase && len(badges) == 0 {
			badge := AwardBadge(userID, "first_purchase", "Made first purchase")
			newBadges = append(newBadges, *badge)
		}
	}

	// Level milestones
	if progress.Level == 10 {
		badge := AwardBadge(userID, "level_10", "Reached level 10")
		newBadges = append(newBadges, *badge)
	}

	return newBadges
}

type BadgeInfo struct {
	Name        string
	Description string
	IconURL     string
	Rarity      string
}

func getBadgeInfo(badgeID string) BadgeInfo {
	badgeMap := map[string]BadgeInfo{
		"first_purchase": {
			Name:        "First Purchase",
			Description: "Made your first purchase",
			IconURL:     "/static/badges/first_purchase.png",
			Rarity:      "common",
		},
		"level_10": {
			Name:        "Rising Star",
			Description: "Reached level 10",
			IconURL:     "/static/badges/level_10.png",
			Rarity:      "rare",
		},
		"streak_7": {
			Name:        "Week Warrior",
			Description: "7 day login streak",
			IconURL:     "/static/badges/streak_7.png",
			Rarity:      "rare",
		},
		"reviewer": {
			Name:        "Reviewer",
			Description: "Written 10 reviews",
			IconURL:     "/static/badges/reviewer.png",
			Rarity:      "common",
		},
	}

	if info, exists := badgeMap[badgeID]; exists {
		return info
	}

	return BadgeInfo{
		Name:        "Unknown Badge",
		Description: "Badge description",
		IconURL:     "/static/badges/default.png",
		Rarity:      "common",
	}
}

func generateDailyMissions(userID string) []Mission {
	return []Mission{
		{
			MissionID:   "daily_purchase",
			Title:       "Make a Purchase",
			Description: "Buy any product today",
			Reward:      100,
			Progress:    0,
			Target:      1,
			Completed:   false,
		},
		{
			MissionID:   "daily_review",
			Title:       "Write a Review",
			Description: "Write a product review",
			Reward:      50,
			Progress:    0,
			Target:      1,
			Completed:   false,
		},
		{
			MissionID:   "daily_share",
			Title:       "Share Products",
			Description: "Share 3 products with friends",
			Reward:      30,
			Progress:    0,
			Target:      3,
			Completed:   false,
		},
	}
}

func generateLeaderboard(leaderboardType, period string) []LeaderboardEntry {
	// In production, query database for real data
	// For demo, return mock data

	return []LeaderboardEntry{
		{UserID: "user1", Username: "Alice", Score: 5000, Rank: 1},
		{UserID: "user2", Username: "Bob", Score: 4500, Rank: 2},
		{UserID: "user3", Username: "Charlie", Score: 4000, Rank: 3},
		{UserID: "user4", Username: "Diana", Score: 3500, Rank: 4},
		{UserID: "user5", Username: "Eve", Score: 3000, Rank: 5},
	}
}

func generateWheelReward() *WheelReward {
	// Weighted random selection
	roll := rand.Float64()

	switch {
	case roll < 0.4: // 40%
		return &WheelReward{
			Type:        "points",
			Value:       10,
			Description: "10 Points",
		}
	case roll < 0.7: // 30%
		return &WheelReward{
			Type:        "points",
			Value:       50,
			Description: "50 Points",
		}
	case roll < 0.9: // 20%
		return &WheelReward{
			Type:        "points",
			Value:       100,
			Description: "100 Points",
		}
	case roll < 0.97: // 7%
		return &WheelReward{
			Type:        "discount",
			Value:       "10%",
			Description: "10% Discount Coupon",
		}
	case roll < 0.99: // 2%
		return &WheelReward{
			Type:        "discount",
			Value:       "20%",
			Description: "20% Discount Coupon",
		}
	default: // 1%
		return &WheelReward{
			Type:        "free_shipping",
			Value:       nil,
			Description: "Free Shipping Voucher",
		}
	}
}

func applyWheelReward(userID string, reward *WheelReward) {
	progress := GetUserProgress(userID)

	if reward.Type == "points" {
		if points, ok := reward.Value.(int); ok {
			progress.Points += points
		}
	}

	// For discount and free_shipping, would save to user's coupons
	// (not implemented in this demo)
}

// GetServiceStats returns service statistics
func GetServiceStats() map[string]interface{} {
	mu.RLock()
	defer mu.RUnlock()

	totalUsers := len(userProgressMap)
	totalBadgesAwarded := 0
	for _, badges := range userBadgesMap {
		totalBadgesAwarded += len(badges)
	}

	return map[string]interface{}{
		"total_users":         totalUsers,
		"total_badges_awarded": totalBadgesAwarded,
		"active_missions":     len(userMissionsMap),
	}
}
