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

func (s *Service) CORS(w http.ResponseWriter, r *http.Request) {
	originalOrigin := r.Header.Get("Origin")

	allowedOrigins := []string{"https://www.1tn.pw", "https://1tn.pw"}
	if s.Config.Local.Development {
		allowedOrigins = append(allowedOrigins, "http://localhost:3000")
	}

	isAllowed := false
	for _, origin := range allowedOrigins {
		if origin == originalOrigin {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", originalOrigin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type")
	w.Header().Set("Access-Control-Max-Age", "86400")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.CORS(w, r)

		if r.Method != http.MethodOptions {
			next.ServeHTTP(w, r)
		}
	})
}

func (s *Service) startHTTP(errChan chan error) {
	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("%s /{url}", http.MethodGet), s.GetShort)
	mux.HandleFunc(fmt.Sprintf("%s /create", http.MethodPost), s.CreateShort)
	mux.HandleFunc(fmt.Sprintf("%s /health", http.MethodGet), healthcheck.HTTP)
	mux.HandleFunc(fmt.Sprintf("%s /probe", http.MethodGet), probe.HTTP)

	middleWare := s.Middleware(mux)

	logs.Local().Infof("Starting HTTP on %d", s.Config.Local.HTTPPort)
	errChan <- http.ListenAndServe(fmt.Sprintf(":%d", s.Config.Local.HTTPPort), middleWare)
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
