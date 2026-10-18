// Package cli implementa la interfaz de línea de comandos de godian.
package cli

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/nefdevelop/godian/internal/application"
	"github.com/nefdevelop/godian/internal/domain"
	"github.com/nefdevelop/godian/internal/i18n"
	"github.com/nefdevelop/godian/internal/infrastructure/check"
	"github.com/nefdevelop/godian/internal/infrastructure/config"
	"github.com/nefdevelop/godian/internal/infrastructure/content"
	"github.com/nefdevelop/godian/internal/infrastructure/data"
	"github.com/nefdevelop/godian/internal/infrastructure/feed"
	"github.com/nefdevelop/godian/internal/infrastructure/fonts"
	"github.com/nefdevelop/godian/internal/infrastructure/images"
	"github.com/nefdevelop/godian/internal/infrastructure/islands"
	"github.com/nefdevelop/godian/internal/infrastructure/markdown"
	"github.com/nefdevelop/godian/internal/infrastructure/og"
	"github.com/nefdevelop/godian/internal/infrastructure/output"
	"github.com/nefdevelop/godian/internal/infrastructure/render"
	"github.com/nefdevelop/godian/internal/infrastructure/theme"
	"github.com/nefdevelop/godian/internal/infrastructure/verify"
	"github.com/nefdevelop/godian/internal/presentation/server"
)

// Version se puede sobrescribir en build con -ldflags "-X ...cli.Version=x.y.z".
var Version = "dev"

const usage = `godian — generador de blog estático en Go

Uso:
  godian build [-config site.yaml] [-drafts]      Genera el sitio estático
  godian serve [-config site.yaml] [-addr :4321]  Dev server con recarga
  godian new [-collection blog] [-title "..."] <slug>  Crea un post nuevo
  godian new-theme <dir>                          Copia el tema por defecto a <dir>
  godian check [-config site.yaml]                Valida el contenido
  godian version                                  Muestra la versión
`

// Run ejecuta el comando indicado y devuelve el código de salida.
func Run(args []string) int {
	if len(args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		return 2
	}
	switch args[1] {
	case "build":
		return runBuild(args[2:])
	case "serve":
		return runServe(args[2:])
	case "new":
		return runNew(args[2:])
	case "new-theme":
		return runNewTheme(args[2:])
	case "check":
		return runCheck(args[2:])
	case "version", "-v", "--version":
		fmt.Printf("godian %s\n", Version)
		return 0
	case "help", "-h", "--help":
		fmt.Print(usage)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "comando desconocido: %q\n\n%s", args[1], usage)
		return 2
	}
}

func runBuild(args []string) int {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	configPath := fs.String("config", "site.yaml", "ruta a la configuración del sitio")
	drafts := fs.Bool("drafts", false, "incluir borradores en la salida")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error cargando configuración: %v\n", err)
		return 1
	}
	summary, err := generate(cfg, *drafts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error en el build: %v\n", err)
		return 1
	}
	printSummary(cfg, summary)
	for _, w := range summary.Warnings {
		fmt.Fprintf(os.Stderr, "aviso: %s\n", w)
	}
	if len(summary.Errors) > 0 {
		for _, e := range summary.Errors {
			fmt.Fprintf(os.Stderr, "referencia rota: %s\n", e)
		}
		return 1
	}
	return 0
}

func runServe(args []string) int {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	configPath := fs.String("config", "site.yaml", "ruta a la configuración del sitio")
	addr := fs.String("addr", ":4321", "dirección de escucha")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error cargando configuración: %v\n", err)
		return 1
	}
	if summary, err := generate(cfg, true); err != nil {
		fmt.Fprintf(os.Stderr, "error en el build inicial: %v\n", err)
		return 1
	} else {
		for _, e := range summary.Errors {
			fmt.Fprintf(os.Stderr, "referencia rota: %s\n", e)
		}
	}

	watch := []string{cfg.Build.ContentDir, cfg.Build.ThemeDir, cfg.Build.AssetsDir, *configPath}
	rebuild := func() error {
		fresh, err := config.Load(*configPath)
		if err != nil {
			return err
		}
		_, err = generate(fresh, true)
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := server.New(*addr, filepath.FromSlash(cfg.Build.OutputDir), watch, rebuild)
	if err := srv.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "servidor: %v\n", err)
		return 1
	}
	return 0
}

const newPostTemplate = `---
title: %s
published: %s
description: ""
tags: []
categories: []
author: []
draft: true
---

`

func runNew(args []string) int {
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	configPath := fs.String("config", "site.yaml", "ruta a la configuración del sitio")
	collection := fs.String("collection", "blog", "colección destino")
	title := fs.String("title", "", "título del post (por defecto, derivado del slug)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	rest := fs.Args()
	if len(rest) < 1 {
		fmt.Fprintln(os.Stderr, `uso: godian new [-config site.yaml] [-collection blog] [-title "..."] <slug>`)
		return 2
	}
	slug := rest[0]

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error cargando configuración: %v\n", err)
		return 1
	}

	folder := *collection
	for _, c := range cfg.Collections {
		if c.Name == *collection && c.Folder != "" {
			folder = c.Folder
			break
		}
	}

	dir := filepath.Join(filepath.FromSlash(cfg.Build.ContentDir), filepath.FromSlash(folder))
	path := filepath.Join(dir, slug+".md")
	if _, statErr := os.Stat(path); statErr == nil {
		fmt.Fprintf(os.Stderr, "ya existe: %s\n", path)
		return 1
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "error creando carpeta: %v\n", err)
		return 1
	}

	t := *title
	if t == "" {
		t = strings.ReplaceAll(slug, "-", " ")
	}
	content := fmt.Sprintf(newPostTemplate, t, time.Now().Format("2006-01-02"))
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "error escribiendo: %v\n", err)
		return 1
	}
	fmt.Println("creado:", path)
	return 0
}

func runNewTheme(args []string) int {
	fs := flag.NewFlagSet("new-theme", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	rest := fs.Args()
	if len(rest) < 1 {
		fmt.Fprintln(os.Stderr, "uso: godian new-theme <dir>")
		return 2
	}
	dir := rest[0]

	if entries, err := os.ReadDir(dir); err == nil && len(entries) > 0 {
		fmt.Fprintf(os.Stderr, "el directorio ya existe y no está vacío: %s\n", dir)
		return 1
	}
	if err := theme.Scaffold(dir); err != nil {
		fmt.Fprintf(os.Stderr, "error copiando el tema: %v\n", err)
		return 1
	}
	fmt.Println("tema creado en:", dir)
	fmt.Printf("Apunta build.theme_dir a %q en site.yaml para usarlo.\n", dir)
	return 0
}

// generate ejecuta el pipeline completo con la configuración dada.
func generate(cfg domain.Config, drafts bool) (application.Summary, error) {
	th, err := theme.Load(cfg.Build.ThemeDir)
	if err != nil {
		return application.Summary{}, fmt.Errorf("cargando tema: %w", err)
	}

	tmplFS, err := th.TemplateFS()
	if err != nil {
		return application.Summary{}, err
	}

	lang := cfg.Site.Language
	if lang == "" {
		lang = "en"
	}
	var requiredKeys []string
	for _, tc := range cfg.Taxonomies {
		requiredKeys = append(requiredKeys, "taxonomy."+tc.Name)
	}
	dict, i18nWarnings := i18n.Load(filepath.FromSlash(cfg.Build.I18nDir), lang, requiredKeys)
	tr := func(key string) string {
		if v, ok := dict[key]; ok && v != "" {
			return v
		}
		return i18n.T(lang, key)
	}

	reg := islands.New(th.FS())
	funcs := template.FuncMap{
		"t": func(key string) string {
			if v, ok := dict[key]; ok && v != "" {
				return v
			}
			return i18n.T(lang, key)
		},
		"dict": dictFunc,
		"island": func(name string, props ...map[string]string) (template.HTML, error) {
			return reg.Island(name, firstMap(props))
		},
		"partial": func(name string, props ...map[string]string) (template.HTML, error) {
			return reg.Partial(name, firstMap(props))
		},
	}
	renderer, err := render.New(tmplFS, funcs)
	if err != nil {
		return application.Summary{}, fmt.Errorf("compilando plantillas: %w", err)
	}
	reg.SetRenderer(renderer)

	assetsFS, err := th.AssetsFS()
	if err != nil {
		return application.Summary{}, err
	}

	siteData, err := data.Load(filepath.FromSlash(cfg.Build.DataDir))
	if err != nil {
		return application.Summary{}, fmt.Errorf("cargando data/: %w", err)
	}

	imagesProc := images.New(cfg.Build.AssetsDir, filepath.FromSlash(cfg.Build.OutputDir), cfg.Build.Optimize, cfg.Build.ImagePresets,
		images.WithImageAttrs(cfg.Render.Images.Loading, cfg.Render.Images.Decoding))

	// Índice de enlaces (Obsidian): wikilinks entre colecciones, transclusión y backlinks.
	linkIndex := application.NewLinkIndex()
	plainMD := markdown.New(
		markdown.WithImageResolver(imagesProc.Picture),
		markdown.WithWikilinkResolver(linkIndex.Wikilink),
		markdown.WithTranslator(tr),
		markdown.WithRenderConfig(cfg.Render),
	)
	linkIndex.SetRenderer(plainMD)

	md := markdown.New(
		markdown.WithImageResolver(imagesProc.Picture),
		markdown.WithWikilinkResolver(linkIndex.Wikilink),
		markdown.WithEmbedResolver(linkIndex.Embed),
		markdown.WithTranslator(tr),
		markdown.WithRenderConfig(cfg.Render),
		markdown.WithImageShortcode(func(args map[string]string) (template.HTML, error) {
			src := args["src"]
			if src == "" {
				return "", fmt.Errorf("{{image}}: falta src")
			}
			alt := args["alt"]
			if preset := args["preset"]; preset != "" {
				if h, ok := imagesProc.Preset(src, alt, preset); ok {
					return h, nil
				}
			}
			return imagesProc.Picture(src, alt), nil
		}),
		markdown.WithIslandResolver(reg.Resolve),
		markdown.WithBlockResolver(reg.ResolveBlock),
	)

	ogGen, err := og.New(filepath.FromSlash(cfg.Build.OutputDir))
	if err != nil {
		return application.Summary{}, fmt.Errorf("preparando OG: %w", err)
	}

	clean := cfg.Build.Clean &&
		cfg.Build.OutputDir != "" &&
		cfg.Build.OutputDir != cfg.Build.ContentDir &&
		cfg.Build.OutputDir != cfg.Build.ThemeDir

	opts := application.Options{
		Site:          cfg.Site,
		Taxonomies:    cfg.Taxonomies,
		Giscus:        cfg.Giscus,
		Nav:           cfg.Nav,
		Social:        cfg.Social,
		Support:       cfg.Support,
		Analytics:     cfg.Analytics,
		Data:          siteData,
		Links:         linkIndex,
		I18n:          dict,
		CSS:           th.Manifest.Assets.CSS,
		JS:            th.Manifest.Assets.JS,
		AssetsFS:      assetsFS,
		Images:        imagesProc,
		OG:            ogGen,
		IncludeDrafts: drafts,
		Minify:        cfg.Build.Minify,
		Fingerprint:   cfg.Build.Fingerprint,
		Clean:         clean,
	}

	// Fuentes autohospedadas: se descargan en el build; si falla, se avisa y sigue.
	if len(cfg.Fonts) > 0 {
		fres, ferr := fonts.Fetch(&http.Client{Timeout: 30 * time.Second}, fontSpecs(cfg.Fonts), "")
		if ferr != nil {
			fmt.Fprintf(os.Stderr, "aviso: no se pudieron descargar las fuentes: %v\n", ferr)
		} else {
			opts.FontsCSS = fres.CSS
			opts.FontFiles = fres.Files
		}
	}

	builder := application.NewBuild(
		content.NewLoader(cfg),
		md,
		renderer,
		output.NewWriter(filepath.FromSlash(cfg.Build.OutputDir)),
		feed.New(),
		opts,
	)

	summary, err := builder.Run()
	if err != nil {
		return summary, err
	}
	summary.Warnings = append(summary.Warnings, i18nWarnings...)
	if problems, verr := verify.Dist(filepath.FromSlash(cfg.Build.OutputDir)); verr == nil {
		summary.Errors = problems
	}
	return summary, nil
}

// fontSpecs convierte la config de fuentes al modelo del paquete fonts.
func fontSpecs(cfg []domain.FontConfig) []fonts.Spec {
	out := make([]fonts.Spec, 0, len(cfg))
	for _, f := range cfg {
		out = append(out, fonts.Spec{
			Family:  f.Family,
			Weights: f.Weights,
			Styles:  f.Styles,
			Subsets: f.Subsets,
			Display: f.Display,
		})
	}
	return out
}

func runCheck(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	configPath := fs.String("config", "site.yaml", "ruta a la configuración del sitio")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error cargando configuración: %v\n", err)
		return 1
	}
	c, err := content.NewLoader(cfg).Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error cargando contenido: %v\n", err)
		return 1
	}
	problems := check.Run(cfg, c)
	for _, p := range problems {
		fmt.Fprintln(os.Stderr, "problema: "+p)
	}
	if len(problems) > 0 {
		fmt.Fprintf(os.Stderr, "%d problema(s)\n", len(problems))
		return 1
	}
	fmt.Println("check: OK")
	return 0
}

func firstMap(maps []map[string]string) map[string]string {
	if len(maps) > 0 && maps[0] != nil {
		return maps[0]
	}
	return map[string]string{}
}

// dictFunc construye un map[string]string desde pares clave/valor.
func dictFunc(values ...any) (map[string]string, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("dict: número impar de argumentos")
	}
	m := make(map[string]string, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		k, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict: la clave debe ser string")
		}
		m[k] = fmt.Sprintf("%v", values[i+1])
	}
	return m, nil
}

func printSummary(cfg domain.Config, s application.Summary) {
	title := cfg.Site.Title
	if title == "" {
		title = "(sin título)"
	}
	fmt.Printf("Sitio: %s\n", title)
	fmt.Printf("Contenido: %s → %s\n", cfg.Build.ContentDir, cfg.Build.OutputDir)
	fmt.Printf("  posts:     %d (%d publicados, %d programados, %d borradores)\n", s.Posts, s.Published, s.Scheduled, s.Drafts)
	fmt.Printf("  docs:      %d\n", s.Docs)
	fmt.Printf("  authors:   %d\n", s.Authors)
	fmt.Printf("  series:    %d\n", s.Series)
	fmt.Printf("  pages:     %d\n", s.Pages)
}
