package index

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/charlieegan3/demo-live-policy-update/pkg/server/handlers"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(
		handlers.Templates,
		"templates/index.html",
		"templates/base.html",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("failed to parse templates: %s", err)))
		return
	}

	w.WriteHeader(http.StatusOK)
	err = tmpl.ExecuteTemplate(w, "base", nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("failed to render template: %s", err)))
		return
	}
}
