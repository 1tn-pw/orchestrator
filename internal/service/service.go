package service

import (
	"encoding/json"
	"fmt"
	"github.com/1tn-pw/orchestrator/internal/config"
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/keloran/go-healthcheck"
	"github.com/keloran/go-probe"
	"github.com/rs/cors"
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
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{url}", s.GetShort)
	mux.HandleFunc("POST /create", s.CreateShort)
	mux.HandleFunc("GET /health", healthcheck.HTTP)
	mux.HandleFunc("GET /probe", probe.HTTP)

	allowedOrigins := []string{"https://www.1tn.pw", "https://1tn.pw"}
	if s.Config.Local.Development {
		allowedOrigins = append(allowedOrigins, "http://localhost:3000")
	}
	c := cors.New(cors.Options{
		AllowedMethods: []string{http.MethodGet, http.MethodPost},
		AllowedOrigins: allowedOrigins,
		AllowedHeaders: []string{"Accept", "Content-Type"},
		Debug:          true,
	})

	logs.Local().Infof("Starting HTTP on %d", s.Config.Local.HTTPPort)
	errChan <- http.ListenAndServe(fmt.Sprintf(":%d", s.Config.Local.HTTPPort), c.Handler(mux))
}

func (s *Service) CreateShort(w http.ResponseWriter, r *http.Request) {
	type CreateRequest struct {
		URL string `json:"url"`
	}
	u := &CreateRequest{}
	if err := json.NewDecoder(r.Body).Decode(u); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if u.URL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	type ShortResponse struct {
		Short  string `json:"short,omitempty"`
		URL    string `json:"url,omitempty"`
		Status string `json:"status,omitempty"`
	}

	resp, err := NewShortService(r.Context(), s.Config).CreateShort(r.Context(), u.URL)
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
