# paprika-3-mcp

An MCP (Model Context Protocol) server that gives Claude access to a Paprika 3 recipe app account. Claude can read recipes, manage the meal plan, manage the grocery list, manage the pantry, and run full meal-cycle workflows via a skill layer.

## Architecture

Three packages, each with a single responsibility:

| Package | Path | Does |
|---|---|---|
| `paprika` | `internal/paprika/` | Paprika API client — all HTTP, auth, types |
| `mcpserver` | `internal/mcpserver/` | MCP server — tool definitions and handlers |
| `aisles` | `internal/aisles/` | Aisle map — load/save/lookup ingredient→aisle |

Entry point: `cmd/paprika-3-mcp/main.go` — reads env vars, wires packages together.

The MCP server pre-loads all recipes as MCP resources at startup and refreshes on a ticker. Tool handlers are thin: validate input, call the paprika client, format output as markdown.

Skills live in `.claude/skills/` and are Claude Code skill files — they orchestrate sequences of MCP tool calls conversationally. They are not compiled; they run as Claude prompts.

## Available Tools

### Recipes

| Tool | What it does |
|---|---|
| `list_recipes` | List all recipes with name, UID, categories, timing, star rating, and last prepared date |
| `get_recipe` | Fetch full recipe by UID — includes last prepared date in the Details section |
| `create_paprika_recipe` | Save a new recipe |
| `update_paprika_recipe` | Update an existing recipe by UID |

### Meal Plan

| Tool | What it does |
|---|---|
| `get_meal_plan` | Fetch meal plan for a date range |
| `add_meal_to_plan` | Add a recipe to a meal slot (idempotent) |
| `remove_meal_from_plan` | Soft-delete a meal plan entry by UID — used to mark a meal as not cooked |

### Grocery List

| Tool | What it does |
|---|---|
| `get_grocery_list` | List all grocery items with UID, aisle, quantity, and purchased status — UIDs required for delete/uncheck |
| `add_grocery_item` | Add an item to the grocery list |
| `update_grocery_item_aisle` | Set aisle on one or more grocery items |
| `setup_aisles` | Bulk-assign aisles from the aisle map — `target`: `"grocery"` (default), `"pantry"`, or `"both"` |
| `uncheck_grocery_items` | Set purchased=false on items by UID (keeps staples on list) |
| `delete_grocery_items` | Remove items from the grocery list by UID |
| `sync_grocery_list_to_pantry` | ⚠️ Deprecated — use the `sync-grocery-list` skill instead |

### Pantry

| Tool | What it does |
|---|---|
| `get_pantry` | List all pantry items with ingredient, quantity, aisle, and in-stock status |
| `add_pantry_item` | Add a new item to the pantry |
| `update_pantry_item` | Update quantity or in-stock status of a pantry item |
| `delete_pantry_item` | Permanently remove a pantry item (soft-delete via deleted=true) |

### Household Rules

| Tool | What it does |
|---|---|
| `get_household_rules` | Fetch all household rules (staples, substitutions, sync anchors, etc.) |
| `set_household_rule` | Create or update a household rule by ID |

Household rules are a typed key-value store persisted in `rules/household.json`. Current rule types:
- `staple` — ingredient kept on the grocery list after shopping (unchecked, not deleted)
- `substitution` — swap one ingredient for another during grocery generation
- `sync` — system-managed anchor dates (e.g., `last-sync-date`)

## Skills

Skills live in `.claude/skills/` and are invoked by Claude Code users. Each skill is a markdown prompt that orchestrates MCP tool calls conversationally.

| Skill | Trigger phrases | What it does |
|---|---|---|
| `plan-the-week` | "plan the week", "what should we eat" | Pantry review → meal selection → meal plan → grocery list |
| `generate-grocery-list` | "generate grocery list", "what do I need to buy" | Turn meal plan into a shopping list, cross-referenced against pantry |
| `sync-grocery-list` | "I'm done shopping", "sync the grocery list" | Post-shopping restock: sync checked items to pantry, uncheck staples, delete non-staples |
| `setup-aisles` | "set up aisles", "fix the aisles" | Bulk-assign Woodman's East aisles to grocery list and/or pantry |
| `close-cycle` | "close the cycle", "I'm done cooking", "full sync" | End-of-cycle canonical sync: deplete pantry from cooked meals → pantry hygiene → restock from shopping → advance `last_sync_date` |

### The Meal Cycle

The intended flow for each meal planning cycle:

1. **`plan-the-week`** — select meals, populate meal plan, generate grocery list
2. *(cook meals during the cycle; delete any you skip via `remove_meal_from_plan`)*
3. **`close-cycle`** — deplete pantry from cooked meals, surface stale pantry items, restock from shopping, advance `last_sync_date`

`last_sync_date` is stored as a household rule (`id: "last-sync-date"`). It marks the start of the next depletion window. On first run (no rule present), depletion is skipped entirely for a clean start.

**Important:** `sync-grocery-list` is for mid-cycle top-up shops only. It does NOT deplete the pantry and does NOT advance `last_sync_date`. Use `close-cycle` for end-of-cycle syncs.

### Last Prepared Date

`list_recipes` and `get_recipe` include a **Last Prepared** date, derived from meal plan history. Meals that stay on the plan were cooked; meals deleted via `remove_meal_from_plan` were skipped. The plan-the-week skill uses Last Prepared to deprioritize recently cooked recipes.

## Build & Deploy

```bash
# Build and install the binary
go install ./cmd/paprika-3-mcp/

# After any code change, reinstall and reconnect MCP in Claude Code
go install ./cmd/paprika-3-mcp/
# then run /mcp in the Claude Code session to reconnect
```

The binary the MCP runner executes is the installed one — editing source does not take effect until you reinstall.

## Testing

Integration tests require real Paprika credentials:

```bash
source ~/.paprika-env  # sets PAPRIKA_USERNAME and PAPRIKA_PASSWORD
go test ./... -timeout 120s
```

Tests hit the real API and clean up after themselves. Never skip integration tests — mock divergence has caused prod issues before.

## Paprika API Quirks

- **Two API versions in use.** V2 uses Bearer token auth (set by the roundTripper middleware). V1 uses HTTP Basic Auth and is required for all write operations (groceries, meals, pantry).
- **All write operations** (V1 and V2) use gzip-compressed JSON wrapped in a multipart form upload — see `gzipBytes` and `buildMultipartBody` helpers in `client.go`.
- **Meal plan dates** for `add_meal_to_plan` must be bare `"YYYY-MM-DD"`. The `"YYYY-MM-DD 00:00:00"` format is rejected by the tool.
- **V2 recipe saves** can return HTTP 200 with an error payload — always run `isErrorResponse` after reading the body on recipe write endpoints.
- **Soft deletes everywhere.** Groceries, meals, and pantry items are deleted by setting `deleted: true` and POSTing to the same V1 sync endpoint. There is no DELETE HTTP method.
- **Meal plan history persists.** The `/api/v2/sync/meals/` endpoint returns all historical entries. Last Prepared is derived by finding the most recent past entry per recipe UID.
- **Never infer an API endpoint by pattern.** The V1/V2 split is inconsistent. Verify against the real API before adding any new endpoint.

## Coding Principles

**Keep it DRY.** Before copying code, extract a helper. The shared helpers for gzip, multipart, and UID generation exist because all save methods need the same mechanics. When you add a new save method, use them.

**Thin handlers.** Tool handlers in `server.go` should validate input, call the client, and format output. Business logic belongs in the `paprika` package, not in handlers.

**No behavior changes in refactors.** When restructuring, tests must pass before and after. Do not sneak in API surface changes or new features in the same commit.

**One purpose per function.** If you can't describe what a function does in one sentence without "and", it's doing too much.

**Error checks on HTTP responses use `>= 400`** for data endpoints. The `!= http.StatusOK` pattern is only appropriate for simple single-status endpoints like login and notify.

**Verify before building.** Never assume an API endpoint exists or behaves a certain way based on patterns from other endpoints. The Paprika API is inconsistent. Test against the real API first.

**YAGNI.** Don't add configuration knobs, abstractions, or fallbacks for scenarios that don't exist yet. Build for what's needed now.

**Array parameters need JSON encoding.** Tools that accept array parameters (`delete_grocery_items`, `uncheck_grocery_items`, `update_grocery_item_aisle`) receive the value as a JSON string when called by Claude. Handlers must attempt `json.Unmarshal` as a fallback if the direct `.([]interface{})` assertion fails.
