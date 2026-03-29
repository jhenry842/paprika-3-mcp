---
name: close-cycle
description: Use this skill when the user is done with their current meal cycle and wants to close the loop — "close the cycle", "I'm done cooking", "full sync", "sync everything", "wrap up", "done with this cycle", "run the sync". This is the ONE canonical end-of-cycle operation. It depletes the pantry from cooked meals, restocks from shopping, and advances last_sync_date. Never run depletion or grocery sync separately as a substitute for this workflow.
---

# Close Cycle

The canonical end-of-cycle sync. Always runs in this order:
1. **Deplete** — mark consumed ingredients out-of-stock based on cooked meals since last sync
2. **Restock** — sync purchased grocery items to pantry
3. **Advance** — set `last_sync_date` to today

Order is non-negotiable: deplete before restock, or items you just bought get incorrectly marked out-of-stock.

---

## Step 1: Get the Sync Anchor

Call `get_household_rules`. Look for a rule with `id: "last-sync-date"`.

**If the rule exists:** extract the date from `params.date`. This is the start of the depletion window.

**If the rule does not exist:** this is a clean start. Skip Steps 2–5 entirely. Tell the user:

> No previous sync date found — skipping depletion for a clean start. I'll restock from your grocery list and set today as the sync anchor going forward.

Then jump to Step 6.

---

## Step 2: Identify Cooked Meals

Call `get_meal_plan` with start = `last_sync_date` and end = today.

Filter to entries that have a `recipe_uid` (skip freeform notes with no recipe attached — these can't be depleted). If an entry has no `recipe_uid`, log it as skipped.

Collect the list of recipe UIDs. If the same recipe appears multiple times (cooked more than once in the cycle), note the count — you'll need to multiply ingredient quantities.

If no recipe-linked meals are found in the window, skip Steps 3–5 and note that nothing was depleted.

---

## Step 3: Fetch Recipes and Parse Ingredients

Call `get_recipe` for each unique recipe UID. Parse the `ingredients` field — it is newline-separated, one ingredient per line.

For each ingredient line, extract:
- **Ingredient name** — canonical lowercase, stripped of preparation notes ("chicken breast, boneless" → "chicken breast", "garlic, minced" → "garlic")
- **Quantity** — number + unit ("2 lbs", "1 can", "3 cloves", "1/2 cup", "2 tbsp")
- **Ignore** — lines that are section headers, garnishes, or "to taste" items

**Aggregate across the full cycle:** if a recipe was cooked N times, multiply all its quantities by N. Then combine across all recipes — if multiple recipes use "chicken breast", sum those quantities together before touching the pantry.

Final result: a list of `{ ingredient_name, total_quantity, source_recipes[] }` entries.

---

## Step 4: Deplete Pantry

Call `get_pantry` to get the current pantry state.

For each aggregated ingredient, attempt to find a matching pantry item:

**Matching rules:**
- Exact match first (case-insensitive)
- Then substring match (pantry item name is a substring of ingredient name, or vice versa)
- If multiple pantry items could match, pick the closest one and note the ambiguity

**Once matched — attempt quantity math:**
- Parse both the recipe quantity and the pantry quantity into number + unit
- If units are compatible (both weight, both volume, or both count), subtract the recipe amount from pantry quantity
  - If result ≤ 0: call `update_pantry_item` with `in_stock: false`
  - If result > 0: call `update_pantry_item` with the remaining quantity string and `in_stock: true`
- If units are incompatible or unparseable: call `update_pantry_item` with `in_stock: false` — and **flag this prominently** in the depletion report (see Step 5)

---

## Step 5: Depletion Report

This report must be shown before proceeding to restock. Be explicit about every decision made.

Format it in three sections:

**Depleted (matched + quantity resolved)**
List each pantry item updated, with: old quantity → new quantity or "marked out of stock".
Example: `chicken breast: 3 lbs → 1 lb`

**Flagged (matched but quantity decision uncertain)**
For every item where units were incompatible, unparseable, or a judgment call was made — show the full reasoning. This is the most important section. Do not summarize or hide anything.

Example:
> - **soy sauce**: recipe used `2 tbsp`, pantry had `"2 packages"` — units incompatible, marked out of stock
> - **olive oil**: recipe used `1/4 cup`, pantry had `"1 bottle"` — can't subtract, marked out of stock
> - **chicken thighs**: recipe used `2 lbs`, pantry had `"2 packages"` — package weight unknown, marked out of stock

**No pantry match**
Every ingredient that had no matching pantry item — list ingredient name and quantity. These are candidates for adding to the pantry.

Example:
> - `sesame oil` (2 tbsp) — not in pantry
> - `fish sauce` (1 tbsp) — not in pantry

Ask the user: "Want me to add any of the unmatched items to the pantry?" before moving on. If yes, call `add_pantry_item` for each one they confirm.

---

## Step 6: Restock from Shopping

Run the sync-grocery-list workflow inline:

1. Call `get_grocery_list`. Show checked (purchased) items.
2. Confirm with the user before proceeding.
3. Call `get_household_rules` to identify staples.
4. For each checked item: `update_pantry_item` (in_stock: true) if it exists, `add_pantry_item` if it doesn't.
5. Uncheck staples (`uncheck_grocery_items`), delete non-staples (`delete_grocery_items`).

If there are no checked items, note it and continue — don't abort the cycle close.

---

## Step 7: Advance the Sync Date

Call `set_household_rule` with:
- `id`: `"last-sync-date"`
- `type`: `"sync"`
- `description`: `"Date of last cycle close — used as the start of the next depletion window"`
- `params`: `{ "date": "<today's date in YYYY-MM-DD>" }`

Confirm to the user: "Sync date set to [date]. Next cycle will deplete from meals cooked after this date."

---

## Step 8: Summary

Wrap up with a brief summary:

> **Cycle closed [date range]**
> - Depleted [N] pantry items from [M] cooked meals
> - [K] items flagged (review above)
> - [J] ingredients had no pantry match
> - Restocked [P] items from shopping
> - Next sync anchor: [today]

Then ask: "Want me to plan next week — or next cycle — now?"

---

## Notes

- **Depletion window is last_sync_date → today.** Future planned meals are not depleted.
- **Meals without recipe_uid are skipped.** Freeform plan entries (no linked recipe) cannot be depleted. If this becomes a recurring gap, those meals should be created as proper recipes.
- **Do not run this skill if only a partial grocery sync is needed.** Use sync-grocery-list for mid-cycle top-ups. That skill does NOT advance last_sync_date.
- **Never call this skill twice in a row without cooking in between.** The depletion window would cover zero meals and the grocery sync would have nothing checked.
