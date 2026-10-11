package webassets

import (
	"bytes"
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEncodingQualityAndHeaders(t *testing.T) {
	for _, test := range []struct {
		header   string
		accepted bool
	}{{"gzip, br", true}, {"gzip;q=0", false}, {"gzip;q=0.5", true}, {"gzip;q=bogus", false}} {
		if accepts(test.header, "gzip") != test.accepted {
			t.Fatal(test)
		}
	}
	gin.SetMode(gin.TestMode)
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "assets"), 0700)
	os.WriteFile(filepath.Join(root, "assets", "app.js"), []byte("source"), 0600)
	os.WriteFile(filepath.Join(root, "assets", "app.js.br"), []byte("compressed"), 0600)
	router := gin.New()
	router.Use(Cache(root), JSONCompression())
	router.Static("/assets", filepath.Join(root, "assets"))
	router.GET("/", func(c *gin.Context) { c.String(200, "html") })
	router.POST("/api/cards", func(c *gin.Context) { c.JSON(200, gin.H{"cards": strings.Repeat("large response ", 1000)}) })
	req := httptest.NewRequest("GET", "/assets/app.js", nil)
	req.Header.Set("Accept-Encoding", "br,gzip")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Header().Get("Content-Encoding") != "br" || !strings.Contains(w.Header().Get("Cache-Control"), "immutable") || w.Body.String() != "compressed" {
		t.Fatal(w.Header(), w.Body.String())
	}
	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Header().Get("Cache-Control") != "no-cache" {
		t.Fatal(w.Header())
	}
	req = httptest.NewRequest("POST", "/api/cards", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Header().Get("Content-Encoding") != "gzip" || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal(w.Header())
	}
	reader, err := gzip.NewReader(bytes.NewReader(w.Body.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(reader)
	if err != nil || !bytes.Contains(data, []byte("large response")) {
		t.Fatal(err)
	}
}
