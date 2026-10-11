package webassets

import (
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"strings"
)

type jsonWriter struct {
	gin.ResponseWriter
	compressor *gzip.Writer
	decided    bool
}

func (w *jsonWriter) Write(data []byte) (int, error) {
	if !w.decided {
		w.decided = true
		if len(data) > 1024 && strings.HasPrefix(w.Header().Get("Content-Type"), "application/json") && w.Header().Get("Content-Encoding") == "" {
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Add("Vary", "Accept-Encoding")
			w.Header().Del("Content-Length")
			w.compressor = gzip.NewWriter(w.ResponseWriter)
		}
	}
	if w.compressor != nil {
		return w.compressor.Write(data)
	}
	return w.ResponseWriter.Write(data)
}
func (w *jsonWriter) WriteString(s string) (int, error) { return w.Write([]byte(s)) }
func (w *jsonWriter) Flush() {
	if w.compressor != nil {
		w.compressor.Flush()
	}
	w.ResponseWriter.Flush()
}

func JSONCompression() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api/") || !accepts(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}
		writer := &jsonWriter{ResponseWriter: c.Writer}
		c.Writer = writer
		c.Next()
		if writer.compressor != nil {
			writer.compressor.Close()
		}
	}
}
