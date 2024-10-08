package storage

import (
	"errors"
	"math/rand"
)

type Storage struct {
	InnerLinks  map[string]string
	OutterLinks map[string]string
}

// randChar - генерирует случайную букву латинского алфавита большую или маленькую или цифру
func randChar() int {
	max := 122
	min := 48
	res := rand.Intn(max-min) + min
	if res > 57 && res < 65 || res > 90 && res < 97 {
		return randChar()
	}
	return res
}

// CreateShortCode - генерирует новую короткую ссылку и проеряет на совпадение в "базе данных" если такая
// строка уже есть то делает рекурсию на саму себя пока не найдет уникальную ссылку
func (s *Storage) CreateShortCode(adr string) string {
	shortURL := []byte("http://" + adr + "/")
	for i := 0; i < 8; i++ {
		shortURL = append(shortURL, byte(randChar()))
	}
	result := string(shortURL)
	_, ok := s.InnerLinks[result]
	if ok {
		return s.CreateShortCode(adr)
	}
	return result
}

// CreateShortURL - получает ссылку которую необходимо сократить и проверяет на наличие ее в "базе данных",
// если  есть, то возвращает уже готовый короткий URL, если нет то запрашивает новую случайную коротную ссылку
func (s *Storage) CreateShortURL(url string, adr string) string {
	val, ok := s.OutterLinks[url]
	if ok {
		return val
	}
	result := s.CreateShortCode(adr)
	s.OutterLinks[url] = result
	s.InnerLinks[result] = url
	return result
}

// GetURL - получает коротную ссылку и проверяет наличие ее в "базе данных" если существует, то возвращяет ее
// если нет, то возвращает ошибку
func (s *Storage) GetURL(url string) (l string, e error) {
	l, ok := s.InnerLinks[url]
	if ok {
		return l, nil
	}
	return "", errors.New("link not found")
}
