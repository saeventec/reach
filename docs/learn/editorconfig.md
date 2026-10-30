# .editorconfig

## What it is

A small file with basic editor settings (character encoding, indentation, line endings) that most editors read automatically. VS Code needs the EditorConfig extension, which is in our recommended extensions.

## Why we use it

Every file gets the same basic formatting, on every computer and in every editor.

## Tiny example

Our file, without its comments:

```ini
root = true

[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true
indent_style = space
indent_size = 2
```

- `root = true`: this is the top-level file, so don't look in parent folders.
- `[*]`: the settings below apply to all files.
- `charset = utf-8`: a text encoding that supports every language.
- `end_of_line = lf`: line endings, the same as in `.gitattributes`.
- `insert_final_newline = true`: every file ends with a newline (a common standard).
- `trim_trailing_whitespace = true`: remove invisible spaces at the end of lines.
- `indent_style = space` and `indent_size = 2`: indent with 2 spaces, not tabs.

Comments in this file must be on their own line, starting with `#`.

Later, we add a `[*.go]` section with `indent_style = tab`, because Go's formatter uses tabs.

## Alternatives

- **Editor-specific settings** (like `.vscode/settings.json`): only work in one editor.
- **Formatters** (Prettier, gofmt): they do much more, and we use them too. EditorConfig covers the basics for every file type.

## Learn more

- https://editorconfig.org
