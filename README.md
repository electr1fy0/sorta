# Sorta

Sorts files in a directory based on keywords, finds duplicates, renames files with AI, or lists largest files.

## Install

### From GitHub Releases

Download the latest binary from [GitHub Releases](https://github.com/electr1fy0/sorta/releases) for your platform.

**Linux/macOS:**

```bash
# Download and extract
curl -LO https://github.com/electr1fy0/sorta/releases/latest/download/sorta-linux-amd64.tar.gz
tar -xzf sorta-linux-amd64.tar.gz
# Move to PATH
sudo mv sorta /usr/local/bin/
```

**Windows:**
Obtain the .exe, then add it to PATH:

1. Move sorta.exe to `C:\Program Files\sorta\`
2. Search "Environment Variables" → Edit "Path" → Add `C:\Program Files\sorta\`

### Using go install

```bash
go install github.com/electr1fy0/sorta@latest
```

If `sorta` command isn't found, add Go's bin directory to PATH:

**Linux/macOS:**

```bash
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc  # or ~/.zshrc
source ~/.bashrc  # or ~/.zshrc
```

**Windows (PowerShell):**

```powershell
$gopath = go env GOPATH
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";$gopath\bin", "User")
```

### From source

```bash
git clone https://github.com/electr1fy0/sorta.git
cd sorta
go build -o sorta
# Move to PATH
sudo mv sorta /usr/local/bin/
```

## Usage

For all commands that accept a `<directory>` argument, relative paths are resolved against the **Current Working Directory (CWD)**. Paths starting with `~` (e.g., `~/Downloads`) are expanded to the user's home directory.

### Interactive Review

When running in a terminal, `sorta` presents an interactive list of planned operations before executing them.

- **Navigation:** Use `↑` / `↓` (or `k` / `j`) to scroll.
- **Selection:** Press `Space` to toggle an individual file operation.
- **Batch:** Press `a` to toggle all operations.
- **Confirm:** Press `Enter` to proceed with the selected operations.
- **Cancel:** Press `q` or `Esc` to abort.

### Sort by keywords

```bash
sorta sort [directory]
# Aliases: s, organize
sorta s ~/Downloads
sorta organize Desktop/messy-folder --dry-run
```

Sorts files based on rules defined in `~/.sorta/config`.
If the directory argument is omitted, `sorta` will prompt for it.
You will be able to review and select which files to move before any changes are made.

**Flags:**

- `--inline "Folder=ext1,ext2"`: Skip config file and use a single one-off rule.
  Example: `sorta sort . --inline "Images=jpg,png"`

**Config format:**

```
# Lines starting with '#' or '//' are comments (only full-line comments are supported)
FolderName=keyword1,keyword2
AnotherFolder=another,set,of,keywords

# Keywords can be regular expressions
OneMoreFolder=regex(your-regular-expression)

# FolderName can also be a relative folder path (creates a nested folder tree)
foo/bar/baz=keyword

# Use '.' as the folder name to keep files in the root directory
.=doc,docx

# Match all remaining files
Misc=*
```

**Example:**

```
Financial=invoice,receipt,bill
Photos=photo,img,picture
Development=code,src,dev
Others=*
```

Use `*` to match everything that doesn't match other rules. Specific keywords always take priority. Rules higher in the config are matched first.

> **Note:** Only full-line comments (lines starting with `#` or `//`) are supported. Inline comments on rule lines are not stripped and will be treated as part of the keyword.

Hidden files and directories (names starting with `.`) are always skipped during scanning.

Ignore patterns:

- Add lines prefixed with `!` in config to ignore paths/files while scanning.
- Patterns support simple glob matching (`*.tmp`, `build`, `archive/*.zip`).
- You can also use ignore files: `<target>/.sortaignore`, `<target>/.sorta/ignore`, and `~/.sorta/ignore`.
- Ignore rules apply to `sort`, `rename`, `duplicates`, and `bench`.

### Smart Rename (beta)

```bash
sorta rename <directory>
# Aliases: rn
sorta rn ~/Downloads
```

Uses Gemini to sanitize filenames into a concise, readable format (Title_Snake_Case).
You can interactively review and deselect specific renames before they are applied.

**Features:**

- Standardizes names like "Operating Systems Sem 5.pdf" to "OS_S5_notes.pdf"
- Removes clutter such as copy, final, v2
- Keeps important acronyms like DSA, TCP, OS
- Strips metadata and software-generated prefixes
- Removes redundancy and shortens verbose language

Requires `GEMINI_API_KEY` environment variable set.

Note: All filenames are sent to Gemini for sanitization for the rename command. Nothing else is shared.

### Find duplicates

```bash
sorta duplicates [directory]
# Aliases: dupl, dedupe, dd
sorta dd ~/Downloads
```

Uses SHA256 checksums. Moves dupes to `duplicates/` folder, keeps the first occurrence. Use `--nuke` to permanently delete the duplicate files instead of moving them. **Operations using `--nuke` cannot be undone.**
Duplicate targets are deterministic and collision-safe: `<name>_<hash8>_<path6>.<ext>`.
Includes an interactive review step to verify files before moving or deleting. If directory is omitted, it will be prompted for.
Duplicate detection uses a bounded parallel hashing pipeline and stores a metadata hash cache in `~/.sorta/hash-cache.json` for faster repeated scans.

### List largest files

```bash
sorta large <directory>
# Aliases: lrg, top, big
sorta top ~/Downloads
```

Shows the top 5 largest files in the directory.

### Microbenchmark duplicate scan

```bash
sorta bench <directory>
```

Runs a non-destructive duplicate-scan benchmark and prints walk/plan timing, hash counts, cache hit/miss counts, and hash throughput (MiB/s).

### Check ignore rules

```bash
sorta check-ignore <path>
sorta check-ignore <directory> <path>
```

Explains whether a given path would be ignored and which rule is responsible. Useful for debugging ignore patterns.

### Manage config

```bash
sorta config list
# Aliases: ls, show
# Lists all rules in a neat table

sorta config path
# Aliases: p, location
# Shows the path of the active configuration file

sorta config add "<foldername> = <keyword1>, <keyword2>..."
# Aliases: new, a
# Example: sorta config add "Images = jpg, png, gif"

sorta config remove <foldername>
# Aliases: rm, del, delete

sorta config edit
# Aliases: e, open
# Opens the config file in your default editor ($EDITOR or $VISUAL)

sorta config init <directory>
# Aliases: setup, create, initialize
# Creates a local .sorta/ folder with copies of your default config
```

Edits `~/.sorta/config` by default. If a local config exists in the current directory, or if `--config-path` is provided, it edits that instead.

`sorta config init` creates a local `.sorta/` folder inside the target directory. When running `sort` in that directory, `sorta` will automatically detect and use the local configuration instead of the global one.

### History & Undo

```bash
sorta history
# Aliases: log, ls
# View past operations

sorta history --oneline
# Compact output (id type count root)

sorta undo [directory]
# Aliases: u, revert
# Revert the last operation in the specified directory
```

### Version

```bash
sorta version
# Print the version number of sorta
```

## Flags

### Global
- `--dry-run` - Preview changes and exit (skips confirmation prompt)
- `--config-path` - Path to config file (default `~/.sorta/config`). Relative paths are resolved against the CWD; paths starting with `~` are expanded to the home directory.
- `--recurse-level` - Maximum folder depth to scan (default: 1024)

### Command Specific
- `--inline` (sort): Define a one-off rule, ignoring config file. Format: `"Folder=kw1,kw2"`.
- `--nuke` (duplicates): Permanently delete duplicate files instead of moving them.

## Examples

**1. One-off sorting with a custom config:**

```bash
sorta sort . --config-path ./my-special-config
```

**2. Sorting with a quick inline rule:**

```bash
sorta sort ~/Downloads --inline "Images=jpg,png,gif"
```

**3. Renaming messy course materials with AI:**

```bash
export GEMINI_API_KEY=your_key
sorta rn ~/Uni/Semester1
```

**4. Cleaning up duplicate photos:**

```bash
sorta dd ~/Photos --nuke
```

**5. Checking why a file is being ignored:**

```bash
sorta check-ignore ~/Downloads node_modules
```

**6. Viewing and undoing the last operation:**

```bash
sorta history
sorta undo ~/Downloads
```