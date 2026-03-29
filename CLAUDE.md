# paprika-3-mcp

An MCP (Model Context Protocol) server that gives Claude access to a Paprika 3 recipe app account. Claude can read recipes, manage the meal plan, manage the grocery list, and manage the pantry.

## Architecture

Three packages, each with a single responsibility:

| Package | Path | Does |
|---|---|---|
| `paprika` | `internal/paprika/` | Paprika API client — all HTTP, auth, types |
| `mcpserver` | `internal/mcpserver/` | MCP server — tool definitions and handlers |
| `aisles` | `internal/aisles/` | Aisle map — load/save/lookup ingredient→aisle |

Entry point: `cmd/paprika-3-mcp/main.go` — reads env vars, wires packages together.

The MCP server pre-loads all recipes as MCP resources at startup and refreshes on a ticker. Tool handlers are thin: validate input, call the paprika client, format output as markdown.

## Available Tools

| Tool | What it does |
|---|---|
| `list_recipes` | List all recipes with name, UID, categories, timing, rating |
| `get_recipe` | Fetch full recipe by UID |
| `create_paprika_recipe` | Save a new recipe |
| `update_paprika_recipe` | Update an existing recipe by UID |
| `get_meal_plan` | Fetch meal plan for a date range |
| `add_meal_to_plan` | Add a recipe to a meal slot (idempotent) |
| `get_grocery_list` | List all grocery items |
| `add_grocery_item` | Add an item to the grocery list |
| `update_grocery_item_aisle` | Set aisle on one or more grocery items |
| `setup_woodmans_aisles` | Bulk-assign aisles to grocery items from the aisle map |
| `get_pantry` | List all pantry items |
| `add_pantry_item` | Add an item to the pantry |
| `update_pantry_item` | Update quantity or in-stock status of a pantry item |
| `setup_pantry_aisles` | Bulk-assign aisles to pantry items from the aisle map |

## Build & Deploy

```bash
# Build and install the binary
go install ./cmd/paprika-3-mcp/

# After any code change, reinstall and reconnect MCP in Claude Code
go install ./cmd/paprika-3-mcp/
# then run /mcp in the Claude Code session to reconnect
```

The binary the MCP runner executes is the installed one — editing source does not take effect until you reinstall.

## Paprika API Quirks

- **Two API versions in use.** V2 uses Bearer token auth (set by the roundTripper middleware). V1 uses HTTP Basic Auth and is required for write operations on groceries, meals, and pantry.
- **All write operations** (V1 and V2) use gzip-compressed JSON wrapped in a multipart form upload — see `gzipBytes` and `buildMultipartBody` helpers in `client.go`.
- **Meal plan and grocery dates** must be in `"YYYY-MM-DD 00:00:00"` format, not bare `"YYYY-MM-DD"`.
- **V2 recipe saves** can return HTTP 200 with an error payload — always run `isErrorResponse` after reading the body on recipe write endpoints.
- **Never infer an API endpoint by pattern.** The V1/V2 split is inconsistent. Verify against the real API before adding any new endpoint.

## Coding Principles

**Keep it DRY.** Before copying code, extract a helper. The shared helpers for gzip, multipart, and UID generation exist because all five save methods need the same mechanics. When you add a new save method, use them.

**Thin handlers.** Tool handlers in `server.go` should validate input, call the client, and format output. Business logic belongs in the `paprika` package, not in handlers.

**No behavior changes in refactors.** When restructuring, tests must pass before and after. Do not sneak in API surface changes or new features in the same commit.

**One purpose per function.** If you can't describe what a function does in one sentence without "and", it's doing too much.

**Error checks on HTTP responses use `>= 400`** for data endpoints. The `!= http.StatusOK` pattern is only appropriate for simple single-status endpoints like login and notify.

**Verify before building.** Never assume an API endpoint exists or behaves a certain way based on patterns from other endpoints. The Paprika API is inconsistent. Test against the real API first.

**YAGNI.** Don't add configuration knobs, abstractions, or fallbacks for scenarios that don't exist yet. Build for what's needed now.
