package main

import (
	"github.com/electr1fy0/sorta/cmd"
)

// TODO:
// ─── FEATURES ──────────────────────────────────────────────────────────────
// • Blacklist / Whitelist folders (gitignore-style)
// • Regex support in config
// • Use MIME type detection instead of file extensions
// • Interactive mode:
//   - Ask actions for unmatched files
//   - Fast mode: prompt once per folder
// • Add concurrency for system calls
// • Option to not go recursive / blacklist all subfolders

//
// ─── SORTING ────────────────────────────────────────────────────
// • Sort by average file size
// • Sort by file data
// • More expressive expressive summary
//
// ─── CACHE & LOGGING ───────────────────────────────────────────────────────
// • Cache file checksums
// • Tree-style logging

func main() {
	cmd.Execute()
}
