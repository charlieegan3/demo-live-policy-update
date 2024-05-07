package static

import (
	"bytes"
	"embed"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"

	"github.com/charlieegan3/demo-live-policy-update/pkg/server/handlers"
	"github.com/charlieegan3/demo-live-policy-update/pkg/utils"
)

//go:embed assets/*
var staticContent embed.FS

func BuildStaticHandler(opts *handlers.Options) (handler func(http.ResponseWriter, *http.Request)) {
	return func(w http.ResponseWriter, req *http.Request) {
		if !opts.DevMode {
			utils.SetCacheControl(w, "public, max-age=3600")
		}

		rootedReq := http.Request{
			URL: &url.URL{
				Path: strings.TrimPrefix(req.URL.Path, "/static/"),
			},
		}

		http.FileServer(http.FS(staticContent)).ServeHTTP(w, &rootedReq)
	}
}

func BuildCSSHandler(opts *handlers.Options) (string, func(http.ResponseWriter, *http.Request), error) {
	sourceFileOrder := []string{
		"tachyons.css",
		"styles.css",
	}

	var bs []byte

	for _, f := range sourceFileOrder {
		fileBytes, err := staticContent.ReadFile("assets/css/" + f)
		if err != nil {
			return "", nil, fmt.Errorf("failed to generate css: %s", err)
		}

		bs = append(bs, fileBytes...)
		bs = append(bs, []byte("\n")...)
	}

	in := bytes.NewBuffer(bs)
	out := bytes.NewBuffer([]byte{})
	if opts.DevMode {
		out = in
	} else {
		m := minify.New()
		m.AddFunc("application/css", css.Minify)

		if err := m.Minify("application/css", out, in); err != nil {
			return "", nil, fmt.Errorf("failed to generate css: %s", err)
		}
	}

	etag := utils.CRC32Hash(out.Bytes())

	return etag, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.Header().Set("Content-Type", "text/css")
		w.Header().Set("ETag", etag)
		if !opts.DevMode {
			utils.SetCacheControl(w, "public, max-age=31622400")
		}

		w.Write(out.Bytes())
	}, nil
}

func BuildJSHandler(opts *handlers.Options) (string, func(http.ResponseWriter, *http.Request), error) {
	sourceFileOrder := []string{
		"htmx.js",
		"script.js",
	}

	var bs []byte

	for _, f := range sourceFileOrder {
		fileBytes, err := staticContent.ReadFile("assets/js/" + f)
		if err != nil {
			return "", nil, fmt.Errorf("failed to generate css: %s", err)
		}

		bs = append(bs, fileBytes...)
		bs = append(bs, []byte("\n")...)
	}

	in := bytes.NewBuffer(bs)
	out := bytes.NewBuffer([]byte{})

	if opts.DevMode {
		out = in
	} else {
		m := minify.New()
		m.AddFunc("application/javascript", js.Minify)

		if err := m.Minify("application/javascript", out, in); err != nil {
			return "", nil, fmt.Errorf("failed to generate js: %s", err)
		}
	}

	etag := utils.CRC32Hash(out.Bytes())

	return etag, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("ETag", etag)
		if !opts.DevMode {
			utils.SetCacheControl(w, "public, max-age=31622400")
		}

		w.Write(out.Bytes())
	}, nil
}
