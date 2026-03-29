# Testing Plan

## How to run

### Unit tests (no credentials)

```bash
go test ./internal/aisles/... -v
```

These run offline and cover the aisle map lookup logic. Always run before committing.

### Integration tests (require Paprika credentials)

```bash
source ~/.paprika-env && \
  go test ./internal/paprika/... -v -timeout 120s
```

These hit the real Paprika API. Always run with real credentials — never skip. Run after any change to `internal/paprika/client.go` or when adding a new API operation.

### All tests

```bash
source ~/.paprika-env && \
  go test ./... -v -timeout 120s
```

---

## What each test covers

### `internal/aisles` (unit)

| Test | Covers |
|---|---|
| `TestLookupExactMatch` | Exact key lookup |
| `TestLookupCaseInsensitive` | Case folding |
| `TestLookupPartialMatch` | Substring match when no exact key |
| `TestLookupExactBeforePartial` | Exact match wins over partial |
| `TestLookupLongestPartialMatchWins` | Longest partial key wins (e.g. "tuna steak" beats "tuna") |
| `TestLookupLongestPartialFrozen` | Longer key wins for frozen variants |
| `TestLookupMiss` | Returns false when no match |
| `TestLoadAndSave` | JSON round-trip via Load/Save |
| `TestLoadMissingFile` | Error on missing file |

### `internal/rules` (unit)

| Test | Covers |
|---|---|
| `TestUpsertAdd` | Adding a new rule |
| `TestUpsertReplace` | Replacing an existing rule by ID |
| `TestUpsertPreservesOrder` | Order preserved on replace |
| `TestLoadSaveRoundTrip` | JSON save/load round-trip |
| `TestLoadMissingFileReturnsEmpty` | Missing file returns empty Rules |
| `TestToMarkdownEmpty` | Empty rules markdown output |
| `TestStapleRuleRoundTrip` | Staple rule type saves/loads with ingredient param |
| `TestToMarkdownContainsRuleID` | Markdown output includes rule ID, type, description |

### `internal/paprika` (integration)

| Test | Covers |
|---|---|
| `TestClient` | Recipe create, get, update, delete, list (full CRUD) |
| `TestGroceryClient` | ListGroceryItems, UpdateGroceryItem (aisle round-trip) |
| `TestSaveAndDeleteGroceryItem` | SaveGroceryItem (create), DeleteGroceryItem (soft-delete via `deleted=true` on V1 sync) |
| `TestUncheckGroceryItem` | SaveGroceryItem (purchased=true), UpdateGroceryItem (uncheck → purchased=false), item stays on list, cleanup |
| `TestMealPlanClient` | ListMealPlanEntries |
| `TestPantryClient` | ListPantryItems, SavePantryItem create + update round-trip |

---

## Known gaps

- **MCP server handlers** (`internal/mcpserver/server.go`) — not unit tested. Handler logic is thin by design (validate → call client → format output), so integration tests on the client layer cover the critical path. If handler logic grows, add handler-level tests.
- **`TestPantryClient` leaves a test item** — no `DeletePantryItem` method exists. Test item persists in the pantry with a timestamped name. Remove manually from the app, or add `DeletePantryItem` if this becomes painful.
- **`sync_grocery_list_to_pantry` tool** (deprecated) — covered indirectly. The skill-driven flow uses `uncheck_grocery_items` and `delete_grocery_items`, both backed by client methods that are individually tested.

---

## After code changes: deploy and smoke test

After any change to Go code:

```bash
go build ./...                        # verify compile
source ~/.paprika-env && \
  go test ./... -v -timeout 120s      # run all tests including integration
go install ./cmd/paprika-3-mcp/       # install updated binary
# then run /mcp in the Claude Code session to reconnect
```

Manual smoke checks to run in Claude after reconnecting:

| Check | Tool | What to verify |
|---|---|---|
| Recipes load | `list_recipes` | Returns recipe list, no error |
| Recipe fetch | `get_recipe` | Returns full ingredients/directions for a known recipe |
| Grocery list | `get_grocery_list` | Returns current list, aisles present |
| Pantry | `get_pantry` | Returns pantry items with in-stock status |
| Meal plan | `get_meal_plan` | Returns current week's plan |
| Aisle map | `setup_woodmans_aisles` (dry_run=true) | Shows proposals without writing |
| Pantry aisles | `setup_pantry_aisles` (dry_run=true) | Shows proposals without writing |
| Sync (careful) | `sync_grocery_list_to_pantry` | Only run if there are purchased items you intend to move |

---

## When to run what

| Situation | Run |
|---|---|
| Before every commit | Unit tests (`./internal/aisles/...`) |
| After editing `client.go` | Integration tests (`./internal/paprika/...`) |
| After editing `server.go` | Integration tests + manual smoke checks |
| After `go install` / MCP reconnect | Manual smoke checks |
| Adding a new API endpoint | Integration test for that endpoint before shipping |
