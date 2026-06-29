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

var DefaultConfig = `rules:
  - folder: Finance
    keywords: [invoice, bill, .pdf]

  - folder: Music
    keywords: [track, song, .mp3]

  - folder: Images
    keywords: [.jpg, .png, .gif]

  - folder: Archive
    keywords: [.zip, .tar.gz]

  - folder: Others
    catch_all: true

ignore:
  - "*.tmp"
`
