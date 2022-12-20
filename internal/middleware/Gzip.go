package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/c0dered273/go-adv-metrics/internal/log"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (gw gzipWriter) Write(b []byte) (int, error) {
	return gw.Writer.Write(b)
}

func GzipResponseEncoder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.DefaultCompression)
		if err != nil {
			log.Error.Println("can`t create gzip encoder ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cw := &gzipWriter{
			ResponseWriter: w,
			Writer:         gz,
		}
		defer gz.Close()

		cw.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(cw, r)
	})
}

func GzipRequestDecoder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			log.Error.Println("can`t create gzip reader ", err)
			next.ServeHTTP(w, r)
			return
		}
		defer gz.Close()

		r.Body = io.NopCloser(gz)

		next.ServeHTTP(w, r)
	})
}
