package demo

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/open-policy-agent/opa/sdk"

	"github.com/charlieegan3/demo-live-policy-update/pkg/server/handlers"
)

func NewDemoHandler(opts *handlers.Options) (http.HandlerFunc, error) {
	if opts == nil || opts.OPAManager == nil {
		return nil, fmt.Errorf("missing required options")
	}

	tmpl, err := template.ParseFS(
		handlers.Templates,
		"templates/demo/demo.html",
		"templates/base.html",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %s", err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ref := strings.TrimPrefix(r.URL.Path, "/demo/")

		opaInstance := opts.OPAManager.Get(ref)
		if opaInstance == nil {
			w.WriteHeader(http.StatusNotFound)
			_, err = w.Write([]byte("OPA instance not found"))
			return
		}

		name := "alice"
		if r.URL.Query().Get("name") != "" {
			name = r.URL.Query().Get("name")
		}

		dr, err := opaInstance.Decision(r.Context(), sdk.DecisionOptions{
			Path: "/policy/allow",
			Input: map[string]interface{}{
				"name": name,
			},
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(err.Error()))
			return
		}

		result, ok := dr.Result.(bool)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte("unexpected decision result"))
			return
		}

		buf := new(bytes.Buffer)

		err = tmpl.ExecuteTemplate(buf, "base", struct {
			Opts    *handlers.Options
			Name    string
			Path    string
			Allowed bool
		}{
			Opts:    opts,
			Name:    name,
			Path:    r.URL.Path,
			Allowed: result,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(err.Error()))
			return
		}

		w.Write(buf.Bytes())

		return
	}, nil
}
