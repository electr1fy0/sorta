# Sorta
A simple file organizer that sorts files in a directory based on file extensions, custom keywords, or finds duplicates.

## What it does
Sorta cleans up messy directories by automatically moving files into organized folders. It has four commands:

1. **Extension-based sorting** - Groups files by type (PDFs → docs, images → images, videos → movies)
2. **Keyword-based sorting** - Groups files based on filename keywords you define
3. **Duplicate detection** - Finds and moves duplicate files using SHA256 checksums
4. **Largest files** - Lists the top 5 largest files in a directory

## Installation
```bash
git clone https://github.com/electr1fy0/sorta.git
cd sorta
go build -o sorta
```

## Usage

### Extension-based sorting
```bash
./sorta ext <directory>
./sorta ext ~/Downloads
./sorta ext Desktop/messy-folder --dry
```

Automatically creates these folders and moves files:
- `docs/` - .pdf, .docx, .pages, .md, .txts files
- `images/` - .png, .jpg, .jpeg, .heic, .heif files
- `movies/` - .mp4, .mov files
- `slides/` - .pptx files

### Keyword-based sorting
```bash
./sorta conf <directory>
./sorta conf ~/Documents
./sorta conf . --dry
```

Uses a config file to define custom sorting rules. The program creates `~/.sorta-config` automatically if it doesn't exist.

**Config format:**
```
keyword1,keyword2,keyword3=FolderName
another,set,of,keywords=AnotherFolder
*=Misc
```

**Example config:**
```
invoice,receipt,bill=Financial
photo,img,picture=Photos
code,src,dev=Development
*=Others
```

**Wildcard support:**
- Use `*` as a keyword to match all files that don't match any other rules
- Files matching specific keywords will always take priority over the wildcard
- Only one wildcard rule is supported

### Duplicate detection
```bash
./sorta dupl <directory>
./sorta dupl ~/Downloads --dry
```

- Calculates SHA256 checksums for all files
- Moves duplicate files to a `duplicates/` folder
- First occurrence of each file stays in place
- Files already in the duplicates folder are skipped

### Find largest files
```bash
./sorta lrg <directory>
./sorta lrg ~/Documents
```

Lists the top 5 largest files in the directory with human-readable sizes.

## Flags
- `--dry` - Dry run mode: shows what would be moved without actually doing it

Can be used with any command:
```bash
./sorta ext ~/Downloads --dry
./sorta conf . --dry
./sorta dupl Desktop --dry
```

## Directory paths
- Relative paths are resolved from your home directory: `Downloads`, `Desktop/project`
- Absolute paths work too: `/Users/you/Documents`, `/tmp/files`

## Summary output
After sorting, Sorta shows:
- Number of files moved
- Number of files skipped (files that didn't match any rules)

## Config file details
- Located at `~/.sorta-config`
- Lines starting with `//` are comments
- Format: `comma,separated,keywords=DestinationFolder`
- **No spaces** between keywords, commas, and the equals sign
- Keywords match anywhere in the filename (case-sensitive)
- First matching rule wins
- Use `*` to catch all remaining files
- Files that don't match any keywords are skipped (unless `*` is used)

## Notes
- Files are moved, not copied
- Hidden files (starting with `.`) are ignored
- Empty directories are automatically cleaned up after sorting
- Creates destination folders automatically if they don't exist
- Requires write permissions in the target directory
- File sizes are displayed in human-readable format (B, KB, MB, GB, etc.)
