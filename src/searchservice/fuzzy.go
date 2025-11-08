package main

import (
	"sort"
	"strings"
)

type FuzzyMatcher struct {
	dictionary []string
}

func NewFuzzyMatcher() *FuzzyMatcher {
	return &FuzzyMatcher{
		dictionary: []string{
			"sunglasses", "tank top", "t-shirt", "sneakers",
			"backpack", "jeans", "jacket", "watch",
			"hat", "shorts", "socks", "belt",
			"shoes", "sandals", "dress", "skirt",
			"sweater", "hoodie", "boots", "scarf",
		},
	}
}

func (fm *FuzzyMatcher) FindSimilar(query string, limit int) []Suggestion {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return []Suggestion{}
	}

	type scored struct {
		word     string
		distance int
		score    int
	}

	results := []scored{}

	for _, word := range fm.dictionary {
		distance := levenshteinDistance(query, word)

		// Only include if distance is reasonable
		if distance <= 3 {
			// Calculate score based on distance (closer = higher score)
			score := 100 - (distance * 20)
			if score < 0 {
				score = 0
			}

			results = append(results, scored{
				word:     word,
				distance: distance,
				score:    score,
			})
		}
	}

	// Sort by distance (ascending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].distance < results[j].distance
	})

	// Convert to Suggestions
	suggestions := []Suggestion{}
	for i, r := range results {
		if i >= limit {
			break
		}

		suggestions = append(suggestions, Suggestion{
			Text:       r.word,
			Score:      r.score,
			Category:   "fuzzy",
			Popularity: 0,
			IsExact:    false,
		})
	}

	return suggestions
}

// Levenshtein distance algorithm
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	// Initialize first column and row
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min3(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func (fm *FuzzyMatcher) AddWord(word string) {
	fm.dictionary = append(fm.dictionary, strings.ToLower(word))
}

func (fm *FuzzyMatcher) SetDictionary(words []string) {
	fm.dictionary = words
	for i := range fm.dictionary {
		fm.dictionary[i] = strings.ToLower(fm.dictionary[i])
	}
}
