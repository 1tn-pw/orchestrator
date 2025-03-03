package service

import (
	"crypto/tls"
	"fmt"
	"github.com/1tn-pw/orchestrator/internal/config"
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/bugfixes/go-bugfixes/middleware"
	"github.com/keloran/go-healthcheck"
	"github.com/keloran/go-probe"
	"golang.org/x/net/context"
	"net/http"
	"time"
)

type Service struct {
	Config *config.Config
}

func New(cfg *config.Config) *Service {
	return &Service{
		Config: cfg,
	}
}

func (s *Service) Start() error {
	errChan := make(chan error)
	go s.startHTTP(errChan)

	return <-errChan
}

func (s *Service) startHTTP(errChan chan error) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{url}", s.GetShort)
	mux.HandleFunc("POST /create", s.CreateShort)
	mux.HandleFunc("GET /health", healthcheck.HTTP)
	mux.HandleFunc("GET /probe", probe.HTTP)

	mw := middleware.NewMiddleware(context.Background())
	mw.AddMiddleware(middleware.SetupLogger(middleware.Error).Logger)
	mw.AddMiddleware(middleware.RequestID)
	mw.AddMiddleware(middleware.Recoverer)
	mw.AddMiddleware(mw.CORS)
	mw.AddAllowedOrigins("https://www.1tn.pw", "https://1tn.pw")
	if s.Config.Local.Development {
		mw.AddAllowedOrigins("http://localhost:3000", "*")
	}

	logs.Infof("Starting HTTP on %d", s.Config.Local.HTTPPort)
	server := http.Server{
		Addr:              fmt.Sprintf(":%d", s.Config.Local.HTTPPort),
		Handler:           mw.Handler(mux),
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		TLSNextProto:      make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	errChan <- server.ListenAndServe()
}
