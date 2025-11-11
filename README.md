# img2ascii

A lightweight CLI that converts images into ASCII art. Implemented in Go with only the standard library.

## Features
- Decodes common formats (PNG, JPEG, GIF, BMP, TIFF)
- Resizes using nearest-neighbor for speed
- Simple luminance-to-ASCII mapping with optional invert
- Interactive selection or multiple input methods
- No external dependencies

## Build
Requires Go (1.21+ recommended).

Windows (PowerShell):

```
$env:GOOS="windows"; go build -o img2ascii.exe
```

Other platforms:

```
go build -o img2ascii
```

## Usage

Basic (interactive when no input provided):
```
img2ascii [-w 80] [--invert]
```

Explicit file or directory:
```
img2ascii -i <path-to-image|directory> [-w 80] [--invert]
```

Glob pattern (non-recursive):
```
img2ascii --glob "*.png" [-w 80] [--invert]
```

Read a path from stdin (first non-empty line):
```
Get-Content path.txt | img2ascii --stdin
```

Disable prompts and take the first match automatically:
```
img2ascii --glob "*.jpg" --interactive=false
```

Examples (Windows):
```
# Print to terminal
./img2ascii.exe -i .\sample.jpg -w 100

# Save to a text file
./img2ascii.exe -i .\sample.jpg -w 120 > out.txt

# Interactive selection from current directory (no flags)
./img2ascii.exe
```

Flags:
- `-i`: input image path or directory (optional; prompts if omitted)
- `-glob`: glob to match images in the current or given directory (e.g. `*.png`)
- `-stdin`: read a path from stdin (first non-empty line)
- `-interactive` (default `true`): prompt when multiple images are found or no input provided
- `-w` (default 80): output width in characters
- `-invert`: invert the brightness mapping

## Notes
- Character aspect ratio is approximated; tweak `charAspect` in `main.go` for different terminals/fonts.
- Large images may take a moment to decode; resizing is O(width*height).
