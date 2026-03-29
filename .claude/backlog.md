---
name: Feature backlog
description: Planned work for paprika-3-mcp — what to build next and what to verify
type: project
---

# paprika-3-mcp Backlog

## 1. ~~Fix Woodman's East aisle ordering~~ — OBSOLETE

Paprika handles aisle ordering natively via the configured aisle list. Woodmans-specific ordering is no longer needed. Dropped.

---

## 2. Pantry write capability — DONE ✅

`add_pantry_item` and `update_pantry_item` shipped (commit 5b4e1a5).

---

## 3. Grocery list write — DONE ✅

`add_grocery_item` working as of 2026-03-28 via `POST /api/v1/sync/groceries/` with Basic Auth + gzip multipart. Aisle auto-assigned from ingredient name via aisle map.

---

## 4. Aisle system extended to pantry — DONE ✅ (2026-03-28)

- `add_pantry_item` now auto-assigns aisle via the aisle map on create
- New `setup_pantry_aisles` tool (with `dry_run` flag) bulk-assigns aisles to existing pantry items
- Aisle map (`aisles/woodmans_east.json`) expanded with ~40 new entries covering snacks, frozen, beverages, condiments, and pantry staples
- Pantry fully populated (88 items) with aisles assigned and backfilled (2026-03-28)

**Still needed (backlog):**
- Remove hardcoded `woodmans_east.json` dependency
- Build LLM-driven or user-configured ingredient→aisle mapping that works for any store

---

## 5. Grocery list generation from meal plan + pantry — DONE ✅ (2026-03-28)

**Delivered as a skill** (`.claude/skills/generate-grocery-list/SKILL.md`). LLM-driven workflow using existing MCP tools. Includes pantry matching, deduplication, general 25-50% buffer, and output summary. Household-specific rules (substitutions, quantity multipliers) are loaded at runtime via `get_household_rules` — see #7.

Also delivered: `sync_grocery_list_to_pantry` MCP tool — moves all checked (purchased) grocery items into the pantry and removes them from the grocery list. See #10 below for details.

---

## 6. ~~Clean up debug/probe code~~ — DONE ✅ (commit 9a26d50)

---

## 7. Household rules system — DONE ✅ (2026-03-28)

- New `internal/rules` package: `Rule` struct, `Rules` slice, Load/Save/Upsert, `ToMarkdown()`
- Config file: `rules/household.json` seeded with substitute-venison and double-proteins rules
- Two new MCP tools: `get_household_rules` (no args, returns markdown) and `set_household_rule` (id, type, description, params JSON)
- Rules persisted to disk on every `set_household_rule` call; loaded at server startup via `-rules` flag
- Grocery generation skill updated to call `get_household_rules` at runtime — rules no longer hardcoded in prompt
- 7 unit tests, all passing

---

## 8. DRY refactor — DONE ✅ (2026-03-28, commit 8799aba)

- Extracted `gzipBytes`, `buildMultipartBody`, `newUID` helpers in `client.go` — all 5 save methods now use them
- `Recipe.asGzip()` delegates to `gzipBytes`
- `setupWoodmansAisles` and `setupPantryAisles` merged into shared `applyAisleAssignments` core
- Normalized HTTP error checks to `>= 400` across data endpoints
- Added `CLAUDE.md` at repo root with architecture overview, tool list, API quirks, coding principles

---

## 9. CLAUDE.md — DONE ✅ (2026-03-28, commit 8799aba)

---

## 10. sync_grocery_list_to_pantry — DONE ✅ (2026-03-28)

New MCP tool: reads all checked (purchased=true) grocery items, upserts each into pantry (update existing to in_stock=true, or add new), then removes them from the grocery list via a soft-delete (deleted=true via V1 sync endpoint).

DeleteGroceryItem verified working via integration test (commit f2d63b6).

---

## 13. Test suite + UpdateGroceryItem bug fix — DONE ✅ (2026-03-28)

- Added `TestSaveAndDeleteGroceryItem` — verifies full grocery create/delete round-trip via V1 sync
- Added `TestPantryClient` — verifies ListPantryItems and SavePantryItem create/update
- Added `.claude/testing.md` — documents how to run tests, coverage map, known gaps, deploy workflow, and manual smoke check checklist
- **Bug fixed:** `UpdateGroceryItem` was using V2 single-item endpoint which 404s on real items; switched to V1 upsert (same as SaveGroceryItem). Affected `setup_woodmans_aisles` and `update_grocery_item_aisle`.
- All 14 tests passing (9 unit, 5 integration)

---

## 11. Weekly planning workflow — NEXT

**Goal:** A single "plan the week" skill that walks the full loop: pantry review → meal planning → grocery list generation.

**Current state:** The tools exist in isolation — pantry, meal plan, grocery list, and the grocery generation skill all work independently. There's no opinionated workflow that connects them.

**Full flow:**
1. **Start from the pantry** — surface what's in stock, flag proteins and key ingredients that should anchor the meal plan, flag items expiring soon.
2. **Choose a planning level** (see below) — determines how meals are selected.
3. **Plan the week** — populate `add_meal_to_plan` for the week's dinners (and lunches/breakfasts if desired).
4. **Generate the grocery list** — invoke the grocery generation skill inline: cross-reference pantry, apply household rules, add items.
5. **Hand off** — user shops, checks off items, then runs `sync_grocery_list_to_pantry`.

**Planning levels:**

- **Tried and True** — pull from the existing Paprika recipe library only. Prioritize recipes that use in-stock proteins and pantry staples. User confirms or swaps suggestions.

- **Try Something New** — same as above but reserve 1-2 slots for recipes the user hasn't made recently (low rating or never rated). May prompt the user with a few options to pick from.

- **Let AI Take the Wheel** — Claude generates a full week of meals from scratch. Web-searches for recipes based on the season, what's in the pantry, and household preferences. Prompts the user with questions (dietary goals, cuisines of interest, how adventurous, etc.), then creates the recipes in Paprika via `create_paprika_recipe` and plans the week.

**Settled:**
- Planning level chosen via conversational prompting (not as a skill argument). The skill opens with a pantry snapshot, then asks how the user wants to plan.
- Should the skill always cover Mon–Sun, or ask for the date range? TBD.

---

## 12. Skill layer over MCP tools — ARCHITECTURE

**Goal:** Most MCP tools should have a corresponding skill that provides context, guardrails, and workflow guidance — so Claude isn't using raw tools cold.

**Why:** Raw MCP tools are powerful but dumb. A skill wrapping `sync_grocery_list_to_pantry`, for example, can remind Claude to confirm the item count before executing, summarize what changed, and suggest next steps. Without a skill, Claude has to reconstruct that context every time.

**Candidates for skill wrapping:**
- `sync_grocery_list_to_pantry` — confirm checked items before syncing, summarize results, prompt to run grocery generation if list is now empty
- `setup_woodmans_aisles` / `setup_pantry_aisles` — guide when to run these and what to do with unknowns
- `add_pantry_item` / `update_pantry_item` — probably only needed if there's a bulk "restock pantry after a big shop" workflow
- `get_pantry` — a pantry review skill that surfaces actionable insights (low stock, out of stock proteins, expiring items) rather than just dumping the table

**Not every tool needs a skill** — `get_recipe`, `list_recipes`, `add_meal_to_plan` are low enough stakes that raw tool use is fine. The bar for a skill is: does the tool benefit from workflow context, confirmation steps, or follow-up suggestions?

---

## How to approach this work

1. ~~**#1**~~ — obsolete
2. ~~**#2**~~ — DONE ✅
3. ~~**#3**~~ — DONE ✅
4. ~~**#4**~~ — DONE ✅
5. ~~**#5 (grocery generation skill)**~~ — DONE ✅
6. ~~**#6**~~ — DONE ✅
7. ~~**#7 (household rules)**~~ — DONE ✅
8. ~~**#8 (DRY refactor)**~~ — DONE ✅
9. ~~**#9 (CLAUDE.md)**~~ — DONE ✅
10. ~~**#10 (sync_grocery_list_to_pantry)**~~ — DONE ✅
11. **#11 (plan the week skill)** — next up
12. **#12 (skill layer over MCP tools)** — architecture work, do alongside #11
13. ~~**#13 (test suite + UpdateGroceryItem fix)**~~ — DONE ✅
