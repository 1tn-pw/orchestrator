package service

import (
	"encoding/json"
	"github.com/bugfixes/go-bugfixes/logs"
	"net/http"
)

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

	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
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
		Long        string `json:"long,omitempty"`
		Title       string `json:"title,omitempty"`
		Favicon     string `json:"favicon,omitempty"`
		Description string `json:"description,omitempty"`

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

	if resp == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(&LongResponse{
		Long:        resp.GetUrl(),
		Title:       resp.GetTitle(),
		Favicon:     resp.GetFavicon(),
		Description: resp.GetDescription(),
	}); err != nil {
		_ = logs.Errorf("Error writing response: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
