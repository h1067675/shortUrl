package storage

import (
	"testing"

	"github.com/h1067675/shortUrl/internal/logger"
)

func createRandURL() string {
	var shortURL []byte
	for i := 0; i < 8; i++ {
		shortURL = append(shortURL, byte(randChar()))
	}
	result := string(shortURL) + ".ru"
	return result
}
func BenchmarkCreateShortCode(b *testing.B) {
	s := NewStorage("")
	for i := 0; i < b.N; i++ {
		s.createShortCode("ya.ru")
	}
}

func BenchmarkCreateShortURL(b *testing.B) {
	b.StopTimer()
	logger.Initialize("info")
	s := NewStorage("")
	for i := 0; i < b.N; i++ {
		url := createRandURL()
		adr := s.createShortCode(url)
		b.StartTimer()
		s.CreateShortURL(url, adr, 0)
	}
}

func BenchmarkGetNewUserID(b *testing.B) {
	s := NewStorage("")
	for i := 0; i < b.N; i++ {
		s.GetNewUserID()
	}
}

func BenchmarkGetURL(b *testing.B) {
	b.StopTimer()
	s := NewStorage("")
	for i := 0; i < b.N; i++ {
		url := createRandURL()
		b.StartTimer()
		s.GetURL(url, 0)
	}
}

func BenchmarkSaveToFile(b *testing.B) {
	b.StopTimer()
	logger.Initialize("info")
	s := NewStorage("")
	for j := 0; j < 1000; j++ {
		url := createRandURL()
		adr := s.createShortCode(url)
		s.CreateShortURL(url, adr, 0)
	}
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		s.SaveToFile("./banckmark_tmp.json")
	}
}
