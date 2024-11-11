package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"math/rand"
	"os"
)

// Структура для хранения ссылок
type Storage struct {
	InnerLinks  map[string]string
	OutterLinks map[string]string
	DB          *SQLDB
}

// Функция создает новое хранилище
func NewStorage(database string) *Storage {
	var r = Storage{
		InnerLinks:  map[string]string{},
		OutterLinks: map[string]string{},
		DB:          newDB(database),
	}
	if !r.checkDBTable() {
		if !r.createDBTable() {
			r.DB.Connected = false
		}
	}
	return &r
}

// структура описывает формат json для хранения данных в файле
type StorageJSON struct {
	ShortLink    string `json:"short_url"`
	OriginalLink string `json:"original_url"`
}

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
func (s *Storage) CreateShortURL(url string, adr string) string {
	var result string
	if s.DB.Connected {
		result = s.saveShortURLBD(url, adr)
	} else {
		val, ok := s.OutterLinks[url]
		if ok {
			return val
		}
		result := s.createShortCode(adr)
		s.OutterLinks[url] = result
		s.InnerLinks[result] = url
	}
	return result
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

// Функция сохранения хранилища в файл
func (s *Storage) SaveToFile(file string) {
	st := []StorageJSON{}
	for i, e := range s.InnerLinks {
		st = append(st, StorageJSON{i, e})
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
	if err := json.Unmarshal(bt, &st); err != nil {
		panic(err)
	}
	for _, e := range st {
		s.OutterLinks[e.OriginalLink] = e.ShortLink
		s.InnerLinks[e.ShortLink] = e.OriginalLink
	}
}
