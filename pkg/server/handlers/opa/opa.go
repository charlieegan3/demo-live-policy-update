package opa

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/open-policy-agent/opa/sdk"

	"github.com/charlieegan3/demo-live-policy-update/pkg/server/handlers"
)

func NewOPACollectionHandler(opts *handlers.Options) (http.HandlerFunc, error) {
	if opts == nil || opts.OPAManager == nil {
		return nil, fmt.Errorf("opts and opa manager must be provided")
	}

	tmpl, err := template.ParseFS(
		handlers.Templates,
		"templates/opa/list.html",
		"templates/base.html",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %s", err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		if r.Method == http.MethodPost {
			err = r.ParseForm()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, err = w.Write([]byte(err.Error()))
				return
			}

			ref := r.PostFormValue("ref")
			if ref == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, err = w.Write([]byte("ref must be provided"))
				return
			}

			if r.Form.Get("_method") == "DELETE" {
				opts.OPAManager.Delete(r.Context(), ref)
				http.Redirect(w, r, "/opas", http.StatusSeeOther)
				return
			}

			systemID := r.Form.Get("system_id")
			if systemID == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, err = w.Write([]byte("system_id must be provided"))
				return
			}

			token := r.Form.Get("token")
			if token == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, err = w.Write([]byte("token must be provided"))
				return
			}

			endpoint := r.Form.Get("endpoint")
			if endpoint == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, err = w.Write([]byte("endpoint must be provided"))
				return
			}

			err = opts.OPAManager.Add(
				r.Context(),
				ref,
				systemID,
				token,
				endpoint,
			)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, err = w.Write([]byte(err.Error()))
				return
			}

			http.Redirect(w, r, fmt.Sprintf("/opas/%s", ref), http.StatusSeeOther)
		}

		buf := bytes.NewBuffer([]byte{})

		opas := opts.OPAManager.List()

		err = tmpl.ExecuteTemplate(buf, "base", struct {
			Opts *handlers.Options
			OPAs []string
		}{
			Opts: opts,
			OPAs: opas,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(err.Error()))
			return
		}

		_, err = io.Copy(w, buf)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(err.Error()))
			return
		}
	}, nil
}

func NewOPAShowHandler(opts *handlers.Options) (http.HandlerFunc, error) {
	if opts == nil || opts.OPAManager == nil {
		return nil, fmt.Errorf("opts and opa manager must be provided")
	}

	tmpl, err := template.ParseFS(
		handlers.Templates,
		"templates/opa/show.html",
		"templates/base.html",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %s", err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		buf := bytes.NewBuffer([]byte{})

		ref := strings.TrimPrefix(r.URL.Path, "/opas/")

		opa := opts.OPAManager.Get(ref)
		if opa == nil {
			w.WriteHeader(http.StatusNotFound)
			_, err = w.Write([]byte("opa not found"))
			return
		}

		err = tmpl.ExecuteTemplate(buf, "base", struct {
			Opts *handlers.Options
			OPA  *sdk.OPA
			Ref  string
		}{
			Opts: opts,
			OPA:  opa,
			Ref:  ref,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(err.Error()))
			return
		}

		_, err = io.Copy(w, buf)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(err.Error()))
			return
		}
	}, nil
}
