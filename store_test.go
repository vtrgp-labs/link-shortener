package main

import (
	"errors"
	"path/filepath"
	"regexp"
	"sync"
	"testing"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewStore(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestCreateAndGet(t *testing.T) {
	s := newTestStore(t)

	code, err := s.Create("https://example.com")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9]{6}$`).MatchString(code) {
		t.Errorf("code %q is not 6 alphanumeric chars", code)
	}

	url, err := s.Get(code)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if url != "https://example.com" {
		t.Errorf("Get = %q, want https://example.com", url)
	}
}

func TestGetUnknown(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.Get("nope00"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get unknown: err = %v, want ErrNotFound", err)
	}
}

func TestCreateUniqueCodes(t *testing.T) {
	s := newTestStore(t)
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		code, err := s.Create("https://example.com")
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		if seen[code] {
			t.Fatalf("duplicate code %q", code)
		}
		seen[code] = true
	}
}

func TestVisitIncrementsClicks(t *testing.T) {
	s := newTestStore(t)
	code, _ := s.Create("https://example.com")

	for i := 0; i < 3; i++ {
		url, err := s.Visit(code)
		if err != nil {
			t.Fatalf("Visit: %v", err)
		}
		if url != "https://example.com" {
			t.Errorf("Visit = %q, want https://example.com", url)
		}
	}

	links, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(links) != 1 || links[0].Clicks != 3 {
		t.Errorf("List = %+v, want one link with 3 clicks", links)
	}
}

func TestVisitUnknown(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.Visit("nope00"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Visit unknown: err = %v, want ErrNotFound", err)
	}
}

func TestVisitConcurrent(t *testing.T) {
	s := newTestStore(t)
	code, _ := s.Create("https://example.com")

	const n = 50
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := s.Visit(code); err != nil {
				t.Errorf("Visit: %v", err)
			}
		}()
	}
	wg.Wait()

	links, _ := s.List()
	if links[0].Clicks != n {
		t.Errorf("clicks = %d, want %d", links[0].Clicks, n)
	}
}

func TestListNewestFirst(t *testing.T) {
	s := newTestStore(t)
	first, _ := s.Create("https://a.example")
	second, _ := s.Create("https://b.example")

	links, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(links) != 2 || links[0].Code != second || links[1].Code != first {
		t.Errorf("List = %+v, want [%s %s]", links, second, first)
	}
}
