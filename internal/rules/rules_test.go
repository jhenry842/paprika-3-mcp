package rules_test

import (
	"path/filepath"
	"testing"

	"github.com/soggycactus/paprika-3-mcp/internal/rules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpsertAdd(t *testing.T) {
	r := rules.Rules{}
	rule := rules.Rule{ID: "double-proteins", Type: "quantity_multiplier", Description: "Buy double proteins"}
	r = r.Upsert(rule)
	assert.Len(t, r, 1)
	assert.Equal(t, rule, r[0])
}

func TestUpsertReplace(t *testing.T) {
	original := rules.Rule{ID: "double-proteins", Type: "quantity_multiplier", Description: "old"}
	updated := rules.Rule{ID: "double-proteins", Type: "quantity_multiplier", Description: "new"}
	r := rules.Rules{original}
	r = r.Upsert(updated)
	assert.Len(t, r, 1)
	assert.Equal(t, "new", r[0].Description)
}

func TestUpsertPreservesOrder(t *testing.T) {
	r := rules.Rules{
		{ID: "a", Type: "note", Description: "first"},
		{ID: "b", Type: "note", Description: "second"},
	}
	r = r.Upsert(rules.Rule{ID: "a", Type: "note", Description: "updated first"})
	assert.Len(t, r, 2)
	assert.Equal(t, "updated first", r[0].Description)
	assert.Equal(t, "second", r[1].Description)
}

func TestLoadSaveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.json")

	original := rules.Rules{
		{
			ID:          "substitute-venison",
			Type:        "substitution",
			Description: "Use ground beef instead",
			Params:      map[string]any{"from": "venison", "to": "ground beef"},
		},
	}
	require.NoError(t, original.Save(path))

	loaded, err := rules.Load(path)
	require.NoError(t, err)
	assert.Len(t, loaded, 1)
	assert.Equal(t, original[0].ID, loaded[0].ID)
	assert.Equal(t, original[0].Type, loaded[0].Type)
	assert.Equal(t, original[0].Description, loaded[0].Description)
}

func TestLoadMissingFileReturnsEmpty(t *testing.T) {
	r, err := rules.Load("/nonexistent/path.json")
	require.NoError(t, err)
	assert.Empty(t, r)
}

func TestToMarkdownEmpty(t *testing.T) {
	r := rules.Rules{}
	assert.Equal(t, "No household rules configured.", r.ToMarkdown())
}

func TestToMarkdownContainsRuleID(t *testing.T) {
	r := rules.Rules{
		{ID: "double-proteins", Type: "quantity_multiplier", Description: "Buy double proteins"},
	}
	md := r.ToMarkdown()
	assert.Contains(t, md, "double-proteins")
	assert.Contains(t, md, "quantity_multiplier")
	assert.Contains(t, md, "Buy double proteins")
}
