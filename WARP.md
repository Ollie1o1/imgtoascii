# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

Project overview
- A minimal Go CLI that converts images to ASCII using only the standard library. Module name: img2ascii (Go 1.21).
- Single entrypoint in main.go; no external deps, no tests at the time of writing.

Commands
- Build
  - Windows (PowerShell):
    - $env:GOOS="windows"; go build -o img2ascii.exe
  - macOS/Linux:
    - go build -o img2ascii
- Run (without building):
  - go run . -i ./path/to/image.jpg -w 100 --invert
- Run (built binary):
  - Windows: .\img2ascii.exe -i .\sample.jpg -w 120 > out.txt
  - macOS/Linux: ./img2ascii -i ./sample.jpg -w 120 > out.txt
- Format and lint
  - Format (simplify and write in-place): gofmt -s -w .
  - Vet (static checks): go vet ./...
  - Tidy modules (keep go.mod/go.sum clean): go mod tidy
- Tests
  - There are no tests in the repo currently. When tests are added:
    - Run all: go test ./...
    - Run a single test by name (regex): go test -run "^TestRenderASCII$" ./...

Architecture and flow
- CLI flags (flag package)
  - -i (required): input image path
  - -w (default 80): output width in characters
  - --invert: reverse the brightness-to-charset mapping
- Image decoding
  - Uses image.Decode with blank imports to register decoders: BMP, GIF, JPEG, PNG, TIFF (image/bmp, image/gif, image/jpeg, image/png, image/tiff).
- Processing pipeline (main.go)
  1) Parse flags and validate (-i required, -w > 0).
  2) Open file and decode into image.Image.
  3) Compute output dimensions: newW = -w; newH scales by charAspect (0.5) to account for character cell aspect ratio.
  4) renderASCII performs nearest-neighbor sampling from the source image into a newW x newH grid.
  5) For each sampled pixel, luminance8 computes 8-bit luma via Rec. 709 weights; this maps to a character from charset ("@%#*+=-:. "). If --invert is set, the charset is reversed once before mapping.
  6) Lines are buffered and written to stdout.
- Key functions
  - fail(error): centralized error reporting to stderr and os.Exit(1).
  - renderASCII(img, newW, newH, invert): sampling + luminance mapping into []string of rows.
  - luminance8(r,g,b): converts 16-bit RGBA to 8-bit luma using Rec. 709.

Usage quick reference
- img2ascii -i <path-to-image> [-w 80] [--invert]
- Character aspect ratio can be tuned by editing charAspect in main.go (defaults to 0.5).
