package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type CarsList map[string]string

var cars = CarsList{
	"id1": "Renault Logan",
	"id2": "Renault Duster",
	"id3": "BMW X6",
	"id4": "BMW M5",
	"id5": "VW Passat",
	"id6": "VW Jetta",
	"id7": "Audi A4",
	"id8": "Audi Q7",
}

// carsListFunc — вспомогательная функция для вывода всех машин.
func carsListFunc(list CarsList) []string {
	var result []string
	for _, c := range list {
		result = append(result, c)
	}
	return result
}

// carFunc — вспомогательная функция для вывода определённой машины.
func carFunc(id string) string {
	if c, ok := cars[id]; ok {
		return c
	}
	return "unknown identifier " + id
}

// srhBrandFunc - функция для поиска всех автомобилей одного бренда
func srhBrandFunc(br string) CarsList {
	var result = make(map[string]string)
	for i, e := range cars {
		if br == strings.Split(e, " ")[0] {
			result[i] = e
		}
	}
	return result
}

// srhModelFunc - функция для поиска всех автомобилей одного бренда
func srhModelFunc(mdl string, list CarsList) string {
	var result string
	for i, e := range list {
		if mdl == strings.Split(e, " ")[1] {
			result = i
		}
	}
	return result
}

func carsHandle(rw http.ResponseWriter, r *http.Request) {
	carsList := carsListFunc(cars)
	io.WriteString(rw, strings.Join(carsList, ", "))
}

func carHandle(rw http.ResponseWriter, r *http.Request) {
	carID := chi.URLParam(r, "id")
	if carID == "" {
		http.Error(rw, "carID param is missed", http.StatusBadRequest)
		return
	}
	rw.Write([]byte(carFunc(carID)))
}

func brandHandle(rw http.ResponseWriter, r *http.Request) {
	brand := chi.URLParam(r, "brand")
	extCars := srhBrandFunc(brand)
	if len(extCars) == 0 {
		http.Error(rw, "Brand param is missed", http.StatusBadRequest)
		return
	}
	carsList := carsListFunc(extCars)
	io.WriteString(rw, strings.Join(carsList, ", "))
}

func modelHandle(rw http.ResponseWriter, r *http.Request) {
	brand := chi.URLParam(r, "brand")
	model := chi.URLParam(r, "model")
	extCars := srhBrandFunc(brand)
	if len(extCars) == 0 {
		http.Error(rw, "Brand params is missed", http.StatusBadRequest)
		return
	}
	carID := srhModelFunc(model, extCars)
	if carID == "" {
		http.Error(rw, "Model params is missed", http.StatusBadRequest)
		return
	}
	rw.Header().Add("Content-Type", "text/html")
	rw.Write([]byte(carFunc(carID)))
}

type NetAddress struct {
	Host string
	Port int
}

func (n *NetAddress) String() string {
	return fmt.Sprint(n.Host)
}

func (n *NetAddress) Set(flagValue string) error {
	var e error
	v := strings.Split(flagValue, "://")
	if len(v) != 2 {
		return fmt.Errorf("%s", "incorrect net address.")
	}
	a := strings.Split(v[1], ":")
	if len(a) < 1 || len(a) > 2 {
		return fmt.Errorf("%s", "incorrect net address.")
	}
	n.Host = a[0]
	n.Port, e = strconv.Atoi(a[1])
	if e != nil {
		return e
	}
	return nil
}

var version = "0.0.1"

// допишите код реализации методов интерфейса
// ...

func main() {
	addr := new(NetAddress)
	// если интерфейс не реализован,
	// здесь будет ошибка компиляции
	_ = flag.Value(addr)
	// проверка реализации
	flag.Var(addr, "addr", "Net address host:port")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Version: %s\nUsage of %s:\n", version, os.Args[0])
		flag.PrintDefaults()

	}
	flag.Parse()
	fmt.Println(addr.Host)
	fmt.Println(addr.Port)

	/*r := chi.NewRouter()

	r.Route("/car", func(r chi.Router) {
		r.Get("/{id}", carHandle) // POST /car
	})
	// определяем хендлер, который выводит все машины
	r.Route("/cars", func(r chi.Router) {
		r.Get("/", carsHandle) // POST /car
		r.Route("/{brand}", func(r chi.Router) {
			r.Get("/", brandHandle)        // GET /car/1234
			r.Get("/{model}", modelHandle) // PUT /car/1234
		})
	})

	log.Fatal(http.ListenAndServe(":8080", r))*/
}
