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
Obtain the `.exe`, then add it to PATH:

1.  Move `sorta.exe` to `C:\Program Files\sorta\`
2.  Search "Environment Variables" → Edit "Path" → Add `C:\Program Files\sorta\`

### Using `go install`

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
```

Uses `~/.sorta-config` to define sorting rules. Creates a default config if it doesn't exist.

**Config format:**

```
# Lines starting with '#' or '//' are comments
FolderName=keyword1,keyword2
AnotherFolder=another,set,of,keywords
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

Uses Gemini to sanitize filenames into a concise, readable format (`snake_case`).

**Features:**

  - **Standardizes:** `Operating Systems Sem 5.pdf` → `os_s5_notes.pdf`
  - **Cleans:** Removes clutter ("copy", "final", "v2").
  - **Smart Formatting:** Preserves acronyms (DSA, TCP) while fixing inconsistencies.
  - **Safety:** Ensures uniqueness and never changes file extensions.

*Requires `GEMINI_API_KEY` environment variable set.*

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

### Manage config

```bash
sorta config add <foldername> <keyword1> <keyword2>...
sorta config remove <foldername>
```

Or manually edit `~/.sorta-config` to adjust folder names, keywords, and match priority.

## Flags

  - `--dry` - Preview changes without moving files
  - `--interactive` - Confirm before each move

## Notes

  - Moves files, doesn't copy
  - Ignores hidden files (starting with `.`)
  - Auto-cleans empty directories
  - Creates destination folders as needed
  - Prompts to undo after sorting
