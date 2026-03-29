---
name: what-can-i-make
description: Use this skill when the user asks "what can I make tonight?", "what can I cook?", "what's for dinner?", "what do I have to cook with?", or wants recipe suggestions based on what's currently in the pantry.
---

# What Can I Make Tonight?

Suggest recipes the user can make from what's already in the pantry. Fast, opinionated, no shopping required.

## Step 1: Get Pantry and Recipes

Call `get_pantry` and `list_recipes` in parallel.

From the pantry, identify **in-stock proteins** and **notable in-stock ingredients** (specialty items, whole vegetables, anything worth using up). Ignore pantry items that are `in_stock: no`.

Proteins: beef (ground, stew, roast, brisket, steak, short ribs), chicken (breast, thigh, drumstick, whole, ground), pork (chops, tenderloin, ground, sausage, bacon, ham), fish and seafood (salmon, cod, tilapia, shrimp, tuna, halibut), lamb, turkey.

## Step 2: Match Recipes to Pantry

Filter the recipe list to candidates that are makeable from in-stock ingredients. Use recipe name and categories as the primary signal — don't call `get_recipe` for every recipe.

**A recipe is a candidate if:**
- Its name or categories suggest it uses at least one in-stock protein, OR
- It's a pantry-heavy recipe (pasta, soup, stir fry, etc.) and the pantry has the key ingredients suggested by the name

**Exclude recipes where:**
- The required protein is clearly not in stock (e.g., "Salmon Tacos" when no fish is in stock)
- The name suggests a specialty ingredient that's unlikely to be on hand unless it's in the pantry

When in doubt, include the recipe — false positives are better than false negatives at this stage.

## Step 3: Rank Candidates

Score each candidate recipe:

1. **Recency** — Prefer recipes not cooked in the last 3 weeks (use Last Prepared from `list_recipes`). Penalize anything cooked in the last 7 days — only surface it if options are very limited.
2. **Rating** — Prefer 4–5 star recipes. 3-star recipes are fine. Below 3 stars, deprioritize unless options are scarce.
3. **Protein match** — Prefer recipes that use in-stock proteins over pantry-only recipes.

## Step 4: Deep-Check Top Candidates

Take the top 6–8 ranked candidates. For each, call `get_recipe` to check the full ingredient list against the pantry.

**Classify each:**
- **Ready to make** — all key ingredients are in stock. Minor pantry staples (salt, olive oil, spices) can be assumed present even if not listed.
- **Almost there** — in-stock protein + most ingredients covered; only 1–3 items missing. List the gaps.
- **Missing too much** — more than 3 key ingredients out of stock. Drop from results.

## Step 5: Present Options

Present results in two groups:

### Ready to Make
List recipes you can cook right now. For each: name, rating, last prepared date (if any), and one sentence on why it fits (e.g., "Uses the ground beef and takes 30 minutes").

### Almost There
List recipes that need 1–3 items. For each: name, and the specific missing ingredients. Keep this list short (3 items max) — it's a nudge, not a shopping list.

Keep the total list to **5 recipes max** across both groups. Opinionated curation beats an exhaustive dump.

## Step 6: Follow-Up

After presenting options, ask:

> Want me to add one of these to tonight's dinner slot on the meal plan?

If yes, call `add_meal_to_plan` with today's date and the chosen recipe UID. Use meal type `"dinner"`.

---

## Notes

- **Speed over completeness.** The point is a quick answer, not an exhaustive pantry audit. Lean toward presenting options fast.
- **Don't call `get_recipe` for all recipes** — only the top candidates from Step 2. The full recipe list can be 100+ items.
- **Pantry staples can be assumed.** Salt, pepper, cooking oil, garlic, onion, and common dried spices don't need to be explicitly in-stock to count a recipe as makeable.
- **If the pantry is bare**, say so clearly and suggest running `plan-the-week` instead.
