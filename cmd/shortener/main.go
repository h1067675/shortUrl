package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
)

func randChar() int {
	max := 122
	min := 48
	res := rand.Intn(max-min) + min
	if res > 57 && res < 65 || res > 90 && res < 97 {
		return randChar()
	}
	return res
}

func createURL(url string) string {
	shortURL := []byte("http://localhost:8080/")
	val, ok := outUrls[url]
	if ok {
		return val
	}
	for i := 0; i < 8; i++ {
		shortURL = append(shortURL, byte(randChar()))
	}
	result := string(shortURL)
	_, ok = shortUrls[result]
	if ok {
		return createURL(url)
	}
	outUrls[url] = result
	shortUrls[result] = url
	return result
}

func getURL(url string) (bool, string) {
	val, ok := shortUrls[url]
	if ok {
		return true, val
	}
	return false, ""
}

func checkHeader(hd http.Header, key string, val string) bool {
	for k, v := range hd {
		if key == k {
			for _, vv := range v {
				if strings.Contains(vv, val) {
					return true
				}
			}
		}
	}
	return false
}

func shorten(responce http.ResponseWriter, request *http.Request) {
	// логгирование работы функции
	fmt.Println("func: shorten")
	fmt.Println("Requesr method: ", request.Method)
	fmt.Println("Requesr URL: ", request.URL.Path)
	for k, v := range request.Header {
		for _, vv := range v {
			fmt.Printf("%s : %s \n", k, vv)
		}
	}
	// логгирование окончено

	// проверяем на content-type
	if checkHeader(request.Header, "Content-Type", "text/plain") {
		var body string
		// если прошли то присваиваем значение content-type: "text/plain" и статус 201
		responce.Header().Add("Content-Type", "text/plain")
		responce.WriteHeader(http.StatusCreated)
		// получаем тело запроса
		url, err := io.ReadAll(request.Body)
		if err != nil {
			panic(err)
		}
		// если тело запроса не пустое, то создаем сокращенный url и выводим в тело ответа
		if len(url) > 0 {
			body = createURL(string(url))
			responce.Write([]byte(body))
		}
		return
	}
	responce.WriteHeader(http.StatusBadRequest)
}

func expand(responce http.ResponseWriter, request *http.Request) {
	url := request.URL.Path

	fmt.Println("func: expand")
	fmt.Println("Requesr method: ", request.Method)
	fmt.Println("Requesr URL: ", url)

	if request.Method == http.MethodGet {
		ok, outURL := getURL("http://localhost:8080" + url)
		if ok {
			responce.Header().Add("Location", outURL)
			responce.WriteHeader(http.StatusTemporaryRedirect)
			return
		}
	}
	responce.WriteHeader(http.StatusBadRequest)
}

var outUrls = make(map[string]string)
var shortUrls = make(map[string]string)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", shorten)
	mux.HandleFunc("GET /", expand)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
