package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"math/rand"
	"os"

	"github.com/h1067675/shortUrl/internal/logger"
	"go.uber.org/zap"
)

type Storager interface {
	CreateShortURL(url string, adr string, useris int) (string, error)
	GetURL(url string) (l string, e error)
	SaveToFile(file string)
	PingDB() bool
	GetNewUserID() (int, error)
	GetUserURLS(int) ([]struct {
		ShortURL string
		URL      string
	}, error)
}

// Структура для хранения ссылок
type Storage struct {
	InnerLinks  map[string]string
	OutterLinks map[string]string
	Users       map[int][]string
	UsersLinks  map[string][]int
	DB          *SQLDB
}

// Функция создает новое хранилище
func NewStorage(database string) *Storage {
	var r = Storage{
		InnerLinks:  map[string]string{},
		OutterLinks: map[string]string{},
		Users:       map[int][]string{},
		UsersLinks:  map[string][]int{},
		DB:          newDB(database),
	}
	r.DB.Exec("DROP TABLE links;")
	r.DB.Exec("DROP TABLE users;")
	r.DB.Exec("DROP TABLE users_links;")
	if !r.checkLinksDBTable() {
		if !r.createLinksDBTable() {
			r.DB.Connected = false
		}
	}
	if !r.checkUsersDBTable() {
		if !r.createUsersDBTables() {
			r.DB.Connected = false
		}
	}
	return &r
}

// структура описывает формат json для хранения данных в файле
type StorageJSON struct {
	ShortLink    string `json:"short_url"`
	OriginalLink string `json:"original_url"`
	UserID       []int  `json:"user_id"`
}

var ErrLinkExsist = errors.New("link already exsist")

// Функция генерирует случайный символ из набора a-z,A-Z,0-9 и возвращает его байтовое представление
func randChar() int {
	min, max := 48, 122
	res := rand.Intn(max-min) + min
	if res > 57 && res < 65 || res > 90 && res < 97 {
		return randChar()
	}
	return res
}

// Функция генерирует новую короткую ссылку и проверяет на совпадение в "базе данных" если такая
// строка уже есть то делает рекурсию на саму себя пока не найдет уникальную ссылку
func (s *Storage) createShortCode(adr string) string {
	shortURL := []byte("http://" + adr + "/")
	for i := 0; i < 8; i++ {
		shortURL = append(shortURL, byte(randChar()))
	}
	result := string(shortURL)
	_, ok := s.InnerLinks[result]
	if ok {
		return s.createShortCode(adr)
	}
	return result
}

// Функция получает ссылку которую необходимо сократить и проверяет на наличие ее в "базе данных",
// если  есть, то возвращает уже готовый короткий URL, если нет то запрашивает новую случайную коротную ссылку
func (s *Storage) CreateShortURL(url string, adr string, userid int) (result string, err error) {
	logger.Log.Debug("DB connection", zap.Bool("is", s.DB.Connected))
	if s.DB.Connected {
		result, err = s.saveShortURLBD(url, adr, userid)
	} else {
		result, ok := s.OutterLinks[url]
		if ok {
			userids, ok := s.UsersLinks[result]
			if ok {
				for _, e := range userids {
					if e == userid {
						return result, ErrLinkExsist
					}
				}
				s.UsersLinks[result] = append(s.UsersLinks[result], userid)
			}
			s.Users[userid] = append(s.Users[userid], result)
			s.UsersLinks[result] = append(s.UsersLinks[result], userid)
			return result, err
		}
		result = s.createShortCode(adr)
		s.OutterLinks[url] = result
		s.InnerLinks[result] = url
		s.Users[userid] = append(s.Users[userid], result)
		s.UsersLinks[result] = append(s.UsersLinks[result], userid)
		return result, err
	}
	return
}

func (s *Storage) GetNewUserID() (result int, err error) {
	if s.DB.Connected {
		result, err = s.getNewUserIDDB()
		if err != nil {
			return -1, err
		}
	} else {
		result = 1
		for i, _ := range s.Users {
			if i > result {
				result = i
			}
		}
	}
	return result, nil
}

// Функция получает коротную ссылку и проверяет наличие ее в "базе данных" если существует, то возвращяет ее
// если нет, то возвращает ошибку
func (s *Storage) GetURL(url string) (l string, e error) {
	if s.DB.Connected {
		l = s.getURLBD(url)
		if l != "" {
			return l, nil
		}
	} else {
		l, ok := s.InnerLinks[url]
		if ok {
			return l, nil
		}
	}
	return "", errors.New("link not found")
}

func (s *Storage) GetUserURLS(id int) (result []struct {
	ShortURL string
	URL      string
}, err error) {
	if s.DB.Connected {
		result, _ = s.getUserURLBD(id)
	} else {
		for _, e := range s.Users[id] {
			result = append(result, struct {
				ShortURL string
				URL      string
			}{
				ShortURL: e,
				URL:      s.OutterLinks[e],
			})
		}
	}
	if len(result) > 0 {
		return result, nil
	}
	return nil, errors.New("links not found")
}

// Функция сохранения хранилища в файл
func (s *Storage) SaveToFile(file string) {
	st := []StorageJSON{}
	for i, e := range s.InnerLinks {
		st = append(st, StorageJSON{i, e, s.UsersLinks[i]})
	}
	tf, err := json.Marshal(st)
	if err != nil {
		panic(err)
	}
	fl, err := os.Create(file)
	if err != nil {
		panic(err)
	}
	defer fl.Close()
	fl.Write(tf)
	logger.Log.Debug("Saved to ", zap.String("file", file))
}

// Функция восстановления ссылок из файла
func (s *Storage) RestoreFromfile(file string) {
	fl, err := os.Open(file)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return
		}
		panic(err)
	}
	defer fl.Close()
	st := []StorageJSON{}
	r := bufio.NewScanner(fl)
	r.Scan()
	bt := r.Bytes()
	if len(bt) > 0 {
		if err := json.Unmarshal(bt, &st); err != nil {
			panic(err)
		}
		for _, e := range st {
			s.OutterLinks[e.OriginalLink] = e.ShortLink
			s.InnerLinks[e.ShortLink] = e.OriginalLink
			for _, k := range e.UserID {
				s.Users[k] = append(s.Users[k], e.ShortLink)
			}
			s.UsersLinks[e.ShortLink] = append(s.UsersLinks[e.ShortLink], e.UserID...)
		}
	}

}
