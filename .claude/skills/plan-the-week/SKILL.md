---
name: plan-the-week
description: Use this skill when the user wants to plan the week's meals, "plan the week", "what should we eat this week", "set up the meal plan", or wants to go from pantry to meal plan to grocery list in one workflow.
---

# Plan the Week

A full weekly planning workflow: pantry review → meal selection → meal plan population → grocery list generation. The skill is conversational — it surfaces information, asks questions, and acts on user responses.

## Step 1: Pantry Snapshot

Call `get_pantry` and `get_household_rules`. Present a brief pantry summary to anchor the conversation:

**Anchor proteins** — list all proteins that are in-stock. These should drive meal selection. Proteins include: beef (ground, stew, roast, brisket, steak, short ribs), chicken (breast, thigh, drumstick, whole, ground), pork (chops, tenderloin, ground, sausage, bacon, ham), fish and seafood (salmon, cod, tilapia, shrimp, tuna, halibut), lamb, turkey.

**Other anchor ingredients** — list any notable in-stock pantry items that are worth using up: whole vegetables, specialty ingredients, anything that might expire soon. Keep this list short (5 items max) — focus on things that meaningfully constrain meal choice.

**Out of stock** — briefly note if any common proteins or staples are out of stock, since they'll need to be bought regardless.

Format as a short, scannable summary. Do not dump the full pantry list.

## Step 2: Confirm the Date Range

Determine the planning week. Default to the current Mon–Sun week (starting the nearest upcoming Monday, or today if today is Monday). Ask the user to confirm or specify a different range before proceeding.

## Step 3: Check the Existing Meal Plan

Call `get_meal_plan` for the confirmed date range. You can pass any future date range — the tool supports it.

**If the plan is empty:** proceed directly to planning level selection.

**If meals are already planned:** list them, then ask:

> I see you already have meals planned for some of this week. Do you want to:
> 1. **Clear the week and start fresh** — remove all current entries and plan from scratch
> 2. **Fill the empty slots only** — keep what's there and plan around it

If the user chooses to clear: call `remove_meal_from_plan` for each existing entry UID before proceeding.

Meal entry UIDs are returned by `get_meal_plan` — use them for deletion.

## Step 4: Choose a Planning Level

Ask the user how they want to plan. Present these three options clearly:

---

**Tried and True** — Pull from your existing Paprika recipe library. Prioritize recipes that use in-stock proteins and pantry staples. I'll suggest a week of meals and you confirm or swap.

**Try Something New** — Same as above but reserve 1–2 slots for recipes you haven't made recently. I'll show you a few options for those slots and you pick.

**Let AI Take the Wheel** — I'll build a full week of meals from scratch. I'll ask you a few questions (cuisines, dietary goals, how adventurous), search for recipes, create them in Paprika, and plan the week.

---

Wait for the user to respond before proceeding.

## Step 5: Select Meals

### Tried and True

Call `list_recipes` to get the full recipe library — the response includes a **Last Prepared** column. Filter to recipes that use in-stock proteins (by category tags or recipe name — use judgment). Propose 5–7 dinners that:

- Use at least one in-stock protein each
- Vary by protein type and cuisine — don't repeat the same protein two nights in a row
- Are practical for a week (not all 4-hour braises)
- **Prefer recipes not cooked in the last 3 weeks** — use Last Prepared to deprioritize anything made recently. Don't repeat a recipe that appeared on the plan in the past 7 days.

Present the proposed week as a simple list (Mon–Sun dinners). Ask the user to confirm, swap specific days, or add lunches/breakfasts if they want.

### Try Something New

Do the same as Tried and True for most slots. For 1–2 "new" slots, use Last Prepared to find recipes the user hasn't made recently (no Last Prepared date, or last prepared > 6 weeks ago) with a good rating (3+ stars). Offer 2–3 options per slot and let the user pick.

If the recipe library has no good "new" candidates, note that and ask if the user wants to search the web for one.

### Let AI Take the Wheel

Ask the user these questions before searching:

1. Any cuisines you're in the mood for this week?
2. Any dietary goals or restrictions (lighter meals, no red meat, etc.)?
3. How adventurous — familiar comfort food, or something you've never made?
4. How many nights do you want to cook vs. leftovers?

Then use web search to find recipes matching the answers and pantry anchors. For each recipe you want to create:
- Find a complete recipe (ingredients + instructions)
- Call `create_paprika_recipe` with full details
- Confirm with the user before adding it to the plan

Present the full proposed week before committing any `add_meal_to_plan` calls.

## Step 6: Confirm and Add to Plan

Once the user approves the week (in full or meal by meal), call `add_meal_to_plan` for each dinner (and any lunches/breakfasts the user requested).

Use meal type:
- `"dinner"` for dinners
- `"lunch"` for lunches
- `"breakfast"` for breakfasts

Date format: `"YYYY-MM-DD"` — bare date, no time component.

**Compute dates with arithmetic, not inference.** Monday = start_date, Tuesday = start_date+1, Wednesday = start_date+2, Thursday = start_date+3, Friday = start_date+4, Saturday = start_date+5, Sunday = start_date+6. Never infer a date from day-of-week name alone — always derive it by adding days to the confirmed start date.

After adding all meals, confirm what was planned with a brief summary.

## Step 7: Generate the Grocery List

Ask: "Want me to generate the grocery list now?"

If yes, run the generate-grocery-list skill inline — do not ask the user to invoke it separately. The meal plan is already set, so proceed directly through the grocery generation workflow.

## Step 8: Hand Off

After the grocery list is generated (or if the user skips it), close with:

> When you're done shopping, check off items in Paprika and then ask me to sync the grocery list to the pantry — that'll mark everything as in-stock for next week.
>
> During the week: **if you don't end up cooking a planned meal, remove it from the plan** — that's how we track what actually got cooked vs. what was skipped. Meals that stay on the plan are assumed cooked.

---

## Notes

- **Never add meals to the plan without user confirmation.** Always present the proposed week first.
- **Respect existing plan entries.** Always ask before clearing — never delete existing meals without explicit user confirmation.
- **Lunches and breakfasts are optional.** Default to dinners only; only ask about other meal types if the user brings it up.
- **Meal type names in the API:** use lowercase `"dinner"`, `"lunch"`, `"breakfast"` — the API does not accept title case.
