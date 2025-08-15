package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// compressWriter описывает структуру необходимую для сжатия данных.
type compressWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

// Write реализует метод интерфейса
func (w compressWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// CompressHandle промежуточный хэндлер отвечающий за сжатие данных
func CompressHandle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			if strings.Contains(r.Header.Get("Content-type"), "application/json") || strings.Contains(r.Header.Get("Content-type"), "text/html") {
				gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
				if err != nil {
					io.WriteString(w, err.Error())
					return
				}
				defer gz.Close()
				w.Header().Set("Content-Encoding", "gzip")
				w = compressWriter{ResponseWriter: w, Writer: gz}
			}
		}
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") && strings.Contains(r.Header.Get("Content-type"), "application/x-gzip") {
			cr, err := gzip.NewReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}
		next.ServeHTTP(w, r)
	})
}
