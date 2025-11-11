package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"image"
	_ "image/bmp"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	_ "image/tiff"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	inPath := flag.String("i", "", "path to input image or directory (optional; interactive when omitted)")
	width := flag.Int("w", 80, "output width in characters")
	invert := flag.Bool("invert", false, "invert brightness mapping")
	glob := flag.String("glob", "", "optional glob to match images (e.g. *.png)")
	fromStdin := flag.Bool("stdin", false, "read an image path from stdin (first non-empty line)")
	interactive := flag.Bool("interactive", true, "prompt to choose when multiple images are found or no input provided")
	flag.Parse()

	if *width <= 0 {
		fail(errors.New("-w must be > 0"))
	}

	// Resolve which image to open.
	imgPath, err := resolveInput(*inPath, *glob, *fromStdin, *interactive)
	if err != nil {
		fail(err)
	}
	if imgPath == "" {
		fail(errors.New("no image selected"))
	}

	f, err := os.Open(imgPath)
	if err != nil {
		fail(fmt.Errorf("open: %w", err))
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		fail(fmt.Errorf("decode: %w", err))
	}

	w := img.Bounds().Dx()
	h := img.Bounds().Dy()
	if w == 0 || h == 0 {
		fail(errors.New("image has zero dimension"))
	}

	// Adjust height to account for character aspect ratio (chars are taller than wide).
	charAspect := 0.5 // tweak to taste (smaller = fewer rows)
	newW := *width
	newH := int(math.Max(1, math.Round(float64(h)*charAspect*float64(newW)/float64(w))))

	ascii := renderASCII(img, newW, newH, *invert)

	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	for _, row := range ascii {
		out.WriteString(row)
		out.WriteByte('\n')
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

// resolveInput determines which image file to use based on flags and environment.
func resolveInput(inPath, glob string, fromStdin, interactive bool) (string, error) {
	// 1) stdin takes precedence
	if fromStdin {
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			p := strings.TrimSpace(s.Text())
			if p == "" {
				continue
			}
			if fileExists(p) && isImageExt(p) {
				return p, nil
			}
			// If directory, try to pick from it
			if isDir(p) {
				cands := imagesInDir(p)
				if len(cands) > 0 {
					if interactive {
						return pickInteractive(cands)
					}
					return cands[0], nil
				}
			}
			// If file exists but not an image, keep scanning
		}
		if err := s.Err(); err != nil {
			return "", fmt.Errorf("stdin: %w", err)
		}
		return "", errors.New("no usable path from stdin")
	}

	// 2) explicit path
	if inPath != "" {
		if isDir(inPath) {
			cands := imagesInDir(inPath)
			cands = filterByGlob(cands, glob)
			if len(cands) == 0 {
				return "", fmt.Errorf("no images found in directory: %s", inPath)
			}
			if interactive {
				return pickInteractive(cands)
			}
			return cands[0], nil
		}
		if fileExists(inPath) {
			if isImageExt(inPath) {
				return inPath, nil
			}
			return "", fmt.Errorf("not an image: %s", inPath)
		}
		return "", fmt.Errorf("path not found: %s", inPath)
	}

	// 3) glob across current directory (non-recursive)
	if glob != "" {
		matches, _ := filepath.Glob(glob)
		cands := filterImages(matches)
		if len(cands) == 0 {
			return "", fmt.Errorf("glob matched no images: %s", glob)
		}
		if interactive {
			return pickInteractive(cands)
		}
		return cands[0], nil
	}

	// 4) interactive from current directory by default
	cands := imagesInDir(".")
	if len(cands) == 0 {
		return "", errors.New("no images found in current directory; pass -i, --glob, or --stdin")
	}
	if interactive {
		return pickInteractive(cands)
	}
	return cands[0], nil
}

func isDir(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

func isImageExt(p string) bool {
	ext := strings.ToLower(filepath.Ext(p))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".tif", ".tiff":
		return true
	default:
		return false
	}
}

func imagesInDir(dir string) []string {
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range ents {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		p := filepath.Join(dir, name)
		if isImageExt(p) {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}

func filterImages(paths []string) []string {
	var out []string
	for _, p := range paths {
		if isImageExt(p) && fileExists(p) {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}

func filterByGlob(paths []string, glob string) []string {
	if glob == "" {
		return paths
	}
	// If glob is present, apply it relative to each file's base name.
	var out []string
	for _, p := range paths {
		match, err := filepath.Match(glob, filepath.Base(p))
		if err == nil && match {
			out = append(out, p)
		}
	}
	return out
}

func pickInteractive(cands []string) (string, error) {
	in := bufio.NewReader(os.Stdin)
	fmt.Fprintln(os.Stderr, "Select an image to render:")
	for i, c := range cands {
		fmt.Fprintf(os.Stderr, "  %d) %s\n", i+1, c)
	}
	fmt.Fprintf(os.Stderr, "Enter number (1-%d) or a path (default 1): ", len(cands))
	line, _ := in.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return cands[0], nil
	}
	// If numeric
	if idx, err := parseIndex(line, len(cands)); err == nil {
		return cands[idx], nil
	}
	// Otherwise treat as path
	if fileExists(line) && isImageExt(line) {
		return line, nil
	}
	if isDir(line) {
		d := imagesInDir(line)
		if len(d) > 0 {
			return d[0], nil
		}
	}
	return "", errors.New("invalid selection")
}

func parseIndex(s string, n int) (int, error) {
	// 1-based to 0-based index
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	if err != nil {
		return 0, err
	}
	i--
	if i < 0 || i >= n {
		return 0, errors.New("out of range")
	}
	return i, nil
}

func renderASCII(img image.Image, newW, newH int, invert bool) []string {
	// From dark to light
	charset := []rune("@%#*+=-:. ")
	if invert {
		// reverse
		for i, j := 0, len(charset)-1; i < j; i, j = i+1, j-1 {
			charset[i], charset[j] = charset[j], charset[i]
		}
	}

	origW := img.Bounds().Dx()
	origH := img.Bounds().Dy()

	rows := make([]string, newH)
	for y := 0; y < newH; y++ {
		// nearest-neighbor sampling
		sy := int(float64(y) * float64(origH) / float64(newH))
		if sy >= origH {
			sy = origH - 1
		}
		buf := make([]rune, newW)
		for x := 0; x < newW; x++ {
			sx := int(float64(x) * float64(origW) / float64(newW))
			if sx >= origW {
				sx = origW - 1
			}
			r, g, b, _ := img.At(img.Bounds().Min.X+sx, img.Bounds().Min.Y+sy).RGBA()
			lum := luminance8(r, g, b) // 0..255
			idx := int(math.Round(float64(lum) * float64(len(charset)-1) / 255.0))
			buf[x] = charset[idx]
		}
		rows[y] = string(buf)
	}
	return rows
}

func luminance8(r, g, b uint32) uint8 {
	// Convert 16-bit per channel to 8-bit and compute luma.
	r8 := float64(r >> 8)
	g8 := float64(g >> 8)
	b8 := float64(b >> 8)
	// Rec. 709 luma approximation
	l := 0.2126*r8 + 0.7152*g8 + 0.0722*b8
	if l < 0 {
		l = 0
	} else if l > 255 {
		l = 255
	}
	return uint8(l + 0.5)
}
