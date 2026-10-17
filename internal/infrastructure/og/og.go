// Package og genera imágenes Open Graph de forma nativa (sin cgo).
package og

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/nefdevelop/godian/internal/domain"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	width  = 1200
	height = 630
)

var (
	colBG     = color.RGBA{0x0b, 0x0b, 0x0d, 0xff}
	colFG     = color.RGBA{0xe5, 0xe7, 0xeb, 0xff}
	colMuted  = color.RGBA{0x9c, 0xa3, 0xaf, 0xff}
	colAccent = color.RGBA{0x60, 0xa5, 0xfa, 0xff}
)

// Generator renderiza imágenes OG.
type Generator struct {
	outputDir string
	regular   *opentype.Font
	bold      *opentype.Font
}

// New prepara el generador (parsea las fuentes embebidas).
func New(outputDir string) (*Generator, error) {
	regular, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil, err
	}
	bold, err := opentype.Parse(gobold.TTF)
	if err != nil {
		return nil, err
	}
	return &Generator{outputDir: outputDir, regular: regular, bold: bold}, nil
}

// Generate crea una imagen OG por cada post en <outputDir>/og/<slug>.png.
func (g *Generator) Generate(site domain.SiteConfig, posts []domain.BlogPost) error {
	dir := filepath.Join(g.outputDir, "og")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, p := range posts {
		img, err := g.drawCard(site, site.Title, p.Title, "")
		if err != nil {
			return err
		}
		if err := writePNG(filepath.Join(dir, p.Slug+".png"), img); err != nil {
			return err
		}
	}

	// Imagen OG por defecto (home y páginas sin OG propio).
	def, err := g.drawCard(site, host(site.URL), site.Title, site.Description)
	if err != nil {
		return err
	}
	return writePNG(filepath.Join(dir, "default.png"), def)
}

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		return errors.Join(err, f.Close())
	}
	return f.Close()
}

func host(rawURL string) string {
	s := strings.TrimPrefix(strings.TrimPrefix(rawURL, "https://"), "http://")
	return strings.TrimRight(s, "/")
}

func (g *Generator) drawCard(site domain.SiteConfig, eyebrow, title, subtitle string) (image.Image, error) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{colBG}, image.Point{}, draw.Src)

	// Barra de acento a la izquierda.
	draw.Draw(img, image.Rect(0, 0, 16, height), &image.Uniform{colAccent}, image.Point{}, draw.Src)

	eyebrowFace, err := g.face(g.bold, 30)
	if err != nil {
		return nil, err
	}
	titleFace, err := g.face(g.bold, 58)
	if err != nil {
		return nil, err
	}
	subFace, err := g.face(g.regular, 28)
	if err != nil {
		return nil, err
	}
	footFace, err := g.face(g.regular, 26)
	if err != nil {
		return nil, err
	}
	defer eyebrowFace.Close()
	defer titleFace.Close()
	defer subFace.Close()
	defer footFace.Close()

	if eyebrow != "" {
		drawText(img, eyebrowFace, colAccent, 80, 120, eyebrow)
	}

	y := 240
	for _, line := range wrap(titleFace, title, width-160) {
		drawText(img, titleFace, colFG, 80, y, line)
		y += 74
	}
	if subtitle != "" {
		y += 10
		for _, line := range wrap(subFace, subtitle, width-160) {
			drawText(img, subFace, colMuted, 80, y, line)
			y += 40
		}
	}

	footer := host(site.URL)
	if footer == "" {
		footer = site.Title
	}
	drawText(img, footFace, colMuted, 80, height-80, footer)

	return img, nil
}

func (g *Generator) face(f *opentype.Font, size float64) (font.Face, error) {
	return opentype.NewFace(f, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
}

func drawText(img *image.RGBA, face font.Face, col color.Color, x, y int, s string) {
	d := &font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{col},
		Face: face,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(s)
}

func wrap(face font.Face, text string, maxWidth int) []string {
	words := strings.Fields(text)
	var lines []string
	var line string
	for _, w := range words {
		candidate := w
		if line != "" {
			candidate = line + " " + w
		}
		if font.MeasureString(face, candidate).Ceil() > maxWidth && line != "" {
			lines = append(lines, line)
			line = w
		} else {
			line = candidate
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		lines = []string{text}
	}
	return lines
}
