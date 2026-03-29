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
| 14 | `close-cycle` skill — deplete → restock → advance `last_sync_date` | 3386d86 |
| 18 | `remove_meal_from_plan` — soft-delete via `deleted=true` on V1 sync endpoint | 4967c6e |
| 20 | Star ratings in `list_recipes`; planner biases toward 4–5 stars; Last Prepared column | 65c3dc3 |
| 20b | Staple items: `uncheck_grocery_items`, `delete_grocery_items`, staple rule type | 1a0c9ea |
| 21 | Last Prepared date — derived from meal plan history, shown in `list_recipes` + `get_recipe` | 65c3dc3 |
| 23 | `sync-grocery-list` guard rail — does NOT advance `last_sync_date` | 3386d86 |
| 24 | `delete_pantry_item` — soft-delete + pantry hygiene step in `close-cycle` | b8ae268 |
| — | Security: scrub credentials from repo history, use `~/.paprika-env` | 1d82d5b |

---

## Active / Planned

### #15 — "What can I make tonight?"
Lightweight skill: `get_pantry` + `list_recipes`, match in-stock ingredients to recipes, surface top options. `list_recipes` now includes Last Prepared and rating — bias toward high-rated recipes not cooked recently. No meal plan write required.

### #16 — Scheduled planning trigger
Wire plan-the-week to a recurring cron (Monday morning default, configurable) via the `schedule` skill. Verify how scheduled remote agents handle conversational skills.

### #17 — Pantry review skill
Standalone pantry health check: `get_pantry` + `get_household_rules`. Flags out-of-stock proteins, low-stock staples. Distinct from the close-cycle hygiene step (which only runs at cycle close) — this is an on-demand mid-cycle check.

### #19 — Recipe deletion (`delete_paprika_recipe`)
`DeleteRecipe` client method already exists (sets `in_trash: true`). Just needs an MCP tool wired to it. Note: only moves to trash; full delete requires in-app.

### #22 — Quantity matching improvements
During `close-cycle` depletion, recipe quantities ("2 tbsp soy sauce") frequently don't map to purchase units ("1 bottle"). Need a system that learns typical purchase quantities per ingredient — e.g., "2 tbsp soy sauce from a 10oz bottle = bottle stays in-stock." Options: household rules with per-ingredient purchase units, or a learned quantity map from grocery history. **Design first, build second.**

**Productionalization note:** `last_sync_date` is currently a household rule (global per-account). Multi-household deployments would need per-household storage.
