package templates

const DefaultPrompt = `You are a file renaming engine. Your goal is to transform filenames based on the provided input and any specific user instructions.

Input: A JSON array of filename strings.
Output: A JSON object containing a "filenames" key with an array of transformed filename strings.

### CRITICAL RULES:
1. Return **ONLY** the raw JSON object. No Markdown, no prose, no explanation.
2. The "filenames" array MUST have the exact same number of elements as the input.
3. Preserve array order and file extensions exactly.
4. If no specific user instructions are provided, apply standard Title_Snake_Case formatting (e.g., "my file.txt" -> "My_File.txt").
5. Handle collisions by appending "_v1", "_v2", etc. if necessary.

PAYLOAD:`

var DefaultConfig = `// Config file for 'sorta'
// Config version: v0.4.2
//
// Each line defines how files should be sorted.
// Format: folderName = key1,key2,key3
//
// - folderName is the target folder for those files.
// - key1, key2, key3, etc are keywords to match in file names.
// - You can list one or many keywords after the '='.
// - Lines starting with '//' are comments and ignored.
// - Add a ! followed by an ignore pattern to skip paths/files while scanning.
//   Examples: !node_modules, !*.tmp, !archive/*.zip
// - Ignore rules are also loaded from .sortaignore, .sorta/ignore, and ~/.sorta/ignore.
// - * as a keyword matches all filenames which don't contain the other keywords
// - . as a foldernames means the root folder that you passed to sorta.
// - To flatten the subfolder tree, use . = *
// - Use regex for kewyords. Wrap your expression with: regex(). No quotes are required.
// - foldername can also be a relative folderpath. e.g. foo/bar/oof = rab creates a folder tree.
//
// Example:
//
// Finance=invoice,bill,txt
// Music=track,song
// Study=notes,book
// 2024-Papers=regex(^PAP.*2024$)
// others=*
//
// Important folder that sorta won't scan:
// !my-secret-folder`
