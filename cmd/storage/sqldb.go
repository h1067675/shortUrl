package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
		LinkID INTEGER,
		is_deleted BOOL
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
			logger.Log.Debug("link alredy exist")
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
	logger.Log.Debug("link added to user", zap.Int("id", userid), zap.String("originurl", url), zap.String("shorturl", result), zap.Int("shorturl", linkid))
	return result, errexit
}

// Функция получения внешнего URL по короткому URL из базы данных
func (s *Storage) getURLBD(url string, userid int) (res string, err error) {
	row := s.DB.QueryRow("SELECT OutterLink, Id FROM links WHERE InnerLink = $1;", url)
	var linkID int
	err = row.Scan(&res, &linkID)
	if err != nil {
		logger.Log.Debug("error on get link from DB")
		return
	}
	row = s.DB.QueryRow("SELECT is_deleted FROM users_links WHERE LinkId = $1 AND Id = $2;", linkID, userid)
	var del *bool
	err = row.Scan(&del)
	if del != nil {
		return res, ErrLinkDeleted
	}
	if err != nil {
		logger.Log.Debug("error on chekker to deleted with a query", zap.String("SELECT", fmt.Sprintf("SELECT is_deleted FROM users_links WHERE LinkId = %v AND Id = %v;", linkID, userid)))
		return res, nil
	}
	return
}

// Функция получения внешнего URL по короткому URL из базы данных
func (s *Storage) getUserURLBD(id int) (result []struct {
	ShortURL string
	URL      string
}, err error) {
	rows, err := s.DB.Query("SELECT InnerLink, OutterLink FROM links WHERE Id IN (SELECT LinkId FROM users_links WHERE Id = $1);", id)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()
	for rows.Next() {
		var in string
		var out string
		err = rows.Scan(&in, &out)
		if err != nil {
			return nil, err
		}
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

func (s Storage) getUserURLByShortLink(userID int, shortLink string) int {
	if userID < 1 || shortLink == "" || len(shortLink) < 8 {
		return -1
	}
	row := s.DB.QueryRow("SELECT linkid FROM users_links WHERE linkid IN (SELECT Id FROM links WHERE InnerLink LIKE '%' || $1) AND id = $2;", shortLink, userID)
	var id int
	err := row.Scan(&id)
	if err != nil {
		logger.Log.Debug("error of getting data from DB")
		return -1
	}
	return id
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

func (s *Storage) deleteFromDB(chIn chan struct {
	userID int
	linkID int
}) {
	tx, err := s.DB.Begin()
	if err != nil {
		return
	}
	defer tx.Commit()
	for ch := range chIn {
		if ch.linkID > 0 {
			_, err = tx.Exec("UPDATE users_links SET is_deleted = TRUE WHERE Id = $1 AND LinkId = $2", ch.userID, ch.linkID)
			logger.Log.Debug("DB query ", zap.String("UPDATE", fmt.Sprintf("UPDATE users_links SET is_deleted = TRUE WHERE Id = %v AND LinkId = %v", ch.userID, ch.linkID)))
			if err != nil {
				// если ошибка, то откатываем изменения
				tx.Rollback()
			}
		}
	}
	tx.Commit()
}

func (s *Storage) GetDB() (result struct {
	links []struct {
		InnerLink  string
		OutterLink string
		IDLink     int
	}
	users []struct {
		id int
		dt string
	}
	usersLinks []struct {
		userid int
		linkid int
	}
}, err error) {
	var rows *sql.Rows
	rows, err = s.DB.Query("SELECT InnerLink, OutterLink, Id FROM links;")
	if err != nil {
		return result, err
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()
	for rows.Next() {
		var in string
		var out string
		var idlink int
		err = rows.Scan(&in, &out, &idlink)
		if err != nil {
			return result, err
		}
		result.links = append(result.links, struct {
			InnerLink  string
			OutterLink string
			IDLink     int
		}{InnerLink: in,
			OutterLink: out,
			IDLink:     idlink,
		})
	}
	rows, err = s.DB.Query("SELECT Id, creation FROM users;")
	if err != nil {
		return result, err
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()
	for rows.Next() {
		var id int
		var cr string
		err = rows.Scan(&id, &cr)
		if err != nil {
			return result, err
		}
		result.users = append(result.users, struct {
			id int
			dt string
		}{id: id,
			dt: cr,
		})
	}
	rows, err = s.DB.Query("SELECT Id, LinkID FROM users_links;")
	if err != nil {
		return result, err
	}
	defer func() {
		_ = rows.Close()
		_ = rows.Err()
	}()
	for rows.Next() {
		var id int
		var linkid int
		err = rows.Scan(&id, &linkid)
		if err != nil {
			return result, err
		}
		result.usersLinks = append(result.usersLinks, struct {
			userid int
			linkid int
		}{userid: id,
			linkid: linkid,
		})
	}
	return result, err
}
