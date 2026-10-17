package images

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	"github.com/gen2brain/webp"
	"github.com/nefdevelop/godian/internal/domain"
	"golang.org/x/image/draw"
)

// lqipData genera un placeholder diminuto y difuminado como data URI (WebP).
func lqipData(img image.Image, b image.Rectangle) (string, error) {
	const w = 24
	h := b.Dy() * w / max(b.Dx(), 1)
	if h < 1 {
		h = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, b, draw.Over, nil)
	boxBlur(dst, 1)

	var buf bytes.Buffer
	if err := webp.Encode(&buf, dst, webp.Options{Quality: 30}); err != nil {
		return "", err
	}
	return "data:image/webp;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// makePreset genera la variante de un preset (con fit/fill, focal y blur).
func (p *Processor) makePreset(img image.Image, b image.Rectangle, dir, base, name string, pr domain.ImagePreset) (*Variant, error) {
	origW, origH := b.Dx(), b.Dy()
	W := pr.Width
	if W <= 0 {
		W = origW
	}

	var dst image.Image
	var H int
	if strings.EqualFold(pr.Fit, "fill") {
		H = pr.Height
		if H <= 0 {
			H = origH * W / max(origW, 1)
		}
		crop := cropRect(b, W, H, pr.Focal)
		rgba := image.NewRGBA(image.Rect(0, 0, W, H))
		draw.CatmullRom.Scale(rgba, rgba.Bounds(), img, crop, draw.Over, nil)
		dst = rgba
	} else {
		H = origH * W / max(origW, 1)
		if H < 1 {
			H = 1
		}
		rgba := image.NewRGBA(image.Rect(0, 0, W, H))
		draw.CatmullRom.Scale(rgba, rgba.Bounds(), img, b, draw.Over, nil)
		dst = rgba
	}
	if pr.Blur > 0 {
		if rgba, ok := dst.(*image.RGBA); ok {
			boxBlur(rgba, pr.Blur)
		}
	}

	q := pr.Quality
	if q <= 0 {
		q = 80
	}
	outRel := filepath.Join(dir, base+"-"+name+".webp")
	full := filepath.Join(p.outputDir, filepath.FromSlash(outRel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return nil, err
	}
	f, err := os.Create(full)
	if err != nil {
		return nil, err
	}
	if err := webp.Encode(f, dst, webp.Options{Quality: q}); err != nil {
		return nil, errors.Join(err, f.Close())
	}
	if err := f.Close(); err != nil {
		return nil, err
	}
	return &Variant{URL: "/" + filepath.ToSlash(outRel), Width: W, Height: H}, nil
}

// cropRect calcula el recorte para llenar W×H según el foco.
func cropRect(b image.Rectangle, W, H int, focal string) image.Rectangle {
	origW, origH := b.Dx(), b.Dy()
	target := float64(W) / float64(H)
	src := float64(origW) / float64(origH)
	if src > target {
		nw := int(float64(origH) * target)
		if nw > origW {
			nw = origW
		}
		off := focalOffset(origW-nw, focal)
		return image.Rect(b.Min.X+off, b.Min.Y, b.Min.X+off+nw, b.Max.Y)
	}
	nh := int(float64(origW) / target)
	if nh > origH {
		nh = origH
	}
	off := focalOffset(origH-nh, focal)
	return image.Rect(b.Min.X, b.Min.Y+off, b.Max.X, b.Min.Y+off+nh)
}

func focalOffset(total int, focal string) int {
	if total <= 0 {
		return 0
	}
	switch strings.ToLower(focal) {
	case "top", "left":
		return 0
	case "bottom", "right":
		return total
	default:
		return total / 2
	}
}

// clamp8 acota un entero al rango de uint8 (0-255).
func clamp8(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v) // #nosec G115 -- acotado a 0..255
}

// boxBlur aplica un desenfoque de caja (horizontal + vertical).
func boxBlur(img *image.RGBA, radius int) {
	if radius < 1 {
		return
	}
	b := img.Bounds()
	tmp := image.NewRGBA(b)

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			var r, g, bl, a, n int
			for dx := -radius; dx <= radius; dx++ {
				xx := x + dx
				if xx < b.Min.X || xx >= b.Max.X {
					continue
				}
				c := img.RGBAAt(xx, y)
				r, g, bl, a, n = r+int(c.R), g+int(c.G), bl+int(c.B), a+int(c.A), n+1
			}
			tmp.SetRGBA(x, y, color.RGBA{clamp8(r / n), clamp8(g / n), clamp8(bl / n), clamp8(a / n)})
		}
	}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			var r, g, bl, a, n int
			for dy := -radius; dy <= radius; dy++ {
				yy := y + dy
				if yy < b.Min.Y || yy >= b.Max.Y {
					continue
				}
				c := tmp.RGBAAt(x, yy)
				r, g, bl, a, n = r+int(c.R), g+int(c.G), bl+int(c.B), a+int(c.A), n+1
			}
			img.SetRGBA(x, y, color.RGBA{clamp8(r / n), clamp8(g / n), clamp8(bl / n), clamp8(a / n)})
		}
	}
}
