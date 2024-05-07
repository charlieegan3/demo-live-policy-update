package opa

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"
	"github.com/open-policy-agent/opa/sdk"
)

func TestNewManager(t *testing.T) {
	var err error

	modulePath := "policy/allow.rego"

	example1Mod := `
package policy

import rego.v1

default allow := false
allow if input.name in {"alice", "bob", "charlie"}
`

	example1Bundle := &bundle.Bundle{
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

	example2Mod := `
package policy

import rego.v1

default allow := false
allow if input.name in {"diane", "erin", "frank"}
`

	example2Bundle := &bundle.Bundle{
		Manifest: bundle.Manifest{
			Revision: "1",
		},
		Modules: []bundle.ModuleFile{
			{
				URL:    modulePath,
				Path:   modulePath,
				Parsed: ast.MustParseModule(example2Mod),
				Raw:    []byte(example2Mod),
			},
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/vnd.openpolicyagent.bundles")

		if r.URL.Path == "/bundles/systems/example1" {
			w.Header().Set("etag", example2Bundle.Manifest.Revision)
			err = bundle.NewWriter(w).Write(*example1Bundle)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			return
		}

		if r.URL.Path == "/bundles/systems/example2" {
			w.Header().Set("etag", example2Bundle.Manifest.Revision)
			err = bundle.NewWriter(w).Write(*example2Bundle)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			return
		}

		t.Fatalf("unexpected request path: %s", r.URL.Path)
	}

	testServer := httptest.NewServer(http.HandlerFunc(handler))
	defer testServer.Close()

	m := NewManager()

	ctx := context.Background()

	err = m.Add(ctx, "example1", "example1", "example1-token", testServer.Listener.Addr().String())
	if err != nil {
		t.Fatalf("unexpected error adding OPA: %s", err)
	}

	err = m.Add(ctx, "example2", "example2", "example2-token", testServer.Listener.Addr().String())
	if err != nil {
		t.Fatalf("unexpected error adding OPA: %s", err)
	}

	ex1SDK := m.Get("example1")
	if ex1SDK == nil {
		t.Fatalf("expected OPA to be present")
	}

	ex2SDK := m.Get("example2")
	if ex2SDK == nil {
		t.Fatalf("expected OPA to be present")
	}

	// test example1 deny
	dr, err := ex1SDK.Decision(ctx, sdk.DecisionOptions{
		Path: "/policy/allow",
		Input: map[string]interface{}{
			"name": "amy",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error evaluating decision: %s", err)
	}
	if dr.Result == nil {
		t.Fatalf("expected result to be non-nil")
	}
	allowed, ok := dr.Result.(bool)
	if !ok {
		t.Fatalf("expected result to be a bool")
	}
	if allowed {
		t.Fatalf("expected result to be false")
	}

	// test example1 allow
	dr, err = ex1SDK.Decision(ctx, sdk.DecisionOptions{
		Path: "/policy/allow",
		Input: map[string]interface{}{
			"name": "alice",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error evaluating decision: %s", err)
	}
	if dr.Result == nil {
		t.Fatalf("expected result to be non-nil")
	}
	allowed, ok = dr.Result.(bool)
	if !ok {
		t.Fatalf("expected result to be a bool")
	}
	if !allowed {
		t.Fatalf("expected result to be true")
	}

	// test example2 deny
	dr, err = ex2SDK.Decision(ctx, sdk.DecisionOptions{
		Path: "/policy/allow",
		Input: map[string]interface{}{
			"name": "amy",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error evaluating decision: %s", err)
	}
	if dr.Result == nil {
		t.Fatalf("expected result to be non-nil")
	}
	allowed, ok = dr.Result.(bool)
	if !ok {
		t.Fatalf("expected result to be a bool")
	}
	if allowed {
		t.Fatalf("expected result to be false")
	}

	// test example2 allow
	dr, err = ex2SDK.Decision(ctx, sdk.DecisionOptions{
		Path: "/policy/allow",
		Input: map[string]interface{}{
			"name": "diane",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error evaluating decision: %s", err)
	}
	if dr.Result == nil {
		t.Fatalf("expected result to be non-nil")
	}
	allowed, ok = dr.Result.(bool)
	if !ok {
		t.Fatalf("expected result to be a bool")
	}
	if !allowed {
		t.Fatalf("expected result to be true")
	}

	// test that polling is happening at 1s intervals and that the updated policy is eventually loaded
	example1ModUpdated := `
package policy

import rego.v1

default allow := false
allow if input.name in {"eve"}
`

	example1Bundle = &bundle.Bundle{
		Manifest: bundle.Manifest{
			Revision: "1",
		},
		Modules: []bundle.ModuleFile{
			{
				URL:    modulePath,
				Path:   modulePath,
				Parsed: ast.MustParseModule(example1ModUpdated),
				Raw:    []byte(example1ModUpdated),
			},
		},
	}

	retries := 5
	for {
		dr, err = ex1SDK.Decision(ctx, sdk.DecisionOptions{
			Path: "/policy/allow",
			Input: map[string]interface{}{
				"name": "alice",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error evaluating decision: %s", err)
		}
		if dr.Result == nil {
			t.Fatalf("expected result to be non-nil")
		}
		allowed, ok = dr.Result.(bool)
		if !ok {
			t.Fatalf("expected result to be a bool")
		}
		if allowed {
			break
		}
		retries--
		if retries == 0 {
			t.Fatalf("expected eve to allowed by updated policy after retries")
		}
		time.Sleep(500 * time.Millisecond)
	}
}
