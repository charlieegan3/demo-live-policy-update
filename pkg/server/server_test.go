package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"

	"github.com/charlieegan3/demo-live-policy-update/pkg/server/config"
	"github.com/charlieegan3/demo-live-policy-update/pkg/utils"
)

func TestNewServer(t *testing.T) {

	modulePath := "policy/allow.rego"

	modv1 := `
package policy

import rego.v1

allow if input.name in {"alice", "bob", "charlie"}
`

	b := &bundle.Bundle{
		Manifest: bundle.Manifest{
			Revision: "1",
		},
		Modules: []bundle.ModuleFile{
			{
				URL:    modulePath,
				Path:   modulePath,
				Parsed: ast.MustParseModule(modv1),
				Raw:    []byte(modv1),
			},
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/vnd.openpolicyagent.bundles")
		w.Header().Set("etag", b.Manifest.Revision)
		err := bundle.NewWriter(w).Write(*b)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	testServer := httptest.NewServer(http.HandlerFunc(handler))
	defer testServer.Close()

	port, err := utils.FreePort()
	if err != nil {
		t.Fatalf("unexpected error finding free port: %s", err)
	}

	serverConfig := &config.Config{
		Port:    port,
		Address: "localhost",
	}

	svr, err := NewServer(serverConfig)
	if err != nil {
		t.Fatalf("unexpected error creating server: %s", err)
	}

	ctx := context.Background()

	err = svr.Start(ctx)
	if err != nil {
		t.Fatalf("unexpected error starting server: %s", err)
	}

	retries := 10
	for {
		conn, err := http.Get(fmt.Sprintf("http://%s:%d", serverConfig.Address, serverConfig.Port))
		if err == nil {
			conn.Body.Close()
			break
		}

		retries--
		if retries == 0 {
			t.Fatalf("unexpected error connecting to server after retries")
		}

		time.Sleep(100 * time.Millisecond)
	}

	err = svr.Stop(ctx)
	if err != nil {
		t.Fatalf("unexpected error stopping server: %s", err)
	}

}
