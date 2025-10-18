# Sorta
Sorts files in a directory based on keywords, finds dupes, or lists largest files.

## Install
### From source
```bash
git clone https://github.com/electr1fy0/sorta.git
cd sorta
go build -o sorta
```
### Using `go install`
```bash
go install github.com/electr1fy0/sorta@latest
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

Use `*` to match everything that doesn't match other rules. Specific keywords always take priority.

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

## Flags
- `--dry` - Preview changes without moving files
- `--interactive` - Confirm before each move

## Notes
- Moves files, doesn't copy
- Ignores hidden files (starting with `.`)
- Auto-cleans empty directories
- Creates destination folders as needed
- Prompts to undo after sorting
