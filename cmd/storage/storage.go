package storage

import (
	"errors"
	"math/rand"
)

type Storage struct {
	InnerLinks  map[string]string
	OutterLinks map[string]string
}

func NewStorage() *Storage {
	var r = Storage{
		InnerLinks:  map[string]string{},
		OutterLinks: map[string]string{},
	}
	return &r
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
	val, ok := s.OutterLinks[url]
	if ok {
		return val
	}
	result := s.createShortCode(adr)
	s.OutterLinks[url] = result
	s.InnerLinks[result] = url
	return result
}

// Функция получает коротную ссылку и проверяет наличие ее в "базе данных" если существует, то возвращяет ее
// если нет, то возвращает ошибку
func (s *Storage) GetURL(url string) (l string, e error) {
	l, ok := s.InnerLinks[url]
	if ok {
		return l, nil
	}
	return "", errors.New("link not found")
}
