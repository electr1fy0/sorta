# Sorta
A simple file organizer that sorts files in a directory based on file extensions, custom keywords, or finds duplicates.

## What it does
Sorta cleans up messy directories by automatically moving files into organized folders. It has three modes:
1. **Extension-based sorting** - Groups files by type (PDFs → docs, images → images, videos → movies)
2. **Keyword-based sorting** - Groups files based on filename keywords you define
3. **Duplicate detection** - Finds and moves duplicate files using SHA256 checksums

## Installation
```bash
git clone https://github.com/electr1fy0/sorta.git
cd sorta
go build -o sorta
```

## Usage
### Basic usage
```bash
./sorta                      # Interactive mode (prompts for directory)
./sorta [directory]
```

### Dry run flag
```bash
./sorta --dry [directory]    # See what would happen without actually moving files
./sorta Desktop/messy-folder --dry
```

You'll be prompted for:
1. **Directory path** (relative to your home directory, e.g., `Downloads` or `Desktop/project`)
2. **Sorting mode**:
   - `0` = Extension-based sorting
   - `1` = Keyword-based sorting
   - `2` = Find and move duplicates

## Sorting modes

### Extension-based sorting (Mode 0)
Automatically creates these folders and moves files:
- `docs/` - .pdf, .docx, .pages, .md, .txt files
- `images/` - .png, .jpg, .jpeg, .heic, .heif, .webp files
- `movies/` - .mp4, .mov files

### Keyword-based sorting (Mode 1)
Uses a config file to define custom sorting rules. The program creates `~/.sorta-config` automatically if it doesn't exist.

**Config format:**
```
keyword1,keyword2,keyword3=FolderName
another,set,of,keywords=AnotherFolder
```

**Example config:**
```
invoice,receipt,bill=Financial
photo,img,picture=Photos
code,src,dev=Development
```

### Duplicate detection (Mode 2)
- Calculates SHA256 checksums for all files
- Shows checksum, filename, and file size for each file
- Moves duplicate files to a `duplicates/` folder
- First occurrence of each file stays in place

## Flags
- `--dry` - Dry run mode: shows what would be moved without actually doing it

## Examples
```bash
# Sort Downloads folder by file type with dry run
./sorta --dry Downloads
# Mode: 0

# Sort project folder using custom keywords
./sorta Desktop/messy-project
# Mode: 1

# Find duplicates in current directory
./sorta .
# Mode: 2
```

## Summary output
After sorting, Sorta shows:
- Number of files moved
- Number of files skipped (files that didn't match any rules)
- File sizes in bytes
- For duplicates mode: checksum information for each file

## Config file details
- Located at `~/.sorta-config`
- Lines starting with `//` are comments
- Format: `comma,separated,keywords=DestinationFolder`
- Keywords match anywhere in the filename
- Case-sensitive matching
- First matching rule wins
- Files that don't match any keywords are skipped

## Notes
- Files are moved, not copied
- Hidden files (starting with `.`) and directories are ignored
- The program shows file sizes and detailed move information
- Requires write permissions in the target directory
- Creates destination folders automatically if they don't exist
