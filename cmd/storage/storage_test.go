package storage

import (
	"fmt"
	"testing"

	"github.com/h1067675/shortUrl/internal/logger"
	"go.uber.org/zap"
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
	err := logger.Initialize("info")
	if err != nil {
		fmt.Print(err)
	}
	s := NewStorage("")
	for i := 0; i < b.N; i++ {
		url := createRandURL()
		adr := s.createShortCode(url)
		b.StartTimer()
		_, err := s.CreateShortURL(url, adr, 0)
		if err != nil {
			logger.Log.Info("Errors create short URL", zap.String("Error", err.Error()))
		}
	}
}

func BenchmarkGetNewUserID(b *testing.B) {
	s := NewStorage("")
	for i := 0; i < b.N; i++ {
		_, err := s.GetNewUserID()
		if err != nil {
			logger.Log.Info("Error get new user ID", zap.String("Error", err.Error()))
		}
	}
}

func BenchmarkGetURL(b *testing.B) {
	b.StopTimer()
	s := NewStorage("")
	for i := 0; i < b.N; i++ {
		url := createRandURL()
		b.StartTimer()
		_, err := s.GetURL(url, 0)
		if err != nil {
			logger.Log.Info("Error getting URL", zap.String("Error", err.Error()))
		}
	}
}

func BenchmarkSaveToFile(b *testing.B) {
	b.StopTimer()
	err := logger.Initialize("info")
	if err != nil {
		fmt.Print(err)
	}
	s := NewStorage("")
	for j := 0; j < 1000; j++ {
		url := createRandURL()
		adr := s.createShortCode(url)
		_, err := s.CreateShortURL(url, adr, 0)
		if err != nil {
			logger.Log.Info("Errors create short URL", zap.String("Error", err.Error()))
		}
	}
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		err := s.SaveToFile("./banckmark_tmp.json")
		if err != nil {
			logger.Log.Info("Errors safe file", zap.String("Error", err.Error()))
		}
	}
}
