# paprika-3-mcp

A [Model Context Protocol (MCP)](https://modelcontextprotocol.io/introduction) server that exposes your **Paprika 3** recipes as LLM-readable resources — and lets an LLM like Claude create, update, and plan meals with your Paprika app.

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
- `list_recipes` — List all recipes with name, UID, categories, prep/cook time, and star rating
- `get_recipe` — Fetch full recipe content (ingredients, directions, notes) by UID
- `create_paprika_recipe` — Save a new recipe to your Paprika app
- `update_paprika_recipe` — Modify an existing recipe

**Grocery List**
- `get_grocery_list` — Fetch all items in a Paprika grocery list with their current aisle assignments
- `update_grocery_item_aisle` — Set the aisle label on one or more grocery items
- `setup_woodmans_aisles` — Map all grocery list items to their Woodman's East aisle (supports dry run)

**Meal Plan**
- `get_meal_plan` — Fetch the meal plan for a date range
- `add_meal_to_plan` — Add a recipe to the meal plan (idempotent)

**Pantry**
- `get_pantry` — Fetch all items in the Paprika pantry with name, quantity, and in-stock status

## ⚙️ Prerequisites

- ✅ A Mac, Linux, or Windows system
- ✅ [Paprika 3](https://www.paprikaapp.com/) installed with cloud sync enabled
- ✅ Your Paprika 3 **username and password**
- ✅ Claude or any LLM client with **MCP tool support** enabled

## 🛠 Installation

You can download a prebuilt binary from the [Releases](https://github.com/soggycactus/paprika-3-mcp/releases) page.

### Build from source

```bash
git clone https://github.com/soggycactus/paprika-3-mcp
cd paprika-3-mcp
go install ./cmd/paprika-3-mcp/
```

### 🍎 macOS (via Homebrew)

If you're on macOS, the easiest way to install is with [Homebrew](https://brew.sh/):

```bash
brew tap soggycactus/tap
brew install paprika-3-mcp
```

### 🐧 Linux / 🪟 Windows

1. Go to the [latest release](https://github.com/soggycactus/paprika-3-mcp/releases).
2. Download the appropriate archive for your operating system and architecture:
   - `paprika-3-mcp_<version>_linux_amd64.zip` for Linux
   - `paprika-3-mcp_<version>_windows_amd64.zip` for Windows
3. Extract the zip archive:
   - **Linux**:
     ```bash
     unzip paprika-3-mcp_<version>_<os>_<arch>.zip
     ```
   - **Windows**:
     - Right-click the `.zip` file and select **Extract All**, or use a tool like 7-Zip.
4. Move the binary to a directory in your system's `$PATH`:

   - Linux:

     ```bash
     sudo mv paprika-3-mcp /usr/local/bin/
     ```

   - Windows:
     - Move `paprika-3-mcp.exe` to any folder in your `PATH` (e.g., `%USERPROFILE%\bin`)

### ✅ Test the installation

```bash
paprika-3-mcp --version
```

## 🤖 Setting up Claude

If you haven't setup MCP before, [first read more about how to install Claude Desktop client & configure an MCP server.](https://modelcontextprotocol.io/quickstart/user)

### Claude Code (recommended)

The repo includes a `.mcp.json` that Claude Code picks up automatically. Set your credentials as environment variables and it just works:

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

Restart Claude and you should see the MCP server tools after clicking on the hammerhead icon:

![MCP server running with Claude](docs/install.png)

## ⚙️ Configuration flags

| Flag | Default | Description |
|------|---------|-------------|
| `--refresh-interval` | `5m` | How often to refresh the recipe resource cache |
| `--aisle-map` | `aisles/woodmans_east.json` | Path to aisle map JSON file (used by `setup_woodmans_aisles`) |
| `--grocery-list` | *(first list)* | Default grocery list name; empty string uses the first list |
| `--version` | — | Print version and exit |

Credentials are read from the `PAPRIKA_USERNAME` and `PAPRIKA_PASSWORD` environment variables.

## 📄 License

This project is open source under the [MIT License](./LICENSE) © 2025 [Lucas Stephens](https://github.com/soggycactus).

---

#### 🗂 Miscellaneous

##### 📄 Where can I see the server logs?

The MCP server writes structured logs using Go's `slog` with rotation via `lumberjack`. Log files are automatically created based on your operating system:

| Operating System | Log File Path                             |
| ---------------- | ----------------------------------------- |
| macOS            | `~/Library/Logs/paprika-3-mcp/server.log` |
| Linux            | `/var/log/paprika-3-mcp/server.log`       |
| Windows          | `%APPDATA%\paprika-3-mcp\server.log`      |
| Other / Unknown  | `/tmp/paprika-3-mcp/server.log`           |

> 💡 Logs are rotated automatically at 100MB, with only 5 backup files kept. Logs are also wiped after 10 days.
