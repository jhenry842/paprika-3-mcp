package aisles_test

import (
	"path/filepath"
	"testing"

	"github.com/soggycactus/paprika-3-mcp/internal/aisles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLookupExactMatch(t *testing.T) {
	m := aisles.AisleMap{"chicken breast": "Meat & Seafood"}
	aisle, ok := m.Lookup("chicken breast")
	assert.True(t, ok)
	assert.Equal(t, "Meat & Seafood", aisle)
}

func TestLookupCaseInsensitive(t *testing.T) {
	m := aisles.AisleMap{"chicken breast": "Meat & Seafood"}
	aisle, ok := m.Lookup("Chicken Breast")
	assert.True(t, ok)
	assert.Equal(t, "Meat & Seafood", aisle)
}

func TestLookupPartialMatch(t *testing.T) {
	m := aisles.AisleMap{"chicken": "Meat & Seafood"}
	aisle, ok := m.Lookup("boneless chicken thighs")
	assert.True(t, ok)
	assert.Equal(t, "Meat & Seafood", aisle)
}

func TestLookupExactBeforePartial(t *testing.T) {
	m := aisles.AisleMap{
		"chicken":       "Meat & Seafood",
		"chicken broth": "Soups & Broths",
	}
	aisle, ok := m.Lookup("chicken broth")
	assert.True(t, ok)
	assert.Equal(t, "Soups & Broths", aisle)
}

func TestLookupLongestPartialMatchWins(t *testing.T) {
	// When no exact match, the longest matching key wins.
	// This prevents "tuna" matching before "tuna steak" for input "tuna steak strips".
	m := aisles.AisleMap{
		"tuna":       "Canned Goods",
		"tuna steak": "Meat & Seafood",
	}
	aisle, ok := m.Lookup("tuna steak strips")
	assert.True(t, ok)
	assert.Equal(t, "Meat & Seafood", aisle)
}

func TestLookupLongestPartialFrozen(t *testing.T) {
	// "frozen corn" should beat "corn" for input "frozen corn kernels"
	m := aisles.AisleMap{
		"corn":        "Produce",
		"frozen corn": "Frozen Foods",
	}
	aisle, ok := m.Lookup("frozen corn kernels")
	assert.True(t, ok)
	assert.Equal(t, "Frozen Foods", aisle)
}

func TestLookupMiss(t *testing.T) {
	m := aisles.AisleMap{"chicken": "Meat & Seafood"}
	_, ok := m.Lookup("dragonfruit")
	assert.False(t, ok)
}

func TestLoadAndSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	original := aisles.AisleMap{"milk": "Dairy & Eggs"}
	require.NoError(t, original.Save(path))

	loaded, err := aisles.Load(path)
	require.NoError(t, err)
	assert.Equal(t, original, loaded)
}

func TestLoadMissingFile(t *testing.T) {
	_, err := aisles.Load("/nonexistent/path.json")
	assert.Error(t, err)
}
