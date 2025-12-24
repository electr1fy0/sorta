# Sorta

Sorts files in a directory based on keywords, finds duplicates, or lists largest files.

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

### Sort by keywords (default)

```bash
sorta sort <directory>
# Aliases: s, organize
sorta s ~/Downloads
sorta organize Desktop/messy-folder --dry-run
```

Uses `~/.sorta/config` to define sorting rules. Creates a default config if it doesn't exist.
Users can manually edit the file at `~/.sorta/config`.

**Config format:**

```
# Lines starting with '#' or '//' are comments
FolderName=keyword1,keyword2
AnotherFolder=another,set,of,keywords

# Keywords can be regular expressions
OneMoreFolder=regex(your-regular-expression)

# FolderName can also be a relative folderpath.
foo/bar/oof = rab creates a folder tree.

# Match all files
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

### Smart Rename

```bash
sorta rename <directory>
# Aliases: rn, mv
sorta rn ~/Downloads --dry-run
```

Uses Gemini to sanitize filenames into a concise, readable format (Title_Snake_Case).

**Features:**

- Standardizes names like "Operating Systems Sem 5.pdf" to "OS_S5_notes.pdf"
- Removes clutter such as copy, final, v2
- Keeps important acronyms like DSA, TCP, OS
- Strips off metadata
- Removes redundancy and shortens verbose language

Requires `GEMINI_API_KEY` environment variable set.

Advanced users can manually modify the prompt at `~/.sorta/prompt`.

Note: All filenames are sent to Gemini for sanitization for the rename command. Nothing else is shared.

### Find duplicates

```bash
sorta duplicates <directory>
# Aliases: dupl, dedupe, dd
sorta dd ~/Downloads --dry-run
```

Uses SHA256 checksums. Moves dupes to `duplicates/` folder, keeps the first occurrence. Use `--nuke` to delete the duplicates folder.

### List largest files

```bash
sorta large <directory>
# Aliases: lrg, top, big
sorta top ~/Downloads
```

Shows top 5 largest files.

### Initialize directory

```bash
sorta init <directory>
# Aliases: setup, create
```

Creates a local `.sorta/` folder inside the target directory with copies of your default config and prompt. This allows per-directory configuration.
When running `sort` in this directory, `sorta` will automatically detect and use this local configuration.

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
# Aliases: rm, del
```

Edits `~/.sorta/config` by default. If a local config exists in the current directory, or if `--config-path` is provided, it edits that instead.

### History & Undo

```bash
sorta history
# Aliases: log, ls
# View past operations

sorta undo <directory>
# Aliases: u, revert
# Revert the last operation in the specified directory
```

## Flags

- `--dry-run` - Preview changes without moving files
- `--config-path` - Path to config file (default `~/.sorta/config`)
- `--recurse-level` - Level of recursion to perform in the directory (Unix only)

## Examples

**1. One-off sorting with custom config:**

```bash
sorta sort . --config-path ./my-special-config
```

**2. Quickly organizing a cluttered Downloads folder:**

```bash
sorta s ~/Downloads
```

**3. Renaming messy course materials:**

```bash
export GEMINI_API_KEY=your_key
sorta rn ~/Uni/Semester1 --dry-run
# Check output, then run without --dry-run
sorta rn ~/Uni/Semester1
```

**4. Cleaning up duplicate photos:**

```bash
sorta dd ~/Photos --nuke
```
