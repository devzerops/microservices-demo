package main

import (
	"testing"
	"time"
)

func TestGetUserProgress(t *testing.T) {
	// Clear map for testing
	mu.Lock()
	userProgressMap = make(map[string]*UserProgress)
	mu.Unlock()

	t.Run("creates new user progress", func(t *testing.T) {
		userID := "test_user_1"
		progress := GetUserProgress(userID)

		if progress == nil {
			t.Fatal("Expected progress to be created")
		}

		if progress.UserID != userID {
			t.Errorf("Expected UserID %s, got %s", userID, progress.UserID)
		}

		if progress.Level != 1 {
			t.Errorf("Expected initial level 1, got %d", progress.Level)
		}

		if progress.XP != 0 {
			t.Errorf("Expected initial XP 0, got %d", progress.XP)
		}

		if progress.Points != 0 {
			t.Errorf("Expected initial Points 0, got %d", progress.Points)
		}
	})

	t.Run("retrieves existing user progress", func(t *testing.T) {
		userID := "test_user_2"

		// Create first time
		progress1 := GetUserProgress(userID)
		progress1.Level = 5
		progress1.Points = 1000

		// Retrieve second time
		progress2 := GetUserProgress(userID)

		if progress2.Level != 5 {
			t.Errorf("Expected level 5, got %d", progress2.Level)
		}

		if progress2.Points != 1000 {
			t.Errorf("Expected points 1000, got %d", progress2.Points)
		}
	})
}

func TestAwardPoints(t *testing.T) {
	// Clear map for testing
	mu.Lock()
	userProgressMap = make(map[string]*UserProgress)
	userBadgesMap = make(map[string][]Badge)
	mu.Unlock()

	t.Run("awards basic points", func(t *testing.T) {
		userID := "test_award_1"
		reward := AwardPoints(userID, "test_action", 100, "Test reward")

		if reward == nil {
			t.Fatal("Expected reward to be returned")
		}

		if reward.Points != 100 {
			t.Errorf("Expected 100 points, got %d", reward.Points)
		}

		if reward.Multiplier < 1.0 {
			t.Errorf("Expected multiplier >= 1.0, got %f", reward.Multiplier)
		}

		progress := GetUserProgress(userID)
		if progress.Points == 0 {
			t.Error("Expected points to be awarded")
		}
	})

	t.Run("applies multiplier correctly", func(t *testing.T) {
		userID := "test_award_2"

		// Award points
		reward := AwardPoints(userID, "test", 100, "Test")

		// Check that total points includes multiplier
		if reward.TotalPoints < reward.Points {
			t.Errorf("Expected TotalPoints >= Points, got %d >= %d", reward.TotalPoints, reward.Points)
		}
	})
}

func TestCalculateMultiplier(t *testing.T) {
	tests := []struct {
		name     string
		level    int
		streak   int
		expected float64
	}{
		{"level 1, no streak", 1, 0, 1.05},
		{"level 5, no streak", 5, 0, 1.25},
		{"level 1, streak 5", 1, 5, 1.15},
		{"level 10, streak 10", 10, 10, 1.70},
		{"high streak capped", 1, 100, 1.55}, // streak bonus capped at 0.5
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progress := &UserProgress{
				Level:  tt.level,
				Streak: tt.streak,
			}

			result := calculateMultiplier(progress)

			if result != tt.expected {
				t.Errorf("Expected multiplier %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestGetUserBadges(t *testing.T) {
	// Clear map for testing
	mu.Lock()
	userBadgesMap = make(map[string][]Badge)
	mu.Unlock()

	t.Run("returns empty array for new user", func(t *testing.T) {
		userID := "test_badges_1"
		badges := GetUserBadges(userID)

		if badges == nil {
			t.Fatal("Expected badges array to be returned")
		}

		if len(badges) != 0 {
			t.Errorf("Expected 0 badges, got %d", len(badges))
		}
	})

	t.Run("returns user badges", func(t *testing.T) {
		userID := "test_badges_2"

		// Award a badge
		AwardBadge(userID, "first_purchase", "Test")

		badges := GetUserBadges(userID)

		if len(badges) != 1 {
			t.Errorf("Expected 1 badge, got %d", len(badges))
		}

		if badges[0].BadgeID != "first_purchase" {
			t.Errorf("Expected badge ID 'first_purchase', got '%s'", badges[0].BadgeID)
		}
	})
}

func TestAwardBadge(t *testing.T) {
	// Clear map for testing
	mu.Lock()
	userBadgesMap = make(map[string][]Badge)
	mu.Unlock()

	t.Run("awards new badge", func(t *testing.T) {
		userID := "test_award_badge_1"
		badge := AwardBadge(userID, "first_purchase", "Test")

		if badge == nil {
			t.Fatal("Expected badge to be returned")
		}

		if badge.BadgeID != "first_purchase" {
			t.Errorf("Expected badge ID 'first_purchase', got '%s'", badge.BadgeID)
		}

		if badge.Name != "First Purchase" {
			t.Errorf("Expected badge name 'First Purchase', got '%s'", badge.Name)
		}
	})

	t.Run("does not award duplicate badge", func(t *testing.T) {
		userID := "test_award_badge_2"

		// Award first time
		badge1 := AwardBadge(userID, "first_purchase", "Test")

		// Award second time
		badge2 := AwardBadge(userID, "first_purchase", "Test")

		// Should return existing badge
		if badge1.AwardedAt != badge2.AwardedAt {
			badges := GetUserBadges(userID)
			if len(badges) != 1 {
				t.Errorf("Expected 1 badge after duplicate award, got %d", len(badges))
			}
		}
	})
}

func TestGetDailyMissions(t *testing.T) {
	// Clear map for testing
	mu.Lock()
	userMissionsMap = make(map[string]map[string]*Mission)
	mu.Unlock()

	t.Run("generates daily missions", func(t *testing.T) {
		userID := "test_missions_1"
		missions := GetDailyMissions(userID)

		if len(missions) == 0 {
			t.Error("Expected daily missions to be generated")
		}

		// Check mission structure
		for _, mission := range missions {
			if mission.MissionID == "" {
				t.Error("Expected mission to have ID")
			}
			if mission.Title == "" {
				t.Error("Expected mission to have title")
			}
			if mission.Reward == 0 {
				t.Error("Expected mission to have reward")
			}
			if mission.Completed {
				t.Error("Expected new mission to not be completed")
			}
		}
	})
}

func TestCompleteMission(t *testing.T) {
	// Clear map for testing
	mu.Lock()
	userMissionsMap = make(map[string]map[string]*Mission)
	userProgressMap = make(map[string]*UserProgress)
	mu.Unlock()

	t.Run("completes mission and awards points", func(t *testing.T) {
		userID := "test_complete_1"

		// Get missions
		missions := GetDailyMissions(userID)
		if len(missions) == 0 {
			t.Fatal("Expected missions to be generated")
		}

		missionID := missions[0].MissionID

		// Complete mission
		reward := CompleteMission(userID, missionID)

		if reward == nil {
			t.Fatal("Expected reward to be returned")
		}

		// Check mission is marked complete
		updatedMissions := GetDailyMissions(userID)
		for _, m := range updatedMissions {
			if m.MissionID == missionID {
				if !m.Completed {
					t.Error("Expected mission to be marked completed")
				}
				if m.Progress != m.Target {
					t.Errorf("Expected progress %d to equal target %d", m.Progress, m.Target)
				}
			}
		}
	})

	t.Run("returns nil for non-existent mission", func(t *testing.T) {
		userID := "test_complete_2"
		reward := CompleteMission(userID, "non_existent_mission")

		if reward != nil {
			t.Error("Expected nil for non-existent mission")
		}
	})
}

func TestGetLeaderboard(t *testing.T) {
	t.Run("returns leaderboard entries", func(t *testing.T) {
		leaderboard := GetLeaderboard("points", "weekly")

		if len(leaderboard) == 0 {
			t.Error("Expected leaderboard to have entries")
		}

		// Check leaderboard structure
		for i, entry := range leaderboard {
			if entry.UserID == "" {
				t.Error("Expected entry to have user ID")
			}
			if entry.Rank != i+1 {
				t.Errorf("Expected rank %d, got %d", i+1, entry.Rank)
			}
			if entry.Score <= 0 {
				t.Error("Expected positive score")
			}
		}
	})
}

func TestSpinWheel(t *testing.T) {
	// Clear map for testing
	mu.Lock()
	userProgressMap = make(map[string]*UserProgress)
	mu.Unlock()

	t.Run("spins wheel with sufficient points", func(t *testing.T) {
		userID := "test_spin_1"

		// Give user enough points
		progress := GetUserProgress(userID)
		progress.Points = 200

		reward, err := SpinWheel(userID)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if reward == nil {
			t.Fatal("Expected reward to be returned")
		}

		if reward.Type == "" {
			t.Error("Expected reward to have type")
		}

		// Check points were deducted
		updatedProgress := GetUserProgress(userID)
		if updatedProgress.Points >= 200 {
			t.Error("Expected points to be deducted")
		}
	})

	t.Run("fails with insufficient points", func(t *testing.T) {
		userID := "test_spin_2"

		// Give user insufficient points
		progress := GetUserProgress(userID)
		progress.Points = 50

		reward, err := SpinWheel(userID)

		if err == nil {
			t.Error("Expected error for insufficient points")
		}

		if reward != nil {
			t.Error("Expected nil reward for insufficient points")
		}

		if err.Error() != "insufficient points" {
			t.Errorf("Expected 'insufficient points' error, got '%s'", err.Error())
		}
	})
}

func TestGetServiceStats(t *testing.T) {
	// Clear map for testing
	mu.Lock()
	userProgressMap = make(map[string]*UserProgress)
	userBadgesMap = make(map[string][]Badge)
	userMissionsMap = make(map[string]map[string]*Mission)
	mu.Unlock()

	t.Run("returns service statistics", func(t *testing.T) {
		// Create some data
		GetUserProgress("user1")
		GetUserProgress("user2")
		AwardBadge("user1", "first_purchase", "Test")

		stats := GetServiceStats()

		if stats["total_users"] != 2 {
			t.Errorf("Expected 2 total users, got %v", stats["total_users"])
		}

		if stats["total_badges_awarded"] != 1 {
			t.Errorf("Expected 1 badge awarded, got %v", stats["total_badges_awarded"])
		}
	})
}

func TestUpdateStreak(t *testing.T) {
	t.Run("increments streak for consecutive days", func(t *testing.T) {
		progress := &UserProgress{
			Streak:    5,
			LastLogin: time.Now().AddDate(0, 0, -1), // Yesterday
		}

		updateStreak(progress)

		if progress.Streak != 6 {
			t.Errorf("Expected streak 6, got %d", progress.Streak)
		}
	})

	t.Run("resets streak for gap > 1 day", func(t *testing.T) {
		progress := &UserProgress{
			Streak:    5,
			LastLogin: time.Now().AddDate(0, 0, -3), // 3 days ago
		}

		updateStreak(progress)

		if progress.Streak != 1 {
			t.Errorf("Expected streak reset to 1, got %d", progress.Streak)
		}
	})

	t.Run("maintains streak for same day login", func(t *testing.T) {
		progress := &UserProgress{
			Streak:    5,
			LastLogin: time.Now(), // Today
		}

		updateStreak(progress)

		// Streak should remain the same
		if progress.Streak != 5 {
			t.Errorf("Expected streak 5, got %d", progress.Streak)
		}
	})
}
