---
name: generate-grocery-list
description: This skill should be used when the user asks to "generate a grocery list", "build the grocery list", "create grocery list from meal plan", "add meals to grocery list", "populate the grocery list for the week", "what do I need to buy", or wants to turn their weekly Paprika meal plan into a shopping list. Always use this skill when the user wants to generate groceries from their meal plan, even if they don't say "skill".
---

# Generate Grocery List from Meal Plan

Generate a grocery list by reading the weekly meal plan, fetching each recipe, cross-referencing the pantry, applying household rules, and adding needed items to the Paprika grocery list.

## Workflow

### Step 1: Get the Meal Plan

Call `get_meal_plan` for the current week (Monday through Sunday). If the user specifies a different date range, use that instead.

Collect the unique set of recipe UIDs across all meal plan entries. Multiple meals using the same recipe count as one recipe to fetch (but will multiply quantities in deduplication).

### Step 2: Fetch Recipes

Call `get_recipe` for each unique recipe UID. Parse the `ingredients` field — it is newline-separated, one ingredient per line. Each line looks like:

- `2 lbs ground beef`
- `1 can (15 oz) black beans`
- `3 cloves garlic, minced`
- `salt and pepper to taste`
- `1/2 cup olive oil`

For each line, extract:
- **ingredient name**: canonical lowercase form, stripped of preparation notes ("garlic, minced" → "garlic", "chicken breast, boneless" → "chicken breast")
- **quantity**: the number + unit ("2 lbs", "1 can", "3 cloves", "1/2 cup")

If a recipe appears multiple times in the meal plan (e.g., dinner Monday and dinner Thursday), multiply its ingredient quantities accordingly before deduplicating.

### Step 3: Get the Pantry

Call `get_pantry` once. For each ingredient needed, check if it exists in the pantry with `in_stock: yes`.

**Matching rules:**
- Match conservatively. "Chicken breast" and "chicken thigh" are fundamentally different — never treat one as covering the other.
- "Ground beef" and "beef stew meat" are different — don't merge.
- Minor variation is OK: "olive oil" and "extra virgin olive oil" can be treated as the same pantry item.
- When in doubt, keep the item on the list. Missing an ingredient is worse than buying something you have.

Only skip an item if its pantry match is `in_stock: yes`. If it's `in_stock: no`, it still needs to be purchased.

### Step 4: Apply Household Rules

Call `get_household_rules` and apply every rule returned to ingredients that are NOT being skipped (i.e., not fully covered by the in-stock pantry).

Interpret rules by type:

**`substitution`** — Replace `params.from` ingredient with `params.to` wherever it appears. Do not add the original ingredient; add the substitute instead. Record the swap in the output summary.

**`quantity_multiplier`** — For every ingredient whose category matches `params.category`, multiply the post-deduplication quantity by `params.multiplier`. For a `category` of `"protein"`, proteins include: beef (ground, stew, roast, brisket, steak, short ribs), chicken (breast, thigh, drumstick, whole, ground — each cut separate), pork (chops, tenderloin, ground, sausage, bacon, ham), fish and seafood (salmon, cod, tilapia, shrimp, tuna, halibut), lamb, turkey, eggs (a dozen = 12 eggs). Do not apply to pantry staples used as flavor (broth, stock, anchovies).

**`note`** — Informational. No mechanical transformation; use the description to guide judgment calls.

Apply substitutions first (they change what ingredient is on the list), then quantity multipliers (they change how much).

---

**General buffer — everything else: +25–50%**

Round up non-protein quantities to practical grocery increments. Use judgment:
- "1 lb flour" → buy 2 lbs (next standard bag size)
- "1 can black beans" → buy 2 cans
- "1/2 cup olive oil" → buy 1 bottle if not in pantry
- "to taste" ingredients (salt, pepper, spices) → if not in pantry, add them at a single unit; no buffer needed
- Fresh produce: round up to the next natural unit (buy 3 onions instead of 2.5)

---

### Step 5: Deduplicate Across Recipes

Before calling `add_grocery_item`, collect all ingredients from all recipes and merge by ingredient name:

- "2 lbs ground beef" (recipe A) + "1 lb ground beef" (recipe B) → 3 lbs ground beef → apply Rule 2 → 6 lbs ground beef
- Apply household rules AFTER deduplication so totals are correct
- "Chicken breast" and "chicken thigh" are different — never merge them
- Combine quantities in the same unit when possible; if units differ (cups vs. cans), list separately

### Step 6: Check the Existing Grocery List

Call `get_grocery_list` to see what's already there. Skip any ingredients already on the list (conservative name match — "ground beef" already present means don't add again).

### Step 7: Add to Grocery List

For each ingredient that needs to be purchased, call `add_grocery_item`:

- `name`: human-readable display name including quantity — e.g., `"Ground Beef (6 lbs)"`, `"Chicken Breast (4 lbs)"`, `"Black Beans (2 cans)"`
- `ingredient`: canonical lowercase ingredient name — e.g., `"ground beef"`, `"chicken breast"`, `"black beans"`
- `quantity`: quantity string — e.g., `"6 lbs"`, `"4 lbs"`, `"2 cans"`
- `recipe`: name of one recipe it's for (pick one if it's used in multiple)
- `recipe_uid`: UID of that recipe

Aisle assignment is automatic — do not set it manually.

### Step 8: Suggest Based on Purchase History

After all recipe-derived items are added, check whether past cycles reveal recurring non-dinner purchases that aren't on this list yet.

1. From the household rules already fetched in Step 4, find the rule with `id: "grocery-history"`. If it doesn't exist yet, skip this step silently.
2. Look at `params.cycles` — use all available entries (up to 5). Find ingredient names that appear in **2 or more** of those cycles.
3. Filter out:
   - Anything already on the current grocery list (added in this run or already present from Step 6)
   - Anything covered by an in-stock pantry item that was skipped in Step 3
   - Anything matching a `staple` rule (`params.ingredient`) — staples have their own path and stay on the list automatically
4. If candidates remain, ask conversationally:
   > "Based on your past shopping, you've consistently bought: **[item A]**, **[item B]**, **[item C]** — none of those are on the list yet. Want me to add any of them?"
   
   Wait for the user's response. Add confirmed items via `add_grocery_item` (omit `recipe` and `recipe_uid` — these aren't recipe-derived).
5. If no candidates, skip silently — don't mention history at all.

## Output Summary

After adding all items, report:

1. **Added** — list of items added with quantities
2. **Skipped (in pantry)** — items already in stock
3. **Substitutions** — any ingredient swaps applied from household rules
4. **Already on list** — items that were already in the grocery list
5. **Needs review** — any ingredient lines that couldn't be parsed cleanly
