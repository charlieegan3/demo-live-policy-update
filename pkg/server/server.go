package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/charlieegan3/demo-live-policy-update/pkg/opa"
	"github.com/charlieegan3/demo-live-policy-update/pkg/server/config"
	"github.com/charlieegan3/demo-live-policy-update/pkg/server/handlers"
	"github.com/charlieegan3/demo-live-policy-update/pkg/server/mux"
)

type Server struct {
	cfg        *config.Config
	httpServer *http.Server
}

func NewServer(cfg *config.Config) (*Server, error) {
	return &Server{
		cfg: cfg,
	}, nil
}

func (s *Server) Start(ctx context.Context) error {
	var err error

	mgr := opa.NewManager()
	if len(s.cfg.OPAs) > 0 {
		for ref, o := range s.cfg.OPAs {
			err = mgr.Add(ctx, ref, o.SystemID, o.Token, o.Endpoint)
			if err != nil {
				return fmt.Errorf("failed to add opa %s: %s", ref, err)
			}
		}
	}

	opts := &handlers.Options{
		OPAManager: mgr,
	}

	m, err := mux.NewMux(opts)
	if err != nil {
		return fmt.Errorf("failed to create mux: %s", err)
	}

	s.httpServer = &http.Server{
		Addr: fmt.Sprintf(
			"%s:%d",
			s.cfg.Address,
			s.cfg.Port,
		),
		Handler: m,
	}

	go func() {
		err = s.httpServer.ListenAndServe()
		if err != nil {
			log.Println(err)
		}
	}()

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err = s.Stop(shutdownCtx)
		if err != nil {
			log.Println(err)
		}
	}()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer != nil {
		err := s.httpServer.Shutdown(ctx)
		if err != nil {
			return err
		}
	}

	s.httpServer = nil

	return nil
}
