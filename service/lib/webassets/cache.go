package webassets

import (
	"github.com/gin-gonic/gin"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func accepts(header, encoding string) bool {
	for _, value := range strings.Split(header, ",") {
		parts := strings.Split(strings.TrimSpace(value), ";")
		if parts[0] != encoding {
			continue
		}
		quality := 1.0
		for _, parameter := range parts[1:] {
			if strings.HasPrefix(strings.TrimSpace(parameter), "q=") {
				var err error
				quality, err = strconv.ParseFloat(strings.TrimPrefix(strings.TrimSpace(parameter), "q="), 64)
				if err != nil {
					quality = 0
				}
			}
		}
		return quality > 0 && quality <= 1
	}
	return false
}

// Compressed files are generated at build time; each request only serves bytes.
func Cache(webRoot string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Header("Cache-Control", "no-store")
		}
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
			c.Next()
			return
		}
		if c.Request.URL.Path == "/" {
			c.Header("Cache-Control", "no-cache")
		}
		if !strings.HasPrefix(c.Request.URL.Path, "/assets/") {
			c.Next()
			return
		}
		relative := strings.TrimPrefix(c.Request.URL.Path, "/assets/")
		if relative == "" || strings.Contains(relative, "..") || strings.Contains(relative, "\\") {
			c.AbortWithStatus(404)
			return
		}
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
		c.Header("Vary", "Accept-Encoding")
		for _, encoding := range []struct{ name, suffix string }{{"br", ".br"}, {"gzip", ".gz"}} {
			if !accepts(c.GetHeader("Accept-Encoding"), encoding.name) {
				continue
			}
			file := filepath.Join(webRoot, "assets", relative) + encoding.suffix
			if info, err := os.Stat(file); err != nil || !info.Mode().IsRegular() {
				continue
			}
			c.Header("Content-Encoding", encoding.name)
			c.Header("Content-Type", mime.TypeByExtension(filepath.Ext(relative)))
			c.File(file)
			c.Abort()
			return
		}
		c.Next()
	}
}
