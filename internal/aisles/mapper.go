package aisles

import (
	"encoding/json"
	"os"
	"strings"
)

// AisleMap maps lowercase ingredient name patterns to aisle labels.
// Exact matches take priority over partial (substring) matches.
// Among partial matches, the longest matching key wins.
type AisleMap map[string]string

// Load reads an AisleMap from a JSON file.
func Load(path string) (AisleMap, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m AisleMap
	return m, json.Unmarshal(data, &m)
}

// Save writes the AisleMap to a JSON file with pretty-printing.
func (m AisleMap) Save(path string) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Lookup finds the aisle for a given ingredient name.
// Priority: exact match > longest partial match (substring).
// Longest-match prevents "tuna" from beating "tuna steak" for input "tuna steak strips".
// Returns ("", false) if no match is found.
func (m AisleMap) Lookup(ingredient string) (string, bool) {
	lower := strings.ToLower(strings.TrimSpace(ingredient))

	// Exact match
	if aisle, ok := m[lower]; ok {
		return aisle, true
	}

	// Longest partial match — pick the longest key that is a substring of the input
	bestAisle := ""
	bestLen := 0
	for pattern, aisle := range m {
		if strings.Contains(lower, pattern) && len(pattern) > bestLen {
			bestLen = len(pattern)
			bestAisle = aisle
		}
	}
	if bestLen > 0 {
		return bestAisle, true
	}

	return "", false
}
