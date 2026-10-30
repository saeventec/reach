# .gitattributes

## What it is

A file that tells git how to treat certain files. We use it for one thing: **line endings**.

Every line in a text file ends with an invisible character. Mac and Linux use `LF`, Windows traditionally uses `CRLF`. If the two get mixed, git shows every line as changed even when nothing really changed.

## Why we use it

You work on both Mac and Windows. This file makes git store all text files with `LF`, so diffs stay clean.

## Tiny example

Our whole file:

```gitattributes
* text=auto eol=lf
```

- `*` means "every file".
- `text=auto` lets git detect which files are text (images and other binary files are left alone).
- `eol=lf` means "end of line: LF", also in your working folder.

## Alternatives

- **The `core.autocrlf` git setting:** set by each person on their own machine, so easy to forget. The file in the repo works for everyone automatically.

## Learn more

- https://git-scm.com/docs/gitattributes
