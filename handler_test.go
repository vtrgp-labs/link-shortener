package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func newTestServer(t *testing.T) (http.Handler, *Store) {
	t.Helper()
	s := newTestStore(t)
	return NewServer(s), s
}

func do(h http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func postShorten(h http.Handler, target string) *httptest.ResponseRecorder {
	form := url.Values{"url": {target}}
	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return do(h, req)
}

func TestIndex(t *testing.T) {
	h, _ := newTestServer(t)
	rec := do(h, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `<form method="post" action="/shorten">`) {
		t.Error("index does not contain the form")
	}
}

func TestShortenValid(t *testing.T) {
	h, s := newTestServer(t)
	rec := postShorten(h, "https://example.com/long")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	links, _ := s.List()
	if len(links) != 1 {
		t.Fatalf("stored %d links, want 1", len(links))
	}
	want := "http://example.com/" + links[0].Code
	if !strings.Contains(rec.Body.String(), want) {
		t.Errorf("response does not contain short URL %q", want)
	}
}

func TestShortenInvalid(t *testing.T) {
	h, s := newTestServer(t)
	for _, in := range []string{"", "not a url", "ftp://example.com", "http://"} {
		rec := postShorten(h, in)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("shorten %q: status = %d, want 400", in, rec.Code)
		}
	}
	if links, _ := s.List(); len(links) != 0 {
		t.Errorf("stored %d links for invalid input, want 0", len(links))
	}
}

func TestRedirect(t *testing.T) {
	h, s := newTestServer(t)
	code, _ := s.Create("https://example.com/target")

	rec := do(h, httptest.NewRequest(http.MethodGet, "/"+code, nil))
	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("status = %d, want 301", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "https://example.com/target" {
		t.Errorf("Location = %q, want https://example.com/target", loc)
	}
}

func TestRedirectUnknown(t *testing.T) {
	h, _ := newTestServer(t)
	rec := do(h, httptest.NewRequest(http.MethodGet, "/zzzzzz", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

// Acceptance test from issue #3.
func TestClickCountShownOnIndex(t *testing.T) {
	h, s := newTestServer(t)
	postShorten(h, "https://example.com/counted")
	links, _ := s.List()
	code := links[0].Code

	for i := 0; i < 3; i++ {
		do(h, httptest.NewRequest(http.MethodGet, "/"+code, nil))
	}

	body := do(h, httptest.NewRequest(http.MethodGet, "/", nil)).Body.String()
	if !strings.Contains(body, `<td class="clicks">3</td>`) {
		t.Errorf("index does not show 3 clicks:\n%s", body)
	}
}
