package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
)

var Config struct {
	NetAddressServerShortener
	NetAddressServerExpand
}

type NetAddressServerShortener struct {
	Host string
	Port int
}

type NetAddressServerExpand struct {
	Host string
	Port int
}

// checkNetAddress - функция проверяющая на корректность указания пары host:port и в случае ошибки передающей значения по умолчанию
func checkNetAddress(s string) (host string, port int, e error) {
	var a []string
	v := strings.Split(s, "://")
	if len(v) > 1 {
		if len(v) > 2 || len(v) < 1 {
			return "localhost", 8080, fmt.Errorf("%s", "incorrect net address.")
		}
		a = strings.Split(v[1], ":")
	} else {
		a = strings.Split(s, ":")
	}
	if len(a) < 1 || len(a) > 2 {
		return "localhost", 8080, fmt.Errorf("%s", "incorrect net address.")
	}
	host = a[0]
	if a[1] != "" {
		port, e = strconv.Atoi(a[1])
		if e != nil {
			return "localhost", 8080, e
		}
	} else {
		port = 80
	}
	return
}

func (n *NetAddressServerShortener) String() string {
	return fmt.Sprint(n.Host + ":" + strconv.Itoa(n.Port))
}

func (n *NetAddressServerShortener) Set(flagValue string) (err error) {
	n.Host, n.Port, err = checkNetAddress(flagValue)
	if err != nil {
		return err
	}
	return nil
}

func (n *NetAddressServerExpand) String() string {
	return fmt.Sprint(n.Host + ":" + strconv.Itoa(n.Port))
}

func (n *NetAddressServerExpand) Set(flagValue string) (err error) {
	n.Host, n.Port, err = checkNetAddress(flagValue)
	if err != nil {
		return err
	}
	return nil
}

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
	shortURL := []byte("http://" + addrShortener.String() + "/")
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
		ok, outURL := getURL("http://" + addrShortener.String() + url)
		if ok {
			responce.Header().Add("Location", outURL)
			responce.WriteHeader(http.StatusTemporaryRedirect)
			return
		}
	}
	responce.WriteHeader(http.StatusBadRequest)
}

func router() chi.Router {
	r := chi.NewRouter()
	// Делаем маршрутизацию
	r.Route("/", func(r chi.Router) {
		r.Post("/", shorten) // POST запрос отправляем на сокращение ссылки
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", expand) // GET запрос с id направляем на извлечение ссылки
		})
	})
	return r
}

var outUrls = make(map[string]string)
var shortUrls = make(map[string]string)
var addrShortener = new(NetAddressServerExpand)
var addrExpand = new(NetAddressServerShortener)

func parseFlags() {
	addrShortener.Host, addrShortener.Port = "localhost", 8080
	addrExpand.Host, addrExpand.Port = "localhost", 8080

	flag.Var(addrShortener, "a", "Net address shortener service (host:port)")
	flag.Var(addrExpand, "b", "Net address expand service (host:port)")
	flag.Parse()
}

func main() {
	parseFlags()

	log.Fatal(http.ListenAndServe(addrShortener.String(), router()))
}
