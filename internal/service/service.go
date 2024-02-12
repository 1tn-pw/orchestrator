package service

import (
	"encoding/json"
	"fmt"
	"github.com/1tn-pw/orchestrator/internal/config"
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/keloran/go-healthcheck"
	"github.com/keloran/go-probe"
	"net/http"
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
	http.HandleFunc("GET /{url}", s.GetShort)
	http.HandleFunc("POST /create", s.CreateShort)
	http.HandleFunc("GET /health", healthcheck.HTTP)
	http.HandleFunc("GET /probe", probe.HTTP)

	logs.Local().Infof("Starting HTTP on %d", s.Config.Local.HTTPPort)
	errChan <- http.ListenAndServe(fmt.Sprintf(":%d", s.Config.Local.HTTPPort), nil)
}

func (s *Service) enableCORS(w http.ResponseWriter) {
	if s.Config.Local.Development {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	} else {
		w.Header().Set("Access-Control-Allow-Origin", "https://www.1tn.pw")
		w.Header().Set("Access-Control-Allow-Origin", "https://1tn.pw")
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
}

func (s *Service) CreateShort(w http.ResponseWriter, r *http.Request) {
	s.enableCORS(w)

	url := r.FormValue("url")
	if url == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	type ShortResponse struct {
		Short  string `json:"short,omitempty"`
		URL    string `json:"url,omitempty"`
		Status string `json:"status,omitempty"`
	}

	resp, err := NewShortService(r.Context(), s.Config).CreateShort(r.Context(), url)
	if err != nil {
		if err := json.NewEncoder(w).Encode(&ShortResponse{
			Status: err.Error(),
		}); err != nil {
			_ = logs.Errorf("Error writing error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	if err := json.NewEncoder(w).Encode(&ShortResponse{
		URL:   resp.GetUrl(),
		Short: resp.GetShortUrl(),
	}); err != nil {
		_ = logs.Errorf("Error writing response: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Service) GetShort(w http.ResponseWriter, r *http.Request) {
	s.enableCORS(w)

	url := r.PathValue("url")
	if url == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	type LongResponse struct {
		Long   string `json:"long,omitempty"`
		Status string `json:"status,omitempty"`
	}

	resp, err := NewShortService(r.Context(), s.Config).GetLong(r.Context(), url)
	if err != nil {
		if err := json.NewEncoder(w).Encode(&LongResponse{
			Status: err.Error(),
		}); err != nil {
			_ = logs.Errorf("Error writing error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	if err := json.NewEncoder(w).Encode(&LongResponse{
		Long: resp.GetUrl(),
	}); err != nil {
		_ = logs.Errorf("Error writing response: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
