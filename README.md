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
sorta <directory>
sorta ~/Downloads
sorta Desktop/messy-folder --dry
sorta Documents --recurselevel 1 # Unix only flag
```

Uses `~/.sorta/config` to define sorting rules. Creates a default config if it doesn't exist.
Users can manually edit the file at `~/.sorta/config`.

**Config format:**

```
# Lines starting with '#' or '//' are comments
FolderName=keyword1,keyword2
AnotherFolder=another,set,of,keywords
OneMoreFolder=regex(your-regular-expression)
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
sorta rename ~/Downloads --dry
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
sorta dupl <directory>
sorta dupl ~/Downloads --dry
```

Uses SHA256 checksums. Moves dupes to `duplicates/` folder, keeps the first occurrence. Use `--nuke` to delete the duplicates folder.

### List largest files

```bash
sorta lrg <directory>
```

Shows top 5 largest files.

### Initialize directory

```bash
sorta init <directory>
```

Creates a local `.sorta/` folder inside the target directory with copies of your default config and prompt. This allows per-directory configuration.

### Manage config

```bash
sorta config add <foldername> "<keyword1>, <keyword2>..."
sorta config remove <foldername>
```

Edits `~/.sorta/config`.
Users may also manually edit the config at the same path.

## Flags

- `--dry` - Preview changes without moving files
- `--interactive` - Confirm before each move
- `--config` - Path to config file (default `~/.sorta/config`)

## Notes

- Moves files, doesn't copy
- Ignores hidden files (starting with `.`)
- Auto-cleans empty directories
- Creates destination folders as needed
- Prompts to undo after sorting
- Foldernames higher up the config are prioritized in conflicts.
