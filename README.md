# Sorta

A simple file organizer that sorts files in a directory based on either file extensions or custom keywords.

## What it does

Sorta cleans up messy directories by automatically moving files into organized folders. It has two modes:

1. **Extension-based sorting** - Groups files by type (PDFs → docs, images → images)
2. **Keyword-based sorting** - Groups files based on filename keywords you define

## Installation

```bash
git clone https://github.com/electr1fy0/sorta.git
cd sorta
go build -o sorta
```

## Usage

Run the program:
```bash
./sorta
```

You'll be prompted for:
1. **Directory path** (relative to your home directory, e.g., `Downloads` or `Desktop/project`)
2. **Sorting mode**:
   - `0` = Extension-based sorting
   - `1` = Keyword-based sorting

### Extension-based sorting (Mode 0)

Automatically creates these folders and moves files:
- `docs/` - PDF files
- `images/` - PNG, JPG, JPEG files

### Keyword-based sorting (Mode 1)

Uses a config file to define custom sorting rules. Create `~/.sorta-config` with this format:

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

## Examples

```bash
# Sort Downloads folder by file type
./sorta
# Enter: Downloads
# Enter: 0

# Sort project folder using custom keywords
./sorta
# Enter: Desktop/messy-project
# Enter: 1
```

## Config file format

Each line in `~/.sorta-config`:
```
comma,separated,keywords=DestinationFolder
```

- Keywords are case-sensitive
- Matches partial filenames (e.g., "img" matches "img001.jpg")
- First matching rule wins
- Files that don't match any keywords stay in the original location

## Notes

- Files are moved, not copied
- Hidden files (starting with `.`) are ignored
- The program lists all files it processes with numbers for reference
- Make sure you have write permissions in the target directory
