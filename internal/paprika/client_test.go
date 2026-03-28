package paprika_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/soggycactus/paprika-3-mcp/internal/paprika"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroceryClient(t *testing.T) {
	username := os.Getenv("PAPRIKA_USERNAME")
	password := os.Getenv("PAPRIKA_PASSWORD")
	if username == "" || password == "" {
		t.Skip("PAPRIKA_USERNAME and PAPRIKA_PASSWORD not set")
	}

	client, err := paprika.NewClient(username, password, "dev", nil)
	require.NoError(t, err)

	ctx := context.Background()

	// List grocery items
	items, err := client.ListGroceryItems(ctx)
	require.NoError(t, err)
	// Just verify no error — list may be empty
	assert.NotNil(t, items)

	// If there are items, update the aisle on the first one and verify it round-trips.
	if len(items) > 0 {
		original := items[0]
		modified := original
		modified.Aisle = "Test Aisle"
		require.NoError(t, client.UpdateGroceryItem(ctx, modified))

		// Re-fetch and verify
		updated, err := client.ListGroceryItems(ctx)
		require.NoError(t, err)
		var found bool
		for _, item := range updated {
			if item.UID == original.UID {
				assert.Equal(t, "Test Aisle", item.Aisle)
				found = true
				break
			}
		}
		assert.True(t, found, "updated item not found in re-fetch")

		// Restore original aisle
		require.NoError(t, client.UpdateGroceryItem(ctx, original))
	}
}

func TestMealPlanClient(t *testing.T) {
	username := os.Getenv("PAPRIKA_USERNAME")
	password := os.Getenv("PAPRIKA_PASSWORD")
	if username == "" || password == "" {
		t.Skip("PAPRIKA_USERNAME and PAPRIKA_PASSWORD not set")
	}

	client, err := paprika.NewClient(username, password, "dev", nil)
	require.NoError(t, err)

	ctx := context.Background()
	start := time.Now()
	end := start.Add(7 * 24 * time.Hour)

	entries, err := client.ListMealPlanEntries(ctx, start, end)
	require.NoError(t, err)
	assert.NotNil(t, entries)
}

func TestClient(t *testing.T) {
	username := os.Getenv("PAPRIKA_USERNAME")
	password := os.Getenv("PAPRIKA_PASSWORD")
	if username == "" || password == "" {
		t.Skip("PAPRIKA_USERNAME and PAPRIKA_PASSWORD not set")
	}
	client, err := paprika.NewClient(username, password, "dev", nil)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	testRecipe := paprika.Recipe{
		Name:        fmt.Sprintf("Test Recipe - %d", time.Now().Unix()),
		Notes:       "Notes",
		Directions:  "Directions",
		Ingredients: "Ingredients",
		Servings:    "Servings",
		Source:      "Source",
		SourceURL:   "URL",
		Categories:  []string{},
	}
	recipe, err := client.SaveRecipe(ctx, testRecipe)
	require.NoError(t, err)

	recipe, err = client.GetRecipe(ctx, recipe.UID)
	require.NoError(t, err)
	assert.NotEmpty(t, recipe.UID)
	assert.Equal(t, testRecipe.Name, recipe.Name)
	assert.Equal(t, testRecipe.Notes, recipe.Notes)
	assert.Equal(t, testRecipe.Directions, recipe.Directions)
	assert.Equal(t, testRecipe.Ingredients, recipe.Ingredients)
	assert.Equal(t, testRecipe.Servings, recipe.Servings)
	assert.Equal(t, testRecipe.Source, recipe.Source)
	assert.Equal(t, testRecipe.SourceURL, recipe.SourceURL)
	assert.Equal(t, testRecipe.Categories, recipe.Categories)

	t.Logf("Created and fetched recipe: %+v", recipe)

	newDescription := "Updated Description"
	recipe.Description = newDescription
	uid := recipe.UID
	recipe, err = client.SaveRecipe(ctx, *recipe)
	require.NoError(t, err)
	assert.Equal(t, newDescription, recipe.Description)
	assert.Equal(t, uid, recipe.UID)
	assert.Equal(t, testRecipe.Name, recipe.Name)
	assert.Equal(t, testRecipe.Notes, recipe.Notes)
	assert.Equal(t, testRecipe.Directions, recipe.Directions)
	assert.Equal(t, testRecipe.Ingredients, recipe.Ingredients)
	assert.Equal(t, testRecipe.Servings, recipe.Servings)
	assert.Equal(t, testRecipe.Source, recipe.Source)
	assert.Equal(t, testRecipe.SourceURL, recipe.SourceURL)
	assert.Equal(t, testRecipe.Categories, recipe.Categories)

	t.Logf("Updated recipe: %+v", recipe)

	_, err = client.DeleteRecipe(ctx, *recipe)
	require.NoError(t, err)
	t.Logf("Deleted recipe: %s", recipe.Name)

	recipes, err := client.ListRecipes(ctx)
	require.NoError(t, err)

	for _, recipe := range recipes.Result {
		r, err := client.GetRecipe(ctx, recipe.UID)
		require.NoError(t, err)

		t.Logf("Recipe: %s - %s", r.Name, r.Created)
		if _, err := json.Marshal(r); err != nil {
			t.Logf("Failed to marshal recipe: %s", err)
		}
	}
}
