package paprika_test

import (
	"bytes"
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

func TestSaveAndDeleteGroceryItem(t *testing.T) {
	username := os.Getenv("PAPRIKA_USERNAME")
	password := os.Getenv("PAPRIKA_PASSWORD")
	if username == "" || password == "" {
		t.Skip("PAPRIKA_USERNAME and PAPRIKA_PASSWORD not set")
	}

	client, err := paprika.NewClient(username, password, "dev", nil)
	require.NoError(t, err)

	ctx := context.Background()

	// Grab list_uid from existing items — it's a required field.
	existing, err := client.ListGroceryItems(ctx)
	require.NoError(t, err)
	var listUID string
	for _, item := range existing {
		if item.ListUID != "" {
			listUID = item.ListUID
			break
		}
	}

	testItem := paprika.GroceryItem{
		Name:       fmt.Sprintf("Test Item %d", time.Now().Unix()),
		Ingredient: fmt.Sprintf("test ingredient %d", time.Now().Unix()),
		Quantity:   "1 unit",
		ListUID:    listUID,
	}

	// Save
	require.NoError(t, client.SaveGroceryItem(ctx, testItem))

	// Verify it appears
	items, err := client.ListGroceryItems(ctx)
	require.NoError(t, err)
	var found *paprika.GroceryItem
	for i := range items {
		if items[i].Ingredient == testItem.Ingredient {
			found = &items[i]
			break
		}
	}
	require.NotNil(t, found, "saved grocery item not found in list")
	assert.Equal(t, testItem.Quantity, found.Quantity)
	t.Logf("Saved grocery item: %s (UID: %s)", found.Name, found.UID)

	// Delete
	require.NoError(t, client.DeleteGroceryItem(ctx, *found))

	// Verify it's gone
	items, err = client.ListGroceryItems(ctx)
	require.NoError(t, err)
	for _, item := range items {
		if item.UID == found.UID {
			t.Errorf("deleted grocery item still present in list: %s", item.UID)
		}
	}
}

func TestUncheckGroceryItem(t *testing.T) {
	username := os.Getenv("PAPRIKA_USERNAME")
	password := os.Getenv("PAPRIKA_PASSWORD")
	if username == "" || password == "" {
		t.Skip("PAPRIKA_USERNAME and PAPRIKA_PASSWORD not set")
	}

	client, err := paprika.NewClient(username, password, "dev", nil)
	require.NoError(t, err)

	ctx := context.Background()

	existing, err := client.ListGroceryItems(ctx)
	require.NoError(t, err)
	var listUID string
	for _, item := range existing {
		if item.ListUID != "" {
			listUID = item.ListUID
			break
		}
	}

	// Create item with purchased=true
	testItem := paprika.GroceryItem{
		Name:       fmt.Sprintf("Test Staple %d", time.Now().Unix()),
		Ingredient: fmt.Sprintf("test staple %d", time.Now().Unix()),
		Quantity:   "1 bunch",
		ListUID:    listUID,
		Purchased:  true,
	}
	require.NoError(t, client.SaveGroceryItem(ctx, testItem))

	// Fetch to get the assigned UID
	items, err := client.ListGroceryItems(ctx)
	require.NoError(t, err)
	var found *paprika.GroceryItem
	for i := range items {
		if items[i].Ingredient == testItem.Ingredient {
			found = &items[i]
			break
		}
	}
	require.NotNil(t, found, "saved grocery item not found")
	assert.True(t, found.Purchased, "item should be purchased=true after save")

	// Uncheck it
	found.Purchased = false
	require.NoError(t, client.UpdateGroceryItem(ctx, *found))

	// Verify it's still on the list but unchecked
	items, err = client.ListGroceryItems(ctx)
	require.NoError(t, err)
	var unchecked *paprika.GroceryItem
	for i := range items {
		if items[i].UID == found.UID {
			unchecked = &items[i]
			break
		}
	}
	require.NotNil(t, unchecked, "item should still be on grocery list after uncheck")
	assert.False(t, unchecked.Purchased, "item should be purchased=false after uncheck")

	// Cleanup
	require.NoError(t, client.DeleteGroceryItem(ctx, *unchecked))
}

func TestPantryClient(t *testing.T) {
	username := os.Getenv("PAPRIKA_USERNAME")
	password := os.Getenv("PAPRIKA_PASSWORD")
	if username == "" || password == "" {
		t.Skip("PAPRIKA_USERNAME and PAPRIKA_PASSWORD not set")
	}

	client, err := paprika.NewClient(username, password, "dev", nil)
	require.NoError(t, err)

	ctx := context.Background()

	// List — just verify no error
	items, err := client.ListPantryItems(ctx)
	require.NoError(t, err)
	assert.NotNil(t, items)

	// Add a test item
	testIngredient := fmt.Sprintf("test pantry item %d", time.Now().Unix())
	testItem := paprika.PantryItem{
		Ingredient: testIngredient,
		Quantity:   "1 unit",
		InStock:    true,
	}
	require.NoError(t, client.SavePantryItem(ctx, testItem))

	// Verify it appears
	items, err = client.ListPantryItems(ctx)
	require.NoError(t, err)
	var found *paprika.PantryItem
	for i := range items {
		if items[i].Ingredient == testIngredient {
			found = &items[i]
			break
		}
	}
	require.NotNil(t, found, "saved pantry item not found in list")
	assert.Equal(t, "1 unit", found.Quantity)
	assert.True(t, found.InStock)
	t.Logf("Saved pantry item: %s (UID: %s)", found.Ingredient, found.UID)

	// Update quantity and in_stock
	found.Quantity = "2 units"
	found.InStock = false
	require.NoError(t, client.SavePantryItem(ctx, *found))

	// Verify update round-trips
	items, err = client.ListPantryItems(ctx)
	require.NoError(t, err)
	var updated *paprika.PantryItem
	for i := range items {
		if items[i].UID == found.UID {
			updated = &items[i]
			break
		}
	}
	require.NotNil(t, updated, "updated pantry item not found")
	assert.Equal(t, "2 units", updated.Quantity)
	assert.False(t, updated.InStock)

	// No DeletePantryItem exists — test item persists in pantry.
	// Remove manually from the app after running, or add DeletePantryItem if cleanup becomes painful.
	t.Logf("Note: test pantry item left in place (no delete method): %s", found.UID)
}

func TestDeleteMealPlanEntry(t *testing.T) {
	username := os.Getenv("PAPRIKA_USERNAME")
	password := os.Getenv("PAPRIKA_PASSWORD")
	if username == "" || password == "" {
		t.Skip("PAPRIKA_USERNAME and PAPRIKA_PASSWORD not set")
	}

	client, err := paprika.NewClient(username, password, "dev", nil)
	require.NoError(t, err)

	ctx := context.Background()

	// Add a test meal plan entry for today
	today := time.Now().Format("2006-01-02") + " 00:00:00"
	entry := paprika.MealPlanEntry{
		RecipeName: "Test Meal Plan Entry (DELETE ME)",
		Date:       today,
		MealType:   paprika.MealTypeDinner,
	}
	require.NoError(t, client.SaveMealPlanEntry(ctx, entry))

	// Fetch to get the assigned UID
	start := time.Now().Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour)
	entries, err := client.ListMealPlanEntries(ctx, start, end)
	require.NoError(t, err)
	var saved *paprika.MealPlanEntry
	for i := range entries {
		if entries[i].RecipeName == entry.RecipeName {
			saved = &entries[i]
			break
		}
	}
	require.NotNil(t, saved, "saved meal plan entry not found in list")
	t.Logf("Created meal plan entry: %s (UID: %s)", saved.RecipeName, saved.UID)

	// Delete it using the soft-delete pattern
	require.NoError(t, client.DeleteMealPlanEntry(ctx, *saved))

	// Verify it's gone from the list
	entries, err = client.ListMealPlanEntries(ctx, start, end)
	require.NoError(t, err)
	for _, e := range entries {
		if e.UID == saved.UID {
			t.Errorf("deleted meal plan entry still present in list: %s", e.UID)
		}
	}
	t.Log("Meal plan entry successfully deleted")
}

func TestRecipeRawFields(t *testing.T) {
	username := os.Getenv("PAPRIKA_USERNAME")
	password := os.Getenv("PAPRIKA_PASSWORD")
	if username == "" || password == "" {
		t.Skip("PAPRIKA_USERNAME and PAPRIKA_PASSWORD not set")
	}

	client, err := paprika.NewClient(username, password, "dev", nil)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Dump raw fields of the first recipe to discover all available API fields
	recipes, err := client.ListRecipes(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, recipes.Result, "no recipes found — need at least one to inspect")

	raw, err := client.GetRecipeRaw(ctx, recipes.Result[0].UID)
	require.NoError(t, err)

	var pretty bytes.Buffer
	require.NoError(t, json.Indent(&pretty, raw, "", "  "))
	t.Logf("Raw recipe fields:\n%s", pretty.String())
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
