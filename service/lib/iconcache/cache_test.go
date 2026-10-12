package iconcache

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func bitmap(t *testing.T) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 350, 350))
	img.SetNRGBA(175, 175, color.NRGBA{255, 50, 100, 128})
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}
func reply(r *http.Request, data []byte) *http.Response {
	return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(bytes.NewReader(data)), Request: r, ContentLength: int64(len(data))}
}

func TestFetchConcurrentPersistentAndImmutable(t *testing.T) {
	data := bitmap(t)
	var hits atomic.Int32
	client := &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		hits.Add(1)
		time.Sleep(10 * time.Millisecond)
		return reply(r, data), nil
	})}
	root := t.TempDir()
	cache := New(root, client)
	source := "https://icons.example.com/icon.png?v=2"
	var wg sync.WaitGroup
	results := make(chan string, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			local, err := cache.Fetch(context.Background(), source)
			if err != nil {
				t.Error(err)
			}
			results <- local
		}()
	}
	wg.Wait()
	close(results)
	var local string
	for result := range results {
		if local != "" && result != local {
			t.Fatal("duplicate assets")
		}
		local = result
	}
	if hits.Load() != 1 {
		t.Fatal("network requests", hits.Load())
	}
	restart := New(root, client)
	if restart.Resolve(source) != local || hits.Load() != 1 {
		t.Fatal("cache not persistent")
	}
	stored, err := os.ReadFile(filepath.Join(root, "assets", strings.TrimPrefix(local, "/icon-cache/")))
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(stored))
	if err != nil || img.Bounds().Dx() != 192 || img.Bounds().Dy() != 192 {
		t.Fatal("image not resized", err)
	}
	if _, _, _, alpha := img.At(0, 0).RGBA(); alpha != 0 {
		t.Fatal("transparency lost")
	}
	w := httptest.NewRecorder()
	restart.ServeHTTP(w, httptest.NewRequest("GET", local, nil))
	if w.Code != 200 || !strings.Contains(w.Header().Get("Cache-Control"), "immutable") || w.Header().Get("Content-Type") != "image/png" {
		t.Fatal(w.Code, w.Header())
	}
	req := httptest.NewRequest("GET", local, nil)
	req.Header.Set("If-None-Match", w.Header().Get("ETag"))
	w = httptest.NewRecorder()
	restart.ServeHTTP(w, req)
	if w.Code != 304 {
		t.Fatal(w.Code)
	}
	w = httptest.NewRecorder()
	restart.ServeHTTP(w, httptest.NewRequest("GET", "/icon-cache/../index/a.json", nil))
	if w.Code != 404 {
		t.Fatal("metadata exposed")
	}
}

func TestPruneSharedAndPending(t *testing.T) {
	data := bitmap(t)
	cache := New(t.TempDir(), &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) { return reply(r, data), nil })})
	a, b := "https://one.example/icon.png", "https://two.example/icon.png"
	local, err := cache.Fetch(context.Background(), a)
	if err != nil {
		t.Fatal(err)
	}
	other, err := cache.Fetch(context.Background(), b)
	if err != nil || other != local {
		t.Fatal(other, err)
	}
	if err := cache.Prune([]string{b}, nil, []string{a}); err != nil {
		t.Fatal(err)
	}
	if cache.readEntry(key(a)).File != "" || cache.readEntry(key(b)).File == "" {
		t.Fatal("incorrect shared metadata cleanup")
	}
	if err := cache.Prune(nil, nil, []string{b}); err != nil {
		t.Fatal(err)
	}
	files, _ := os.ReadDir(filepath.Join(cache.Root, "assets"))
	if len(files) != 0 {
		t.Fatal("orphan assets", len(files))
	}
	started, release := make(chan struct{}), make(chan struct{})
	cache = New(t.TempDir(), &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) { close(started); <-release; return reply(r, data), nil })})
	cache.Resolve(a)
	<-started
	cache.Prune(nil, nil, []string{a})
	close(release)
	cache.mu.Lock()
	pending := cache.jobs[key(a)]
	cache.mu.Unlock()
	if pending != nil {
		<-pending.done
	}
	if cache.readEntry(key(a)).File != "" {
		t.Fatal("deleted icon returned after pending download")
	}
}

func TestDiscoveryPrefersAppleTouchAndPreservesQuery(t *testing.T) {
	data := bitmap(t)
	var paths []string
	var mu sync.Mutex
	cache := New(t.TempDir(), &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		mu.Lock()
		paths = append(paths, r.URL.RequestURI())
		mu.Unlock()
		if r.URL.Path == "/deep/page" {
			return reply(r, []byte(`<html><title> Example Site </title><meta name="description" content="A useful website"><link rel="icon" sizes="32x32" href="/small.png"><link rel="apple-touch-icon" sizes="180x180" href="../apple.png?v=7"></html>`)), nil
		}
		return reply(r, data), nil
	})})
	info, err := cache.Discover(context.Background(), "https://example.com/deep/page")
	if err != nil {
		t.Fatal(err)
	}
	if info.OriginalIconURL != "https://example.com/apple.png?v=7" || info.Title != "Example Site" || info.Description != "A useful website" || len(paths) != 2 {
		t.Fatal(info, paths)
	}
}

func TestRejectsUnsafeOrInvalidImages(t *testing.T) {
	for _, data := range [][]byte{[]byte(`<html><h1>blocked</h1></html>`), []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`), []byte(`<svg><path onload="x()"/></svg>`), []byte(`<svg><use href="https://evil.example/a"/></svg>`), {0, 0, 1, 0, 1, 0}} {
		if _, _, err := prepare(data); err == nil {
			t.Fatalf("accepted %q", data)
		}
	}
	valid := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><path fill="#f00" d="M0 0h100v100H0z"/></svg>`)
	prepared, ext, err := prepare(valid)
	if err != nil || ext != ".svg" {
		t.Fatal(err)
	}
	if bytes.Count(prepared, []byte(`xmlns="http://www.w3.org/2000/svg"`)) < 1 {
		t.Fatal("namespace lost")
	}
	if _, _, err := prepare(prepared); err != nil {
		t.Fatal("invalid encoded SVG", err)
	}
}

func TestSiteNamePrefersShortNames(t *testing.T) {
	page := func(raw string) *url.URL { u, _ := url.Parse(raw); return u }
	for _, c := range []struct{ title, site, page, want string }{
		{"GitHub · Change is constant. GitHub keeps you ahead. · GitHub", "GitHub", "https://github.com", "GitHub"},
		{"requests · PyPI", "PyPI", "https://pypi.org/project/requests/", "requests"},
		{"PyPI · The Python Package Index", "", "https://pypi.org/", "PyPI"},
		{"Welcome | Example Docs", "", "https://docs.example.com/", "Example Docs"},
		{"哔哩哔哩 (゜-゜)つロ 干杯~-bilibili", "", "https://www.bilibili.com/", "哔哩哔哩 (゜-゜)つロ 干杯~-bil"},
		{"", "", "https://empty.example/", ""},
	} {
		if got := siteName(c.title, c.site, page(c.page)); got != c.want {
			t.Errorf("%q -> %q, want %q", c.title, got, c.want)
		}
	}
}
