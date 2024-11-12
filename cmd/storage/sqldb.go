package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/h1067675/shortUrl/internal/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type SQLDB struct {
	*sql.DB
	Connected bool
}

func newDB(DatabasePath string) *SQLDB {
	r, err := sql.Open("pgx", DatabasePath)
	if err != nil {
		return &SQLDB{Connected: false}
	}
	return &SQLDB{DB: r, Connected: true}
}

// Функция проверяет соединение с базой данных
func (s *Storage) PingDB() bool {
	if s.DB.Connected {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.DB.DB.PingContext(ctx); err == nil {
			return true
		}
	}
	return false
}

// Функция проверяет наличие таблицы в базе данных
func (s *Storage) checkDBTable() bool {
	rows, err := s.DB.Query("SELECT * FROM links LIMIT 1")
	if err != nil {
		logger.Log.Debug("data base don't exist.", zap.Error(err))
		return false
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err() // or modify return value
	}()
	return true
}

// Функция проверяет наличие таблицы в базе данных
func (s *Storage) createDBTable() bool {
	_, err := s.DB.Exec(`CREATE TABLE links (
		Id SERIAL PRIMARY KEY, 
		InnerLink CHARACTER VARYING(256), 
		OutterLink CHARACTER VARYING(256))`)
	if err != nil {
		logger.Log.Debug("data base don't exist.", zap.Error(err))
	}
	return err == nil
}

// Функция создания короткого URL и сохранения его в базу данных
func (s *Storage) saveShortURLBD(url string, adr string) string {
	row := s.DB.QueryRow("SELECT InnerLink FROM links WHERE OutterLink = $1", url)
	var result string
	err := row.Scan(&result)
	if err != nil {
		result = s.createShortCode(adr)
		_, err := s.DB.Exec("INSERT INTO links (InnerLink, OutterLink) VALUES ($1, $2)", result, url)
		if err != nil {
			return ""
		}
	}
	return result
}

// Функция получения внешнего URL по короткому URL из базы данных
func (s *Storage) getURLBD(url string) string {
	row := s.DB.QueryRow("SELECT OutterLink FROM links WHERE InnerLink = '$1'", url)
	var result string
	err := row.Scan(&result)
	if err != nil {
		return ""
	}
	return result
}
