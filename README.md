# td - Terminal Todo List Manager

A fast, clean TUI (Terminal User Interface) todo list manager built with Go. Designed for local usage with powerful features like tagging, filtering, and multi-line descriptions.

## Features

- **Clean TUI Interface** - Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Multi-line Todos** - First line is the title, additional lines are description
- **Hashtag Support** - Use `#tags` anywhere in your todos for organization
- **Tag Filtering** - Filter by multiple tags with AND logic
- **Timestamps** - Automatic tracking of creation and completion times (displayed in local timezone)
- **Smart Sorting** - Active items by manual order, completed items by completion date
- **Persistent Storage** - JSON file storage with atomic writes
- **Configurable Location** - Store your todos wherever you want

## Installation

### Build from Source

```bash
git clone https://github.com/jashort/td.git
cd td
go build -o td
sudo mv td /usr/local/bin/  # Optional: install globally
```

### Run Directly

```bash
go run .
```

## Usage

### Starting the App

```bash
# Use default location (~/.td/todos.json)
td

# Use custom file location
td --file ~/projects/work-todos.json
td -f ./todos.json

# Use environment variable
export TD_FILE=~/work-todos.json
td
```

### Keybindings

#### Navigation
- `j` or `↓` - Move cursor down
- `k` or `↑` - Move cursor up
- `g` - Jump to top
- `G` - Jump to bottom

#### Actions
- `a` - Add new todo
- `e` - Edit selected todo
- `d` - Delete selected todo
- `Space` - Toggle completion status
- `t` - Toggle show/hide completed items
- `f` - Filter by tags
- `J` or `Ctrl+J` - Move todo down in order (active items only)
- `K` or `Ctrl+K` - Move todo up in order (active items only)

#### General
- `?` or `h` - Show help
- `q` or `Ctrl+C` - Quit (auto-saves)

### Add/Edit Mode
- `Enter` - New line (multi-line support)
- `Ctrl+S` - Save
- `Esc` - Cancel

### Filter Mode
- Type tag names separated by spaces (e.g., `work urgent`)
- Matches todos that have ALL specified tags (AND logic)
- `Enter` - Apply filter
- `Esc` - Clear filter

## Features in Detail

### Hashtags / Tags

Simply add hashtags anywhere in your todo text:

```
Buy groceries #personal #shopping
Fix bug in parser #work #urgent
Write documentation
This is a multi-line todo
with multiple lines #work
```

Tags are automatically extracted and can be used for filtering.

### Multi-line Todos

Press `Enter` while adding or editing a todo to create multiple lines:
- **First line** = Title (displayed prominently)
- **Subsequent lines** = Description (indented, slightly dimmed)

### Timestamps

Every todo automatically tracks:
- **Created**: When the todo was added
- **Completed**: When it was marked as done (if applicable)

All timestamps are stored in UTC but displayed in your local timezone.

### Completed Items

- Hidden by default to keep your view clean
- Press `t` to toggle visibility
- Always shown at the bottom when visible
- Sorted by completion date (most recent first)

### Reordering

- Use `J` and `K` (or `Ctrl+J`/`Ctrl+K`) to reorder active todos
- Completed items automatically sort to the bottom
- Order is preserved across sessions

### Storage

Todos are stored in a single JSON file:
- **Default location**: `~/.td/todos.json`
- **Custom location**: Use `--file` flag or `TD_FILE` environment variable
- **Auto-save**: Saves after every change
- **Atomic writes**: Uses temporary file + rename for safety
- **Human-readable**: Pretty-printed JSON for manual editing if needed

## Configuration Priority

The app determines where to store todos in this order:

1. `--file` / `-f` command-line flag (highest priority)
2. `TD_FILE` environment variable
3. `~/.td/todos.json` (default)

## Data Format

The JSON format is simple and human-readable:

```json
{
  "todos": [
    {
      "id": "uuid",
      "title": "Buy groceries",
      "description": "",
      "completed": false,
      "created_at": "2026-02-13T18:00:00Z",
      "completed_at": null,
      "tags": ["personal", "shopping"],
      "order": 0
    }
  ],
  "next_order": 1
}
```

## Development

### Running Tests

```bash
go test -v
```

### Building

```bash
go build -o td
```

### Dependencies

- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [bubbles/textarea](https://github.com/charmbracelet/bubbles) - Multi-line input
- [uuid](https://github.com/google/uuid) - UUID generation

## License

MIT License - See LICENSE file for details

## Contributing

Contributions welcome! Please open an issue or PR.

## Author

Built with Go and Bubble Tea
