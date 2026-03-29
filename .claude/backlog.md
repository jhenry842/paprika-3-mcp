---
name: Feature backlog
description: Planned work for paprika-3-mcp — what to build next and what to verify
type: project
---

# paprika-3-mcp Backlog

## Done ✅

| # | Item | Commit |
|---|---|---|
| 1 | Fix Woodman's aisle ordering | OBSOLETE — Paprika handles natively |
| 2 | Pantry write (`add_pantry_item`, `update_pantry_item`) | 5b4e1a5 |
| 3 | Grocery list write (`add_grocery_item`) | 9a26d50 |
| 4 | Aisle system extended to pantry | c89124b |
| 5 | Grocery generation skill | c058b46 |
| 6 | Clean up debug/probe code | 9a26d50 |
| 7 | Household rules system (`get/set_household_rule`) | 5fd5ef6 |
| 8 | DRY refactor (shared helpers, normalized error checks) | 8799aba |
| 9 | CLAUDE.md | 8799aba |
| 10 | `sync_grocery_list_to_pantry` | c058b46 |
| 11 | plan-the-week skill | dd02a7e |
| 12 | Skill layer: `sync-grocery-list` + `setup-aisles` skills | 2026-03-29 |
| 13 | Test suite + `UpdateGroceryItem` bug fix (14 tests) | f2d63b6 |
| 20b | Staple items: `uncheck_grocery_items`, `delete_grocery_items`, staple rule type | 2026-03-29 |

---

## Active / Planned

### #14 — Pantry depletion tracking ⚠️ most important gap
Nothing marks ingredients as consumed when you cook — pantry only grows. Options: conversational "anything you've run out of?" check at plan-the-week time, or a "meal cooked" step that marks recipe ingredients as depleted. Pantry accuracy is load-bearing for grocery generation.

### #15 — "What can I make tonight?"
Lightweight skill: `get_pantry` + `list_recipes`, match in-stock ingredients to recipes, surface top options. No meal plan write required.

### #16 — Scheduled weekly planning trigger
Wire plan-the-week to a Monday morning cron via the `schedule` skill. Verify how scheduled remote agents handle conversational skills.

### #17 — Pantry review skill
Wraps `get_pantry` + `get_household_rules`. Flags out-of-stock proteins, stale items, low-stock staples — actionable insights rather than a data dump.

### #18 — Meal plan deletion (`remove_meal_from_plan`)
New `DeleteMealPlanEntry` client method — soft-delete pattern (set `deleted=true`, POST to `/api/v1/sync/meals/`). **Verify against real API first.** Unblocks plan-the-week "clear and restart" flow.

### #19 — Recipe deletion (`delete_paprika_recipe`)
`DeleteRecipe` client method already exists (sets `in_trash: true`). Just needs an MCP tool wired to it. Note: only moves to trash; full delete requires in-app.

### #20 — Favorites via star ratings
`rating` field likely already in recipe objects — verify. Expose in `list_recipes`, bias planner toward 4–5 star recipes. Also feed into #15.

### #21 — Cook log and history
`last_cooked_on` exists on recipes — unclear if editable override or true timestamp. Investigate whether a cook history API endpoint exists (**verify before building**). Cook frequency + star rating = stronger "favorite" signal than rating alone.
