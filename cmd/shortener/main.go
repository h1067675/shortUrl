package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
)

const form = `<html>
    <head>
    <title></title>
    </head>
    <body>
        <form action="/" method="post" enctype="text/plain">
            <label>Ссылка <input type="text" name="url"></label>
            <button type="submit">yandex.ru</button>
        </form>
    </body>
</html>`

func randChar() int {
	max := 122
	min := 48
	res := rand.Intn(max-min) + min
	if res > 57 && res < 65 || res > 90 && res < 97 {
		return randChar()
	}
	return res
}

func createUrl(url string) string {
	sortUrl := []byte("http://localhost:8080/")
	val, ok := outUrls[url]
	if ok {
		return val
	}
	for i := 0; i < 8; i++ {
		sortUrl = append(sortUrl, byte(randChar()))
	}
	result := string(sortUrl)
	_, ok = shortUrls[result]
	if ok {
		return createUrl(url)
	}
	outUrls[url] = result
	shortUrls[result] = url
	return result
}

func getUrl(url string) (bool, string) {
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
				if vv == val {
					return true
				}
			}
		}
	}
	return false
}

func basePath(responce http.ResponseWriter, request *http.Request) {
	url := request.URL.Path

	fmt.Println(request.Method)
	fmt.Println(url)
	for k, v := range request.Header {
		for _, vv := range v {
			fmt.Printf("%s : %s \n", k, vv)
		}
	}

	if request.Method == http.MethodPost && url == "/" && checkHeader(request.Header, "Content-Type", "text/plain") {
		responce.Header().Add("Content-Type", "text/plain")
		responce.WriteHeader(http.StatusCreated)
		request.ParseForm()
		url, err := io.ReadAll(request.Body)
		if err != nil {
			panic(err)
		}
		fmt.Println(url)
		body := createUrl(string(url))
		responce.Write([]byte(body))
		return
	} else if request.Method == http.MethodGet {
		ok, outUrl := getUrl("http://localhost:8080" + url)
		if ok {
			responce.Header().Add("Location", outUrl)
			responce.WriteHeader(http.StatusTemporaryRedirect)
			return
		}
	}
	responce.WriteHeader(http.StatusBadRequest)
	//responce.Write([]byte(form))
}

var outUrls = make(map[string]string)
var shortUrls = make(map[string]string)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", basePath)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
