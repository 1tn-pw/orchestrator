package service

import (
	"cmp"
	"encoding/json"
	"errors"
	"net/http"

	pb "github.com/1tn-pw/protobufs/generated/short_service/v1"
	"github.com/bugfixes/go-bugfixes/logs"
)

func (s *Service) CreateShort(w http.ResponseWriter, r *http.Request) {
	// Title and Description are optional; given a title, the short service uses them instead of reading the page.
	type CreateRequest struct {
		URL         string `json:"url"`
		Title       string `json:"title"`
		Description string `json:"description"`
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

	resp, err := NewShortService(r.Context(), s.Config).CreateShort(r.Context(), &pb.CreateURLRequest{
		Url: u.URL, Title: &u.Title, Description: &u.Description,
	})
	writeCreated(w, resp, err)
}

// writeCreated answers a create: the short link, or 502 with why. The short service reports most
// failures in the response's Error rather than as a call error; both are failures, never an empty 200.
func writeCreated(w http.ResponseWriter, resp *pb.CreateURLResponse, err error) {
	type ShortResponse struct {
		Short  string `json:"short,omitempty"`
		URL    string `json:"url,omitempty"`
		Status string `json:"status,omitempty"`
	}
	w.Header().Set("Content-Type", "application/json")
	if err == nil && resp == nil {
		err = errors.New("no response from the short service")
	}
	if err == nil && (resp.GetError() != "" || resp.GetShortUrl() == "") {
		err = errors.New(cmp.Or(resp.GetError(), "no short link made"))
	}
	if err != nil {
		_ = logs.Errorf("CreateShort: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(&ShortResponse{Status: err.Error()})
		return
	}
	if err := json.NewEncoder(w).Encode(&ShortResponse{URL: resp.GetUrl(), Short: resp.GetShortUrl()}); err != nil {
		_ = logs.Errorf("Error writing response: %s", err.Error())
	}
}

func (s *Service) GetShort(w http.ResponseWriter, r *http.Request) {
	url := r.PathValue("url")
	if url == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, err := NewShortService(r.Context(), s.Config).GetLong(r.Context(), url)
	writeLong(w, resp, err)
}

// writeLong answers a lookup: the long link and its details, 404 when there's no such short link, or
// 502 when the short service couldn't say.
func writeLong(w http.ResponseWriter, resp *pb.GetURLResponse, err error) {
	type LongResponse struct {
		Long        string `json:"long,omitempty"`
		Title       string `json:"title,omitempty"`
		Favicon     string `json:"favicon,omitempty"`
		Description string `json:"description,omitempty"`
		Status      string `json:"status,omitempty"`
	}
	w.Header().Set("Content-Type", "application/json")
	if err != nil || resp == nil {
		_ = logs.Errorf("GetShort: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(&LongResponse{Status: "short service unavailable"})
		return
	}
	if resp.GetError() != "" || resp.GetUrl() == "" {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(&LongResponse{Status: cmp.Or(resp.GetError(), "not found")})
		return
	}
	if err := json.NewEncoder(w).Encode(&LongResponse{
		Long: resp.GetUrl(), Title: resp.GetTitle(), Favicon: resp.GetFavicon(), Description: resp.GetDescription(),
	}); err != nil {
		_ = logs.Errorf("Error writing response: %s", err.Error())
	}
}
