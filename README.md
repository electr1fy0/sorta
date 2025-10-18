# Sorta
A simple file organizer that sorts files in a directory based on custom keywords, finds duplicates, or lists the largest files.

## What it does
Sorta cleans up messy directories by automatically moving files into organized folders. It has three main commands:

1.  **Keyword-based sorting** - The default command. Groups files based on filename keywords you define in a configuration file.
2.  **Duplicate detection** - Finds and moves duplicate files using SHA256 checksums.
3.  **Largest files** - Lists the top 5 largest files in a directory.

## Installation
```bash
git clone https://github.com/electr1fy0/sorta.git
cd sorta
go build -o sorta
```

## Usage

### Keyword-based sorting (Default)
To sort files based on your configuration, simply provide a directory path.
```bash
./sorta <directory>
./sorta ~/Downloads
./sorta Desktop/messy-folder --dry
```

This command uses a config file at `~/.sorta-config` to define custom sorting rules. If the file doesn't exist, `sorta` will create a default one for you to edit.

**Config format:**
```
# Lines starting with '#' or '//' are comments.
FolderName=keyword1,keyword2
AnotherFolder=another,set,of,keywords
Misc=*
```

**Example config:**
```
Financial=invoice,receipt,bill
Photos=photo,img,picture
Development=code,src,dev
Others=*
```

**Wildcard support:**
- Use `*` as a keyword to match all files that don't match any other rules.
- Files matching specific keywords will always take priority over the wildcard.

### Duplicate detection
```bash
./sorta dupl <directory>
./sorta dupl ~/Downloads --dry
```

- Calculates SHA256 checksums for all files.
- Moves duplicate files to a `duplicates/` folder.
- The first occurrence of each file is left in its original place.

### Find largest files
```bash
./sorta lrg <directory>
./sorta lrg ~/Documents
```

Lists the top 5 largest files in the specified directory with human-readable sizes.

### Manage Configuration
You can easily add or remove rules from your `~/.sorta-config` file.

**Add a rule:**
```bash
./sorta config add <keyword> <foldername>
./sorta config add invoice Financial
```

**Remove a rule:**
```bash
./sorta config remove <keyword>
./sorta config remove invoice
```

## Global Flags
These flags can be used with any command.

-   `--dry` - Dry run mode: shows what would be moved without making any changes.
-   `--interactive` - Interactive mode: prompts for confirmation before moving each file.

**Examples:**
```bash
./sorta ~/Downloads --dry
./sorta dupl Desktop --interactive
```

## Notes
-   Files are moved, not copied.
-   Hidden files (starting with `.`) are ignored.
-   Empty directories are automatically cleaned up after sorting.
-   Destination folders are created automatically if they don't exist.
-   After a sort, you will be prompted to undo the changes.