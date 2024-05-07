package mux

import (
	"fmt"
	"net/http"

	"github.com/charlieegan3/demo-live-policy-update/pkg/server/handlers"
	"github.com/charlieegan3/demo-live-policy-update/pkg/server/handlers/demo"
	"github.com/charlieegan3/demo-live-policy-update/pkg/server/handlers/index"
	"github.com/charlieegan3/demo-live-policy-update/pkg/server/handlers/opa"
	"github.com/charlieegan3/demo-live-policy-update/pkg/server/handlers/static"
)

func NewMux(opts *handlers.Options) (*http.ServeMux, error) {

	mux := http.NewServeMux()

	stylesEtag, stylesHandler, err := static.BuildCSSHandler(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to build styles handler: %s", err)
	}

	scriptETag, scriptHandler, err := static.BuildJSHandler(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to build script handler: %s", err)
	}

	opts.EtagStyles = stylesEtag
	opts.EtagScript = scriptETag

	mux.Handle("/script.js", http.HandlerFunc(scriptHandler))
	mux.Handle("/styles.css", http.HandlerFunc(stylesHandler))

	osh, err := opa.NewOPAShowHandler(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to build opa show handler: %s", err)
	}
	mux.Handle("/opas/", osh)

	och, err := opa.NewOPACollectionHandler(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to build opa list handler: %s", err)
	}
	mux.Handle("/opas", och)

	dh, err := demo.NewDemoHandler(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to build demo handler: %s", err)
	}
	if dh != nil {
		mux.Handle("/demo/", dh)
	}

	mux.Handle("/", http.HandlerFunc(index.IndexHandler))

	return mux, nil
}
