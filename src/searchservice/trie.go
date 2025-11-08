package main

import (
	"sort"
	"strings"
	"sync"
)

type TrieNode struct {
	children map[rune]*TrieNode
	isEnd    bool
	word     string
	score    int
	category string
	count    int // Search frequency
}

type Trie struct {
	root *TrieNode
	mu   sync.RWMutex
	size int
}

func NewTrie() *Trie {
	return &Trie{
		root: &TrieNode{
			children: make(map[rune]*TrieNode),
		},
	}
}

func (t *Trie) Insert(word string, score int, category string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	word = strings.ToLower(strings.TrimSpace(word))
	if word == "" {
		return
	}

	node := t.root

	for _, char := range word {
		if _, exists := node.children[char]; !exists {
			node.children[char] = &TrieNode{
				children: make(map[rune]*TrieNode),
			}
		}
		node = node.children[char]
	}

	if !node.isEnd {
		t.size++
	}

	node.isEnd = true
	node.word = word
	node.score = score
	node.category = category
	node.count++
}

func (t *Trie) Search(word string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	word = strings.ToLower(strings.TrimSpace(word))
	node := t.root

	for _, char := range word {
		if child, exists := node.children[char]; exists {
			node = child
		} else {
			return false
		}
	}

	return node.isEnd
}

func (t *Trie) Autocomplete(prefix string, limit int) []Suggestion {
	t.mu.RLock()
	defer t.mu.RUnlock()

	prefix = strings.ToLower(strings.TrimSpace(prefix))
	if prefix == "" {
		return []Suggestion{}
	}

	// Find the node for the prefix
	node := t.root
	for _, char := range prefix {
		if child, exists := node.children[char]; exists {
			node = child
		} else {
			return []Suggestion{} // Prefix not found
		}
	}

	// Collect all words with this prefix
	suggestions := []Suggestion{}
	t.collectWords(node, &suggestions)

	// Sort by score (descending) and count (descending)
	sort.Slice(suggestions, func(i, j int) bool {
		if suggestions[i].Score == suggestions[j].Score {
			return suggestions[i].Popularity > suggestions[j].Popularity
		}
		return suggestions[i].Score > suggestions[j].Score
	})

	// Limit results
	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	// Mark exact matches
	for i := range suggestions {
		suggestions[i].IsExact = strings.HasPrefix(suggestions[i].Text, prefix)
	}

	return suggestions
}

func (t *Trie) collectWords(node *TrieNode, suggestions *[]Suggestion) {
	if node.isEnd {
		*suggestions = append(*suggestions, Suggestion{
			Text:       node.word,
			Score:      node.score,
			Category:   node.category,
			Popularity: node.count,
			IsExact:    true,
		})
	}

	for _, child := range node.children {
		t.collectWords(child, suggestions)
	}
}

func (t *Trie) Size() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.size
}

func (t *Trie) Delete(word string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	word = strings.ToLower(strings.TrimSpace(word))
	if word == "" {
		return false
	}

	node := t.root
	path := []*TrieNode{node}

	// Traverse to the word
	for _, char := range word {
		if child, exists := node.children[char]; exists {
			node = child
			path = append(path, node)
		} else {
			return false // Word not found
		}
	}

	if !node.isEnd {
		return false // Word doesn't exist
	}

	node.isEnd = false
	t.size--

	// Clean up empty nodes
	for i := len(path) - 1; i > 0; i-- {
		current := path[i]
		if !current.isEnd && len(current.children) == 0 {
			parent := path[i-1]
			for char, child := range parent.children {
				if child == current {
					delete(parent.children, char)
					break
				}
			}
		} else {
			break
		}
	}

	return true
}

func (t *Trie) GetAllWords() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	words := []string{}
	var collect func(*TrieNode)

	collect = func(node *TrieNode) {
		if node.isEnd {
			words = append(words, node.word)
		}
		for _, child := range node.children {
			collect(child)
		}
	}

	collect(t.root)
	return words
}

func (t *Trie) IncrementCount(word string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	word = strings.ToLower(strings.TrimSpace(word))
	node := t.root

	for _, char := range word {
		if child, exists := node.children[char]; exists {
			node = child
		} else {
			return
		}
	}

	if node.isEnd {
		node.count++
	}
}
