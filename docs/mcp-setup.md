# MCP Setup — paprika-3-mcp (fork)

## Build

```bash
cd /home/john-henry-erikson/code/cooking
go build -o bin/paprika-3-mcp ./cmd/paprika-3-mcp/
```

## Configure Claude Code

Add to `~/.claude/settings.json` under `mcpServers`:

```json
{
  "mcpServers": {
    "paprika-3": {
      "command": "/home/john-henry-erikson/code/cooking/bin/paprika-3-mcp",
      "args": [
        "--refresh-interval", "5m",
        "--aisle-map", "/home/john-henry-erikson/code/cooking/aisles/woodmans_east.json",
        "--grocery-list", ""
      ],
      "env": {
        "PAPRIKA_USERNAME": "your@email.com",
        "PAPRIKA_PASSWORD": "your-unique-paprika-password",
        "TZ": "UTC"
      }
    }
  }
}
```

> **Path note:** The `command` and `--aisle-map` paths above are hardcoded to `/home/john-henry-erikson/code/cooking`. Update them if the repo is ever moved or cloned to a different location.

## Secure the config file

```bash
chmod 600 ~/.claude/settings.json
```

## Verify

Restart Claude Code. Ask: "List my Paprika recipes." You should see recipe resources available.

## Tools available

| Tool | Purpose |
|---|---|
| `create_paprika_recipe` | Add a new recipe |
| `update_paprika_recipe` | Edit an existing recipe (categories, rating, etc.) |
| `get_grocery_list` | View grocery list items and current aisle assignments |
| `update_grocery_item_aisle` | Manually update aisle labels |
| `setup_woodmans_aisles` | One-time fix of all aisle assignments to Woodman's East layout |
| `get_meal_plan` | View the meal plan for a date range |
| `add_meal_to_plan` | Add a recipe to the meal plan |
