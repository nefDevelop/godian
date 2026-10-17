// Package images optimiza imágenes: genera variantes WebP redimensionadas y
// emite <picture> con srcset. Los archivos no imagen de AssetsDir se copian.
package images

import (
	"errors"
	"fmt"
	"html/template"
	"image"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gen2brain/webp"
	"github.com/nefdevelop/godian/internal/domain"
	"golang.org/x/image/draw"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

var widths = []int{480, 768, 1024, 1440}

var imageExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

// Variant es una variante generada.
type Variant struct {
	URL    string
	Width  int
	Height int
}

// Processor genera variantes de imagen y resuelve la etiqueta <picture>.
type Processor struct {
	assetsDir  string
	outputDir  string
	optimize   bool
	presets    map[string]domain.ImagePreset
	variants   map[string][]Variant
	avif       map[string][]Variant
	presetVars map[string]map[string]Variant
	lqip       map[string]string
	dimensions map[string][2]int
}

// New crea un Processor.
func New(assetsDir, outputDir string, optimize bool, presets map[string]domain.ImagePreset) *Processor {
	return &Processor{
		assetsDir:  assetsDir,
		outputDir:  outputDir,
		optimize:   optimize,
		presets:    presets,
		variants:   map[string][]Variant{},
		avif:       map[string][]Variant{},
		presetVars: map[string]map[string]Variant{},
		lqip:       map[string]string{},
		dimensions: map[string][2]int{},
	}
}

// Process recorre AssetsDir: copia los archivos y genera variantes WebP de las imágenes.
func (p *Processor) Process() error {
	info, err := os.Stat(p.assetsDir)
	if err != nil || !info.IsDir() {
		return nil
	}
	return filepath.WalkDir(p.assetsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(p.assetsDir, path)
		if err != nil {
			return err
		}
		ext := strings.ToLower(filepath.Ext(path))
		if p.optimize && imageExts[ext] {
			return p.processImage(path, rel)
		}
		return p.copyFile(path, rel)
	})
}

func (p *Processor) processImage(path, rel string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		if _, seekErr := f.Seek(0, io.SeekStart); seekErr == nil {
			img, err = webp.Decode(f)
		}
	}
	if err != nil {
		return nil // imagen no decodificable: se ignora
	}

	publicSrc := "/" + filepath.ToSlash(rel)
	b := img.Bounds()
	origW, origH := b.Dx(), b.Dy()
	p.dimensions[publicSrc] = [2]int{origW, origH}
	if data, err := lqipData(img, b); err == nil {
		p.lqip[publicSrc] = data
	}

	if err := p.copyFile(path, rel); err != nil {
		return err
	}

	dir := filepath.Dir(rel)
	base := strings.TrimSuffix(filepath.Base(rel), filepath.Ext(rel))

	vs := []Variant{{URL: "/" + filepath.ToSlash(filepath.Join(dir, base+".webp")), Width: origW}}
	if err := p.writeWebP(img, filepath.Join(dir, base+".webp")); err != nil {
		return err
	}

	var avs []Variant
	if avifSupported {
		fullRel := filepath.Join(dir, base+".avif")
		if err := p.writeAVIF(img, fullRel); err != nil {
			return err
		}
		avs = append(avs, Variant{URL: "/" + filepath.ToSlash(fullRel), Width: origW})
	}

	for _, w := range widths {
		if w >= origW {
			continue
		}
		h := origH * w / origW
		dst := image.NewRGBA(image.Rect(0, 0, w, h))
		draw.CatmullRom.Scale(dst, dst.Bounds(), img, b, draw.Over, nil)
		outRel := filepath.Join(dir, fmt.Sprintf("%s-%d.webp", base, w))
		if err := p.writeWebP(dst, outRel); err != nil {
			return err
		}
		vs = append(vs, Variant{URL: "/" + filepath.ToSlash(outRel), Width: w})

		if avifSupported {
			avifRel := filepath.Join(dir, fmt.Sprintf("%s-%d.avif", base, w))
			if err := p.writeAVIF(dst, avifRel); err != nil {
				return err
			}
			avs = append(avs, Variant{URL: "/" + filepath.ToSlash(avifRel), Width: w})
		}
	}

	sort.Slice(vs, func(i, j int) bool { return vs[i].Width < vs[j].Width })
	p.variants[publicSrc] = vs
	if len(avs) > 0 {
		sort.Slice(avs, func(i, j int) bool { return avs[i].Width < avs[j].Width })
		p.avif[publicSrc] = avs
	}

	if len(p.presets) > 0 {
		p.presetVars[publicSrc] = map[string]Variant{}
		for name, pr := range p.presets {
			v, err := p.makePreset(img, b, dir, base, name, pr)
			if err != nil {
				return err
			}
			if v != nil {
				p.presetVars[publicSrc][name] = *v
			}
		}
	}
	return nil
}

// Picture devuelve un <picture> con srcset WebP si hay variantes; si no, un <img>.
func (p *Processor) Picture(src, alt string) template.HTML {
	vs, ok := p.variants[src]
	if !ok {
		return template.HTML(p.imgTag(src, alt, 0, 0))
	}
	const sizes = ` sizes="(min-width: 64rem) 64rem, 100vw"`
	out := `<picture>`
	if avs, ok := p.avif[src]; ok {
		out += `<source type="image/avif" srcset="` + srcset(avs) + `"` + sizes + `>`
	}
	out += `<source type="image/webp" srcset="` + srcset(vs) + `"` + sizes + `>`
	out += p.imgTag(src, alt, 0, 0)
	out += `</picture>`
	return template.HTML(out)
}

// Preset devuelve un <picture> con una variante de preset (false si no existe).
func (p *Processor) Preset(src, alt, preset string) (template.HTML, bool) {
	byName, ok := p.presetVars[src]
	if !ok {
		return "", false
	}
	v, ok := byName[preset]
	if !ok {
		return "", false
	}
	out := `<picture><source type="image/webp" srcset="` + v.URL + ` ` + fmt.Sprintf("%d", v.Width) + `w">` +
		p.imgTag(src, alt, v.Width, v.Height) + `</picture>`
	return template.HTML(out), true
}

// imgTag construye la etiqueta <img> (con dimensiones y LQIP si se conocen).
func (p *Processor) imgTag(src, alt string, w, h int) string {
	if w == 0 || h == 0 {
		if d, ok := p.dimensions[src]; ok {
			w, h = d[0], d[1]
		}
	}
	s := `<img src="` + template.HTMLEscapeString(src) + `" alt="` + template.HTMLEscapeString(alt) + `" loading="lazy" decoding="async"`
	if w > 0 && h > 0 {
		s += fmt.Sprintf(` width="%d" height="%d"`, w, h)
	}
	if lq, ok := p.lqip[src]; ok {
		s += ` style="background-image:url(` + lq + `);background-size:cover;background-position:center"`
	}
	return s + `>`
}

func srcset(vs []Variant) string {
	var b strings.Builder
	for i, v := range vs {
		if i > 0 {
			b.WriteString(", ")
		}
		fmt.Fprintf(&b, "%s %dw", v.URL, v.Width)
	}
	return b.String()
}

func (p *Processor) writeAVIF(img image.Image, rel string) error {
	full := filepath.Join(p.outputDir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	f, err := os.Create(full)
	if err != nil {
		return err
	}
	if err := encodeAVIF(f, img); err != nil {
		return errors.Join(err, f.Close())
	}
	return f.Close()
}

func (p *Processor) writeWebP(img image.Image, rel string) error {
	full := filepath.Join(p.outputDir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	f, err := os.Create(full)
	if err != nil {
		return err
	}
	defer f.Close()
	return webp.Encode(f, img, webp.Options{Quality: 80})
}

func (p *Processor) copyFile(src, rel string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	full := filepath.Join(p.outputDir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, data, 0o644)
}
