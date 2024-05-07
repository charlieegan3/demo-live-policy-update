package opa

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/open-policy-agent/opa/sdk"
)

type Manager struct {
	opas     map[string]*sdk.OPA
	opasLock sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		opas: make(map[string]*sdk.OPA),
	}
}

func (m *Manager) Add(
	ctx context.Context,
	ref string,
	systemID string,
	token string,
	endpoint string,
) error {
	m.opasLock.Lock()
	defer m.opasLock.Unlock()

	if endpoint == "" {
		return fmt.Errorf("endpoint must be provided")
	}

	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "http://" + endpoint
	}

	cfg := fmt.Sprintf(`
bundles:
  systems/%s:
    polling:
      max_delay_seconds: 1
      min_delay_seconds: 1
    resource: /bundles/systems/%s
    service: styra
services:
- credentials:
    bearer:
      token: %s
  name: styra
  url: %s
`, systemID, systemID, token, endpoint)

	opa, err := sdk.New(ctx, sdk.Options{
		Config: strings.NewReader(cfg),
	})
	if err != nil {
		return fmt.Errorf("unexpected error creating OPA instance: %w", err)
	}

	m.opas[ref] = opa

	return nil
}

func (m *Manager) Get(ref string) *sdk.OPA {
	m.opasLock.RLock()
	defer m.opasLock.RUnlock()

	return m.opas[ref]
}

func (m *Manager) Delete(ctx context.Context, ref string) {
	m.opasLock.Lock()
	defer m.opasLock.Unlock()

	s, ok := m.opas[ref]
	if !ok {
		return
	}

	s.Stop(ctx)

	delete(m.opas, ref)
}

func (m *Manager) List() []string {
	m.opasLock.RLock()
	defer m.opasLock.RUnlock()

	refs := make([]string, 0, len(m.opas))
	for ref := range m.opas {
		refs = append(refs, ref)
	}

	return refs
}
