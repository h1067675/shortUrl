package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Log глобальная переменная отвечающая за логгер.
var Log *zap.Logger

type (
	// responseData хранит сведений об ответе.
	responseData struct {
		status int
		size   int
	}

	// loggingResponseWriter реализуем http.ResponseWriter.
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

// Write записывает размер ответа.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader  записывает код ответа.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// Initialize инициализирует логгер zap определяя настройки.
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cnf := zap.NewProductionConfig()
	cnf.Level = lvl
	cnf.OutputPaths = ([]string{"stdout"})
	cnf.Encoding = "console"
	zapLog, err := cnf.Build()
	if err != nil {
		return err
	}
	Log = zapLog
	return nil
}

// RequestLogger промежуточный хандлер логгирующий информацию о запросе
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		nw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		next.ServeHTTP(&nw, r)

		Log.Debug("Request headers:", zap.Any("values", r.Header))

		Log.Debug("User request:",
			zap.String("URL", r.RequestURI),
			zap.String("method", r.Method),
			zap.Duration("execution time", time.Since(start)),
			zap.Int("size", nw.responseData.size),
			zap.Int("status", nw.responseData.status))

	})
}
