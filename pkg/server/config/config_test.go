package config

import "testing"

func TestParseConfig(t *testing.T) {

	rawConfig := []byte(`
address: "localhost"
port: 8080

opas:
  alice:
    endpoint: "http://localhost:8181"
    token: "alice-token"
    system_id: "alice-system"
  bob:
    endpoint: "http://localhost:8182"
    token: "bob-token"
    system_id: "bob-system"
`)

	cfg, err := ParseConfig(rawConfig)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if cfg.Address != "localhost" {
		t.Fatalf("unexpected address: %s", cfg.Address)
	}

	if cfg.Port != 8080 {
		t.Fatalf("unexpected port: %d", cfg.Port)
	}

	if len(cfg.OPAs) != 2 {
		t.Fatalf("unexpected number of opas: %d", len(cfg.OPAs))
	}

	if cfg.OPAs["alice"].Endpoint != "http://localhost:8181" {
		t.Fatalf("unexpected alice endpoint: %s", cfg.OPAs["alice"].Endpoint)
	}

	if cfg.OPAs["alice"].Token != "alice-token" {
		t.Fatalf("unexpected alice token: %s", cfg.OPAs["alice"].Token)
	}

	if cfg.OPAs["alice"].SystemID != "alice-system" {
		t.Fatalf("unexpected alice system_id: %s", cfg.OPAs["alice"].SystemID)
	}

	if cfg.OPAs["bob"].Endpoint != "http://localhost:8182" {
		t.Fatalf("unexpected bob endpoint: %s", cfg.OPAs["bob"].Endpoint)
	}

	if cfg.OPAs["bob"].Token != "bob-token" {
		t.Fatalf("unexpected bob token: %s", cfg.OPAs["bob"].Token)
	}

	if cfg.OPAs["bob"].SystemID != "bob-system" {
		t.Fatalf("unexpected bob system_id: %s", cfg.OPAs["bob"].SystemID)
	}
}
