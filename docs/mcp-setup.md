# MCP Setup — paprika-3-mcp (fork)

## Build

```bash
cd /home/john-henry-erikson/code/paprika-3-mcp
go install ./cmd/paprika-3-mcp/
```

## Configure Claude Code

The repo includes a `.mcp.json` that Claude Code loads automatically. No manual settings needed — just make sure the binary is installed to `$GOPATH/bin`.

If you need to add it manually to `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "paprika": {
      "command": "/home/john-henry-erikson/gopath/bin/paprika-3-mcp",
      "args": [
        "--refresh-interval", "5m",
        "--aisle-map", "/home/john-henry-erikson/code/paprika-3-mcp/aisles/woodmans_east.json",
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
