// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

var (
	logger *log.Logger
)

func init() {
	logger = log.New(os.Stdout, "[GAMIFICATION] ", log.LstdFlags)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8094"
	}
	logger.Printf("Starting Gamification Service on port %s", port)

	router := mux.NewRouter()

	// Health check
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")
	router.HandleFunc("/", rootHandler).Methods("GET")

	// User points and level
	router.HandleFunc("/users/{user_id}/points", getUserPointsHandler).Methods("GET")
	router.HandleFunc("/users/{user_id}/points", awardPointsHandler).Methods("POST")
	router.HandleFunc("/users/{user_id}/level", getUserLevelHandler).Methods("GET")

	// Badges
	router.HandleFunc("/users/{user_id}/badges", getUserBadgesHandler).Methods("GET")
	router.HandleFunc("/users/{user_id}/badges", awardBadgeHandler).Methods("POST")

	// Missions
	router.HandleFunc("/users/{user_id}/missions", getDailyMissionsHandler).Methods("GET")
	router.HandleFunc("/users/{user_id}/missions/{mission_id}/complete", completeMissionHandler).Methods("POST")

	// Leaderboard
	router.HandleFunc("/leaderboard/{type}", getLeaderboardHandler).Methods("GET")

	// Spin wheel
	router.HandleFunc("/users/{user_id}/spin", spinWheelHandler).Methods("POST")

	// Stats
	router.HandleFunc("/stats", getStatsHandler).Methods("GET")

	logger.Fatal(http.ListenAndServe(":"+port, router))
}

// Handlers

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"service":   "gamification-service",
		"timestamp": time.Now().UTC(),
	})
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"service": "Gamification Service",
		"version": "1.0.0",
		"endpoints": []string{
			"/health",
			"/users/{user_id}/points",
			"/users/{user_id}/badges",
			"/users/{user_id}/missions",
			"/leaderboard/{type}",
		},
	})
}

func getUserPointsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	progress := GetUserProgress(userID)

	respondJSON(w, http.StatusOK, progress)
}

func awardPointsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	var req struct {
		Points int    `json:"points"`
		Action string `json:"action"`
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	reward := AwardPoints(userID, req.Action, req.Points, req.Reason)

	respondJSON(w, http.StatusOK, reward)
}

func getUserLevelHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	progress := GetUserProgress(userID)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user_id": userID,
		"level":   progress.Level,
		"xp":      progress.XP,
		"next_level_xp": calculateNextLevelXP(progress.Level),
	})
}

func getUserBadgesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	badges := GetUserBadges(userID)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user_id": userID,
		"badges":  badges,
		"total":   len(badges),
	})
}

func awardBadgeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	var req struct {
		BadgeID string `json:"badge_id"`
		Reason  string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	badge := AwardBadge(userID, req.BadgeID, req.Reason)

	respondJSON(w, http.StatusOK, badge)
}

func getDailyMissionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	missions := GetDailyMissions(userID)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":  userID,
		"missions": missions,
		"date":     time.Now().Format("2006-01-02"),
	})
}

func completeMissionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	missionID := vars["mission_id"]

	reward := CompleteMission(userID, missionID)

	if reward == nil {
		respondError(w, http.StatusNotFound, "Mission not found or already completed")
		return
	}

	respondJSON(w, http.StatusOK, reward)
}

func getLeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	leaderboardType := vars["type"]

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "weekly"
	}

	leaderboard := GetLeaderboard(leaderboardType, period)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"type":        leaderboardType,
		"period":      period,
		"leaderboard": leaderboard,
		"updated_at":  time.Now().UTC(),
	})
}

func spinWheelHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	reward, err := SpinWheel(userID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, reward)
}

func getStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats := GetServiceStats()
	respondJSON(w, http.StatusOK, stats)
}

// Helper functions

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

func respondError(w http.ResponseWriter, code int, message string) {
	respondJSON(w, code, map[string]string{"error": message})
}

func calculateNextLevelXP(currentLevel int) int {
	// Formula: (level * 100) + (level * 50)
	return (currentLevel * 100) + (currentLevel * 50)
}
