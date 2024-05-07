package demo

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"

	"github.com/charlieegan3/demo-live-policy-update/pkg/opa"
	"github.com/charlieegan3/demo-live-policy-update/pkg/server/handlers"
)

func TestDemo(t *testing.T) {
	var err error

	modulePath := "policy/allow.rego"
	example1Mod := `package policy
import rego.v1
default allow := true
allow if input.name in {"alice", "bob"}
`

	exampleBundle := &bundle.Bundle{
		Manifest: bundle.Manifest{
			Revision: "1",
		},
		Modules: []bundle.ModuleFile{
			{
				URL:    modulePath,
				Path:   modulePath,
				Parsed: ast.MustParseModule(example1Mod),
				Raw:    []byte(example1Mod),
			},
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/vnd.openpolicyagent.bundles")

		w.Header().Set("etag", exampleBundle.Manifest.Revision)
		err = bundle.NewWriter(w).Write(*exampleBundle)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return

	}

	testServer := httptest.NewServer(http.HandlerFunc(handler))
	defer testServer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	m := opa.NewManager()
	err = m.Add(
		ctx,
		"example1",
		"example1",
		"example1-token",
		testServer.Listener.Addr().String(),
	)
	if err != nil {
		t.Fatalf("unexpected error adding OPA: %s", err)
	}

	h, err := NewDemoHandler(&handlers.Options{
		OPAManager: m,
	})
	if err != nil {
		t.Fatalf("unexpected error creating OPA show handler: %s", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/demo/example1", nil)
	h.ServeHTTP(rr, req)

	bs, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("unexpected error reading response body: %s", err)
	}

	if rr.Code != http.StatusOK {
		t.Log(string(bs))
		t.Fatalf("unexpected status code: %d", rr.Code)
	}

	bodyString := string(bs)

	if !strings.Contains(bodyString, "demo") {
		t.Fatalf("expected example1 to be present")
	}
}
