package storage

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/h1067675/shortUrl/internal/logger"
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
func (s *Storage) checkLinksDBTable() bool {
	rows, err := s.DB.Query("SELECT * FROM links LIMIT 1;")
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
func (s *Storage) checkUsersDBTable() bool {
	rows, err := s.DB.Query("SELECT * FROM users LIMIT 1;")
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

// Функция создает таблицу ссылок в базе данных
func (s *Storage) createLinksDBTable() bool {
	_, err := s.DB.Exec(`CREATE TABLE links (
		Id SERIAL PRIMARY KEY, 
		InnerLink TEXT, 
		OutterLink TEXT UNIQUE
		);`)
	if err != nil {
		logger.Log.Debug("data base don't exist.", zap.Error(err))
	}
	return err == nil
}

// Функция создает таблицы пользователей в базе данных
func (s *Storage) createUsersDBTables() bool {
	var err error
	_, err = s.DB.Exec(`CREATE TABLE users (
		Id SERIAL PRIMARY KEY,
		creation TIMESTAMP
		);`)
	if err != nil {
		logger.Log.Debug("data base don't exist.", zap.Error(err))
	}
	_, err = s.DB.Exec(`CREATE TABLE users_links (
		Id INTEGER,
		LinkID INTEGER
		);`)
	if err != nil {
		logger.Log.Debug("data base don't exist.", zap.Error(err))
	}
	return err == nil
}

// Функция создания короткого URL и сохранения его в базу данных
func (s *Storage) saveShortURLBD(url string, adr string, userid int) (result string, errexit error) {
	var linkid int
	var err error
	result = s.createShortCode(adr)
	_, err = s.DB.Exec("INSERT INTO links (InnerLink, OutterLink) VALUES ($1, $2);", result, url)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			errexit = ErrLinkExsist
		} else {
			return "", err
		}
	}
	row := s.DB.QueryRow("SELECT Id, InnerLink FROM links WHERE OutterLink = $1;", url)
	err = row.Scan(&linkid, &result)
	if err != nil {
		return "", err
	}
	_, err = s.DB.Exec("INSERT INTO users_links (Id, LinkId) VALUES ($1, $2);", userid, linkid)
	if err != nil {
		return "", err
	}
	return result, errexit
}

// Функция получения внешнего URL по короткому URL из базы данных
func (s *Storage) getURLBD(url string) string {
	row := s.DB.QueryRow("SELECT OutterLink FROM links WHERE InnerLink = $1;", url)
	var result string
	err := row.Scan(&result)
	if err != nil {
		return ""
	}
	return result
}

// Функция получения внешнего URL по короткому URL из базы данных
func (s *Storage) getUserURLBD(id int) (result []struct {
	ShortURL string
	URL      string
}, err error) {
	rows, err := s.DB.Query("SELECT InnerLink, OutterLink FROM links WHERE Id IN (SELECT LinkId FROM users_links WHERE Id = $1);", id)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var in string
		var out string
		err = rows.Scan(&in, &out)
		result = append(result, struct {
			ShortURL string
			URL      string
		}{
			ShortURL: in,
			URL:      out,
		})
	}
	return result, err
}

func (s *Storage) getNewUserIDDB() (result int, err error) {
	_, err = s.DB.Exec("INSERT INTO users (creation) VALUES (current_timestamp);")
	if err != nil {
		return -1, err
	}
	row := s.DB.QueryRow("SELECT MAX(Id) FROM users;")
	err = row.Scan(&result)
	if err != nil {
		return -1, err
	}
	return result, nil
}
