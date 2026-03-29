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
| 12 | Skill layer: `sync-grocery-list` + `setup-aisles` skills | 1a0c9ea |
| 13 | Test suite + `UpdateGroceryItem` bug fix (14 tests) | f2d63b6 |
| 20b | Staple items: `uncheck_grocery_items`, `delete_grocery_items`, staple rule type | 1a0c9ea |
| 18 | `remove_meal_from_plan` — soft-delete via `deleted=true` on V1 sync endpoint | 4967c6e |
| 21 | Last Prepared date — derived from meal plan history, shown in `list_recipes` + `get_recipe` | 65c3dc3 |
| — | Security: scrub credentials from repo history, use `~/.paprika-env` | 1d82d5b |

---

## Active / Planned

### ~~#14 — done~~ → shipped as `close-cycle` skill

**Design:** `close-cycle` is the one canonical end-of-cycle sync — deplete from cooked meals → restock from shopping → advance `last_sync_date`. Always runs in that order. `last_sync_date` stored as a household rule (type: "sync", id: "last-sync-date"). Meals that stay on the plan = cooked; deleted via `remove_meal_from_plan` = skipped. First run with no `last_sync_date`: skips depletion entirely (clean start).

**Notes for productionalization:** `last_sync_date` is a household rule (global per-account). In a multi-user/multi-household system this would need to be per-household.

### #22 — Quantity matching: recipe amounts vs. typical purchase units
During `close-cycle` depletion, recipe quantities ("2 tbsp soy sauce") frequently don't map to purchase units ("1 bottle"). Need a system that learns typical purchase quantities per ingredient and uses them to make smarter depletion decisions — e.g., "2 tbsp soy sauce from a 10oz bottle = not depleted, bottle stays in-stock." Options: household rules with per-ingredient purchase units, or a learned quantity map built from grocery history. **Design first, build second.**

| 23 | `sync-grocery-list` guard rail — note added that it does NOT advance `last_sync_date` | 3386d86 |
| 24 | `delete_pantry_item` — soft-delete via `deleted=true` on V1 sync endpoint; pantry hygiene in `close-cycle` | TBD |

### #15 — "What can I make tonight?"
Lightweight skill: `get_pantry` + `list_recipes`, match in-stock ingredients to recipes, surface top options. No meal plan write required.

### #16 — Scheduled weekly planning trigger
Wire plan-the-week to a Monday morning cron via the `schedule` skill. Verify how scheduled remote agents handle conversational skills.

### #17 — Pantry review skill
Wraps `get_pantry` + `get_household_rules`. Flags out-of-stock proteins, stale items, low-stock staples — actionable insights rather than a data dump.

### ~~#18 — done~~

### #19 — Recipe deletion (`delete_paprika_recipe`)
`DeleteRecipe` client method already exists (sets `in_trash: true`). Just needs an MCP tool wired to it. Note: only moves to trash; full delete requires in-app.

### #20 — Favorites via star ratings
`rating` field likely already in recipe objects — verify. Expose in `list_recipes`, bias planner toward 4–5 star recipes. Also feed into #15.

### ~~#21 — done~~
