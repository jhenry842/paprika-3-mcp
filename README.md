# paprika-3-mcp

A [Model Context Protocol (MCP)](https://modelcontextprotocol.io/introduction) server that gives Claude full access to your **Paprika 3** recipe app — recipes, meal planning, grocery lists, and pantry management — plus a skill layer for end-to-end meal cycle workflows.

### 🖼️ Example: Claude using the Paprika MCP server

<p align="center">
  <img src="docs/example.png" alt="MCP server running with Claude" />
</p>

## 🚀 Features

See anything missing? Open an issue on this repo to request a feature!

#### 📄 **Resources**

- Recipes ✅
- Recipe Photos 🚧

#### 🛠 **Tools**

**Recipes**
- `list_recipes` — List all recipes with name, UID, categories, prep/cook time, star rating, and last prepared date
- `get_recipe` — Fetch full recipe content (ingredients, directions, notes, last prepared date) by UID
- `create_paprika_recipe` — Save a new recipe to your Paprika app
- `update_paprika_recipe` — Modify an existing recipe

**Meal Plan**
- `get_meal_plan` — Fetch the meal plan for a date range
- `add_meal_to_plan` — Add a recipe to the meal plan (idempotent)
- `remove_meal_from_plan` — Remove a meal plan entry by UID; used to mark a meal as skipped (not cooked)

**Grocery List**
- `get_grocery_list` — Fetch all grocery list items with UID, aisle, quantity, and purchased status (UIDs required for delete/uncheck)
- `add_grocery_item` — Add an item to the grocery list
- `update_grocery_item_aisle` — Set the aisle label on one or more grocery items
- `setup_woodmans_aisles` — Bulk-assign Woodman's East aisles to all grocery list items
- `uncheck_grocery_items` — Set purchased=false on items by UID (for staples that stay on the list)
- `delete_grocery_items` — Remove items from the grocery list by UID

**Pantry**
- `get_pantry` — Fetch all pantry items with ingredient, quantity, aisle, and in-stock status
- `add_pantry_item` — Add a new item to the pantry
- `update_pantry_item` — Update quantity or in-stock status of a pantry item
- `delete_pantry_item` — Permanently remove a pantry item
- `setup_pantry_aisles` — Bulk-assign Woodman's East aisles to all pantry items

**Household Rules**
- `get_household_rules` — Fetch all household rules (staples, substitutions, sync anchors)
- `set_household_rule` — Create or update a household rule

#### 🤖 **Claude Code Skills** *(requires Claude Code)*

Skills are conversational workflows that orchestrate multiple MCP tools in sequence. They're invoked by natural language and designed for the full meal planning cycle.

| Skill | How to invoke | What it does |
|---|---|---|
| `plan-the-week` | "plan the week", "what should we eat this week" | Pantry review → meal selection → meal plan population → grocery list generation |
| `generate-grocery-list` | "generate grocery list", "what do I need to buy" | Turn a meal plan into a shopping list, cross-referenced against your pantry |
| `sync-grocery-list` | "I'm done shopping", "sync the grocery list" | Post-shopping restock: sync checked items to pantry, handle staples |
| `setup-aisles` | "set up aisles", "fix the aisles" | Bulk-assign Woodman's East aisles to the grocery list and/or pantry |
| `close-cycle` | "close the cycle", "I'm done cooking", "full sync" | End-of-cycle sync: deplete pantry from cooked meals → pantry hygiene → restock from shopping |

## 🔄 The Meal Cycle

The intended weekly (or bi-weekly, or monthly) workflow:

```
plan-the-week
  → select meals, build meal plan, generate grocery list

[cook during the cycle]
  → if you skip a planned meal: say "remove [meal] from the plan"

close-cycle
  → depletes pantry based on what you cooked
  → surfaces stale pantry items for cleanup
  → restocks pantry from your shopping
  → records the sync date for the next cycle
```

**`close-cycle` is the canonical end-of-cycle operation.** It's the only workflow that advances the sync date. Use `sync-grocery-list` only for mid-cycle top-up shops.

## ⚙️ Prerequisites

- ✅ A Mac, Linux, or Windows system
- ✅ [Paprika 3](https://www.paprikaapp.com/) installed with cloud sync enabled
- ✅ Your Paprika 3 **username and password**
- ✅ Claude or any LLM client with **MCP tool support** enabled
- ✅ [Claude Code](https://claude.ai/code) for skills (optional but recommended)

## 🛠 Installation

You can download a prebuilt binary from the [Releases](https://github.com/soggycactus/paprika-3-mcp/releases) page.

### Build from source

```bash
git clone https://github.com/soggycactus/paprika-3-mcp
cd paprika-3-mcp
go install ./cmd/paprika-3-mcp/
```

### 🍎 macOS (via Homebrew)

```bash
brew tap soggycactus/tap
brew install paprika-3-mcp
```

### 🐧 Linux / 🪟 Windows

1. Go to the [latest release](https://github.com/soggycactus/paprika-3-mcp/releases).
2. Download the archive for your OS and architecture.
3. Extract and move the binary to a directory in your `$PATH`.

### ✅ Test the installation

```bash
paprika-3-mcp --version
```

## 🤖 Setting up Claude

If you haven't set up MCP before, [read how to install the Claude Desktop client and configure an MCP server](https://modelcontextprotocol.io/quickstart/user).

### Claude Code (recommended)

The repo includes a `.mcp.json` that Claude Code picks up automatically. Set your credentials as environment variables:

```json
{
  "mcpServers": {
    "paprika": {
      "command": "paprika-3-mcp",
      "args": [
        "--refresh-interval", "5m",
        "--aisle-map", "/path/to/aisles/woodmans_east.json",
        "--grocery-list", ""
      ],
      "env": {
        "PAPRIKA_USERNAME": "you@example.com",
        "PAPRIKA_PASSWORD": "yourpassword"
      }
    }
  }
}
```

To use the skill layer, clone this repo and open it in Claude Code — the skills in `.claude/skills/` are picked up automatically.

### Claude Desktop

Add an entry to the `mcpServers` section of your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "paprika": {
      "command": "paprika-3-mcp",
      "env": {
        "PAPRIKA_USERNAME": "you@example.com",
        "PAPRIKA_PASSWORD": "yourpassword"
      }
    }
  }
}
```

Restart Claude and you should see the MCP server tools after clicking on the hammer icon.

## ⚙️ Configuration flags

| Flag | Default | Description |
|------|---------|-------------|
| `--refresh-interval` | `5m` | How often to refresh the recipe resource cache |
| `--aisle-map` | `aisles/woodmans_east.json` | Path to aisle map JSON file (used by `setup_woodmans_aisles` and `setup_pantry_aisles`) |
| `--grocery-list` | *(first list)* | Default grocery list name; empty string uses the first list |
| `--version` | — | Print version and exit |

Credentials are read from `PAPRIKA_USERNAME` and `PAPRIKA_PASSWORD` environment variables.

## 📄 License

This project is open source under the [MIT License](./LICENSE) © 2025 [Lucas Stephens](https://github.com/soggycactus).

---

#### 🗂 Miscellaneous

##### 📄 Where can I see the server logs?

The MCP server writes structured logs using Go's `slog`. Log files are created based on your operating system:

| Operating System | Log File Path                             |
| ---------------- | ----------------------------------------- |
| macOS            | `~/Library/Logs/paprika-3-mcp/server.log` |
| Linux            | `/var/log/paprika-3-mcp/server.log`       |
| Windows          | `%APPDATA%\paprika-3-mcp\server.log`      |
| Other / Unknown  | `/tmp/paprika-3-mcp/server.log`           |

> 💡 Logs are rotated automatically at 100MB, with only 5 backup files kept and a 10-day retention window.
