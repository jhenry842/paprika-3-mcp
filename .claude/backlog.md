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
| 23b | `sync-grocery-list` Step 0 guard — disambiguates mid-cycle top-up from end-of-cycle close | 8e724b7 |
| 26 | `get_grocery_list` now exposes UIDs — required for delete/uncheck flows | 39add92 |
| 28 | CLAUDE.md meal date format corrected — `add_meal_to_plan` takes `"YYYY-MM-DD"` not `"YYYY-MM-DD 00:00:00"` | 73ac989 |
| 19 | `delete_paprika_recipe` tool — wires existing `DeleteRecipe` client method | 73ac989 |
| 27 | `plan-the-week` date format + arithmetic bug — fix `"YYYY-MM-DD 00:00:00"` → `"YYYY-MM-DD"`, explicit offset arithmetic | 73ac989 |
| 15 | `what-can-i-make` skill — pantry → recipe match, ranked by rating + recency, optional meal plan write | 73ac989 |
| 17 | `pantry-review` skill — on-demand mid-cycle health check: proteins, staples, low-stock flags | — |
| 30 | Purchase history learning — `grocery-history` rule records purchased ingredients per cycle; `generate-grocery-list` surfaces recurring items not on current list | — |
| — | Array param JSON string fallback — `delete_grocery_items`, `uncheck_grocery_items`, `update_grocery_item_aisle` | cf17637, ea9672f |
| — | `get_meal_plan` now exposes entry UIDs and recipe UIDs — required for `remove_meal_from_plan` | c7135cf |
| — | Security: scrub credentials from repo history, use `~/.paprika-env` | 1d82d5b |
| — | Boolean param string fallback — `setup_woodmans_aisles`, `setup_pantry_aisles` `dry_run` param arrived as `"false"` string | 0d8e9bc |
| — | Aisle map expansion — 29 new entries; corrected ranch/mayo/tartar sauce out of Oils; int'l sauces to International Cuisine | 3cd4992 |
| — | `setup_woodmans_aisles` + `setup_pantry_aisles` collapsed into single `setup_aisles(target)` tool; add `AisleUID` to `PantryItem`; aisle shown in `get_pantry` | 4810f70 |

---

## Active / Planned

### #16 — Scheduled planning trigger
Wire plan-the-week to a recurring cron (Monday morning default, configurable) via the `schedule` skill. Verify how scheduled remote agents handle conversational skills.


### #22 — Quantity matching improvements
During `close-cycle` depletion, recipe quantities ("2 tbsp soy sauce") frequently don't map to purchase units ("1 bottle"). Need a system that learns typical purchase quantities per ingredient — e.g., "2 tbsp soy sauce from a 10oz bottle = bottle stays in-stock." Options: household rules with per-ingredient purchase units, or a learned quantity map from grocery history. **Design first, build second.**

**Productionalization note:** `last_sync_date` is currently a household rule (global per-account). Multi-household deployments would need per-household storage.

### #25 — Aisle map self-correction from manual Paprika changes
When generating a grocery list or running setup-aisles, compare each item's current aisle in Paprika against what the aisle map would assign. If Paprika has a non-empty aisle that differs from the map, treat the Paprika value as the ground truth and update the map. Manual user corrections are always improvements. Only trigger when current aisle is non-empty AND conflicts with the map value — never overwrite a blank.

### #29 — Pantry item age tracking
Track when each pantry item was last restocked so that close-cycle and plan-the-week can surface items that have been sitting a long time. Enables waste minimization: deprioritize buying more of something that's been in the pantry untouched for multiple cycles, and flag long-sitting perishables during pantry hygiene. Options: store a `restocked_date` field on pantry items (requires API support or a local sidecar), or derive age from `last_sync_date` history. **Design first — verify what the Paprika pantry item schema supports before building.**
