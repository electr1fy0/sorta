# sorta

sorta is a rule based CLI file organization tool that combines rule-based sorting with agentic capabilities. It categorizes files, renames them, and removes duplicates.

## Features

- **Rule-based Sorting:** Categorize files based on keywords, extensions, or regex patterns.
- **Agentic Renaming:** Use LLMs (via Groq) to rename files based on their current titles.
- **Natural Language Goals:** Describe a directory organization goal in natural language; sorta plans and executes it.
- **Deduplication:** Remove duplicate files using content hashing.
- **Safety:**
  - **Dry Run:** Preview changes before application.
  - **Interactive TUI:** Selectively approve operations.
  - **Undo:** Revert the last operation.
  - **History:** Log of all operations.
- **Configuration:** Support for global (~/.sorta/config), local (./.sorta/config), and inline rules.
- **Utilities:** Largest file listing, ignore-rule debugging, and config management.

## Installation

```bash
go install github.com/electr1fy0/sorta@latest
```

## Setup

1. Initialize sorta in a directory to use local configuration:
   ```bash
   sorta config init .
   ```

2. For agentic features (rename and goal), you need a Groq API key:
   ```bash
   export GROQ_API_KEY=your_api_key_here
   ```

## Usage

### Global Flags
- `--dry-run`: Preview changes without applying them.
- `--config-path <path>`: Specify a custom config file path.
- `--recurse-level <int>`: Set directory recursion depth (default: 1024).

### Rule-based Sorting
Organize a directory based on configuration rules:
```bash
sorta sort ./Downloads
```
Use inline rules:
```bash
sorta sort ./Downloads --inline "Images = jpg, png; Docs = pdf, docx"
```

### Agentic Renaming
Let the agent rename files to be more descriptive:
```bash
sorta rename ./Screenshots --model "openai/gpt-oss-20b"
```

### Natural Language Goals
Execute organization tasks using natural language:
```bash
sorta goal ./Projects "Group by project and file type and deduplicate the folder"
```

### Deduplication
Find and handle duplicate files:
```bash
sorta duplicates ./Photos
sorta duplicates ./Photos --nuke
```

### Safety and History
```bash
sorta undo ./Downloads
sorta history --oneline
```

### Configuration Management
- `sorta config list`: Show active rules and ignore patterns.
- `sorta config edit`: Open config in default editor.
- `sorta config add "Folder = key1, key2"`: Append a rule to the config.
- `sorta config remove Folder`: Remove a rule by folder name.
- `sorta config path`: Show the path to the active config file.

### Utility Commands
- `sorta large <dir>`: Show the top 5 largest files.
- `sorta check-ignore <path>`: Debug ignore rules for a path.
- `sorta version`: Show version and build info.

## Configuration

Sorting rules are defined in a simple format: FolderName = keyword1, keyword2, extension.

Example ~/.sorta/config:
```text
Finance = invoice, bill, .pdf
Music = track, song, .mp3
Images = .jpg, .png, .gif
Archive = .zip, .tar.gz
Others = *
!node_modules
!*.tmp
```
- `*` matches any file not caught by other rules.
- `!` prefixes ignore patterns (supports .gitignore style globbing).

## Architecture

sorta uses a modular architecture to separate file discovery, operation planning, and execution.

### Core Components

- **Discovery & Ignore Engine:** Recursively scans directories while respecting `.gitignore`, `.sortaignore`, and config-defined patterns.
- **Sorter Interface:** Pluggable logic for determining file actions. Implementations include `ConfigSorter` (rule-based), `Renamer` (LLM-based), and `DuplicateFinder` (hash-based).
- **Planner:** Aggregates potential operations and handles conflict resolution (e.g., naming collisions).
- **Worker Pool:** Orchestrates concurrent I/O operations. Uses a dynamically sized pool of workers (up to 16) to parallelize hashing and AI batch processing.
- **Executor & Transaction Log:** Applies file system changes. Every operation is recorded in a local JSON-line history file to enable undos.

Each operation is stored as a JSON object:
```json
{
  "id": "uuid",
  "type": "move",
  "source": "/path/to/old",
  "dest": "/path/to/new",
  "timestamp": "..."
}
```
The `undo` command reads the last transaction and executes the inverse operations (e.g., moving files back to their source paths).

### Agentic Goal Execution

The `goal` command uses a feedback loop:
1. **Instruction:** User provides a goal.
2. **Analysis:** Agent examines the directory structure.
3. **Plan:** Agent generates a sequence of operations.
4. **Validation:** User reviews the plan.
5. **Execution:** Executor applies the plan.

## License

MIT
