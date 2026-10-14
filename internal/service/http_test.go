package service

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	pb "github.com/1tn-pw/protobufs/generated/short_service/v1"
)

func str(s string) *string { return &s }

// A create the short service failed is a 502 with why, never an empty 200 (Tracklines got "{}" when
// a page behind a sign-in couldn't be read); a made one is the short link.
func TestWriteCreated(t *testing.T) {
	for name, c := range map[string]struct {
		resp *pb.CreateURLResponse
		err  error
		code int
		has  string
	}{
		"made":          {&pb.CreateURLResponse{ShortUrl: "yJjWqM", Url: "https://x"}, nil, 200, `"short":"yJjWqM"`},
		"service error": {&pb.CreateURLResponse{Error: str("fetchTitle: 403 Forbidden")}, nil, 502, `"status":"fetchTitle: 403 Forbidden"`},
		"no short":      {&pb.CreateURLResponse{}, nil, 502, `"status":"no short link made"`},
		"call failed":   {nil, errors.New("connection refused"), 502, `"status":"connection refused"`},
		"no response":   {nil, nil, 502, `"status"`},
	} {
		rec := httptest.NewRecorder()
		writeCreated(rec, c.resp, c.err)
		if rec.Code != c.code || !strings.Contains(rec.Body.String(), c.has) {
			t.Errorf("%s: %d %s", name, rec.Code, rec.Body)
		}
	}
}

// A lookup: the long link, 404 when there's no such short link, 502 when the service can't say.
func TestWriteLong(t *testing.T) {
	for name, c := range map[string]struct {
		resp *pb.GetURLResponse
		err  error
		code int
		has  string
	}{
		"found":       {&pb.GetURLResponse{Url: "https://tracklin.es/tickets/x", Title: "Tracklines"}, nil, 200, `"long":"https://tracklin.es/tickets/x"`},
		"not found":   {&pb.GetURLResponse{Error: str("mongo: no documents in result")}, nil, 404, `"status"`},
		"empty":       {&pb.GetURLResponse{}, nil, 404, `"status":"not found"`},
		"call failed": {nil, errors.New("down"), 502, `"status"`},
	} {
		rec := httptest.NewRecorder()
		writeLong(rec, c.resp, c.err)
		if rec.Code != c.code || !strings.Contains(rec.Body.String(), c.has) {
			t.Errorf("%s: %d %s", name, rec.Code, rec.Body)
		}
	}
}
