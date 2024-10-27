package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var Log *zap.Logger

type (
	// берём структуру для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// добавляем реализацию http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

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

		Log.Debug("User request:",
			zap.String("URL", r.RequestURI),
			zap.String("method", r.Method),
			zap.Duration("execution time", time.Since(start)),
			zap.Int("size", nw.responseData.size),
			zap.Int("status", nw.responseData.status))

	})
}
