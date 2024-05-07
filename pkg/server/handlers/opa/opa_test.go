package opa

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"

	"github.com/charlieegan3/demo-live-policy-update/pkg/opa"
	"github.com/charlieegan3/demo-live-policy-update/pkg/server/handlers"
)

func TestCreateOPA(t *testing.T) {
	var err error

	modulePath := "policy/allow.rego"
	example1Mod := `package policy
default allow := true`

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

	m := opa.NewManager()
	h, err := NewOPACollectionHandler(&handlers.Options{
		OPAManager: m,
	})
	if err != nil {
		t.Fatalf("unexpected error creating OPA list handler: %s", err)
	}

	rr := httptest.NewRecorder()

	p := url.Values{}
	p.Add("ref", "styra-engineer-1")
	p.Add("system_id", "foobar")
	p.Add("token", "foobar-token")
	p.Add("endpoint", testServer.Listener.Addr().String())

	req := httptest.NewRequest("POST", "/", strings.NewReader(p.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("unexpected status code: %d", rr.Code)
	}

	if rr.Header().Get("Location") != "/opas/styra-engineer-1" {
		t.Fatalf("unexpected location header: %s", rr.Header().Get("Location"))
	}
}

func TestListOPAs(t *testing.T) {
	var err error

	modulePath := "policy/allow.rego"
	example1Mod := `package policy
default allow := true`

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

	ctx := context.Background()

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

	err = m.Add(
		ctx,
		"example2",
		"example2",
		"example2-token",
		testServer.Listener.Addr().String(),
	)
	if err != nil {
		t.Fatalf("unexpected error adding OPA: %s", err)
	}

	h, err := NewOPACollectionHandler(&handlers.Options{
		OPAManager: m,
	})
	if err != nil {
		t.Fatalf("unexpected error creating OPA list handler: %s", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
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

	if !strings.Contains(bodyString, "example1") {
		t.Fatalf("expected example1 to be present")
	}

	if !strings.Contains(bodyString, "example2") {
		t.Fatalf("expected example2 to be present")
	}
}

func TestShowOPA(t *testing.T) {
	var err error

	modulePath := "policy/allow.rego"
	example1Mod := `package policy
default allow := true`

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

	h, err := NewOPAShowHandler(&handlers.Options{
		OPAManager: m,
	})
	if err != nil {
		t.Fatalf("unexpected error creating OPA show handler: %s", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/opas/example1", nil)
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

	if !strings.Contains(bodyString, "example1") {
		t.Fatalf("expected example1 to be present")
	}

	if !strings.Contains(bodyString, "Delete") {
		t.Fatalf("expected delete button to be present")
	}
}
