package storage

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
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
