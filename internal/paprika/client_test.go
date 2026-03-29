package paprika_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
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

	// Clean up
	require.NoError(t, client.DeletePantryItem(ctx, *updated))
	items, err = client.ListPantryItems(ctx)
	require.NoError(t, err)
	for _, item := range items {
		if item.UID == found.UID {
			t.Errorf("deleted pantry item still present: %s", item.UID)
		}
	}
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

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	recipes, err := client.ListRecipes(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, recipes.Result, "no recipes found — need at least one to inspect")

	// Known fields already mapped in the Recipe struct.
	knownFields := map[string]bool{
		"uid": true, "name": true, "ingredients": true, "directions": true,
		"description": true, "notes": true, "nutritional_info": true, "servings": true,
		"difficulty": true, "prep_time": true, "cook_time": true, "total_time": true,
		"source": true, "source_url": true, "image_url": true, "photo": true,
		"photo_hash": true, "photo_large": true, "scale": true, "hash": true,
		"categories": true, "rating": true, "in_trash": true, "is_pinned": true,
		"on_favorites": true, "on_grocery_list": true, "created": true, "photo_url": true,
	}

	unknownFields := map[string]interface{}{}

	// Scan all recipes for any field not in knownFields, or any field with "cook",
	// "prepared", "last", "history", or "date" in its name.
	t.Logf("Scanning %d recipes for unknown/cook-related fields...", len(recipes.Result))
	for _, r := range recipes.Result {
		raw, err := client.GetRecipeRaw(ctx, r.UID)
		if err != nil {
			t.Logf("  skip %s: %v", r.UID, err)
			continue
		}

		var wrapper struct {
			Result map[string]interface{} `json:"result"`
		}
		if err := json.Unmarshal(raw, &wrapper); err != nil {
			continue
		}

		for k, v := range wrapper.Result {
			if !knownFields[k] {
				if _, seen := unknownFields[k]; !seen {
					unknownFields[k] = v
					t.Logf("  UNKNOWN FIELD %q = %v (recipe %s)", k, v, r.UID)
				} else if v != nil && unknownFields[k] == nil {
					// Update to a non-nil value if we found one
					unknownFields[k] = v
					t.Logf("  UNKNOWN FIELD %q = %v (non-nil, recipe %s)", k, v, r.UID)
				}
			}
			// Also flag known fields that contain cook/prepared/last/date/history
			name := strings.ToLower(k)
			if strings.Contains(name, "cook") || strings.Contains(name, "prepared") ||
				strings.Contains(name, "last") || strings.Contains(name, "history") {
				t.Logf("  COOK-RELATED FIELD %q = %v", k, v)
			}
		}
	}

	if len(unknownFields) == 0 {
		t.Log("No unknown fields found across all recipes.")
	} else {
		t.Logf("Found %d unknown field(s) total.", len(unknownFields))
	}
}

func TestProbeHistoryEndpoints(t *testing.T) {
	username := os.Getenv("PAPRIKA_USERNAME")
	password := os.Getenv("PAPRIKA_PASSWORD")
	if username == "" || password == "" {
		t.Skip("PAPRIKA_USERNAME and PAPRIKA_PASSWORD not set")
	}

	client, err := paprika.NewClient(username, password, "dev", nil)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	candidates := []string{
		// History / cook log guesses
		"https://paprikaapp.com/api/v2/sync/history/",
		"https://paprikaapp.com/api/v1/sync/history/",
		"https://paprikaapp.com/api/v2/sync/cookhistory/",
		"https://paprikaapp.com/api/v2/sync/cooked/",
		"https://paprikaapp.com/api/v2/sync/logs/",
		"https://paprikaapp.com/api/v2/sync/timeline/",
		"https://paprikaapp.com/api/v2/sync/activity/",
		// More candidates
		"https://paprikaapp.com/api/v2/sync/cooklogs/",
		"https://paprikaapp.com/api/v2/sync/made/",
		"https://paprikaapp.com/api/v2/sync/prepared/",
		"https://paprikaapp.com/api/v2/sync/diary/",
		"https://paprikaapp.com/api/v2/sync/journal/",
		"https://paprikaapp.com/api/v1/sync/cooklogs/",
		"https://paprikaapp.com/api/v1/sync/made/",
		"https://paprikaapp.com/api/v1/sync/prepared/",
	}

	for _, url := range candidates {
		status, body, err := client.ProbeEndpoint(ctx, url)
		if err != nil {
			t.Logf("  %s → error: %v", url, err)
			continue
		}
		preview := string(body)
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		t.Logf("  %s → %d: %s", url, status, preview)
	}
}

func TestMealPlanHistory(t *testing.T) {
	username := os.Getenv("PAPRIKA_USERNAME")
	password := os.Getenv("PAPRIKA_PASSWORD")
	if username == "" || password == "" {
		t.Skip("PAPRIKA_USERNAME and PAPRIKA_PASSWORD not set")
	}

	client, err := paprika.NewClient(username, password, "dev", nil)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch meal plan entries going back 1 year to see if historical entries persist.
	start := time.Now().AddDate(-1, 0, 0)
	end := time.Now()
	entries, err := client.ListMealPlanEntries(ctx, start, end)
	require.NoError(t, err)

	t.Logf("Found %d meal plan entries over past year", len(entries))
	for _, e := range entries {
		t.Logf("  %s — %s (%s)", e.Date, e.RecipeName, e.UID)
	}

	// Check if the raw meals endpoint returns more fields than what MealPlanEntry maps.
	rawStatus, rawBody, err := client.ProbeEndpoint(ctx, "https://paprikaapp.com/api/v2/sync/meals/")
	require.NoError(t, err)
	require.Equal(t, 200, rawStatus)

	var mealsResp struct {
		Result []map[string]interface{} `json:"result"`
	}
	require.NoError(t, json.Unmarshal(rawBody, &mealsResp))
	if len(mealsResp.Result) > 0 {
		t.Logf("Raw meal plan entry fields: %v", mealsResp.Result[0])
	}
}

func TestDeletePantryItem(t *testing.T) {
	username := os.Getenv("PAPRIKA_USERNAME")
	password := os.Getenv("PAPRIKA_PASSWORD")
	if username == "" || password == "" {
		t.Skip("PAPRIKA_USERNAME and PAPRIKA_PASSWORD not set")
	}

	client, err := paprika.NewClient(username, password, "dev", nil)
	require.NoError(t, err)

	ctx := context.Background()

	testItem := paprika.PantryItem{
		Ingredient: fmt.Sprintf("test delete pantry item %d", time.Now().Unix()),
		Quantity:   "1 unit",
		InStock:    true,
	}
	require.NoError(t, client.SavePantryItem(ctx, testItem))

	items, err := client.ListPantryItems(ctx)
	require.NoError(t, err)
	var saved *paprika.PantryItem
	for i := range items {
		if items[i].Ingredient == testItem.Ingredient {
			saved = &items[i]
			break
		}
	}
	require.NotNil(t, saved, "saved pantry item not found")
	t.Logf("Created pantry item: %s (UID: %s)", saved.Ingredient, saved.UID)

	require.NoError(t, client.DeletePantryItem(ctx, *saved))

	items, err = client.ListPantryItems(ctx)
	require.NoError(t, err)
	for _, item := range items {
		if item.UID == saved.UID {
			t.Errorf("deleted pantry item still present: %s", item.UID)
		}
	}
	t.Log("Pantry item successfully deleted")
}

func TestGetLastPreparedDates(t *testing.T) {
	username := os.Getenv("PAPRIKA_USERNAME")
	password := os.Getenv("PAPRIKA_PASSWORD")
	if username == "" || password == "" {
		t.Skip("PAPRIKA_USERNAME and PAPRIKA_PASSWORD not set")
	}

	client, err := paprika.NewClient(username, password, "dev", nil)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dates, err := client.GetLastPreparedDates(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, dates, "expected at least one recipe with a last prepared date")

	// Verify against the Slow Cooker Chicken Burrito Bowl — app shows Last Prepared 2/23/26.
	// Look up its recipe UID from the recipes list rather than hardcoding.
	recipes, err := client.ListRecipes(ctx)
	require.NoError(t, err)
	for _, r := range recipes.Result {
		if lp, ok := dates[r.UID]; ok {
			full, err := client.GetRecipe(ctx, r.UID)
			if err != nil {
				continue
			}
			if strings.Contains(full.Name, "Burrito Bowl") {
				assert.Equal(t, "2026-02-23", lp.Format("2006-01-02"),
					"Slow Cooker Chicken Burrito Bowl last prepared date should match app display of 2/23/26")
				t.Logf("Slow Cooker Chicken Burrito Bowl (UID %s) last prepared: %s ✓", r.UID, lp.Format("2006-01-02"))
				break
			}
		}
	}

	t.Logf("Total recipes with last prepared dates: %d", len(dates))
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
