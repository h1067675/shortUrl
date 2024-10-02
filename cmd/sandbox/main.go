package main

import (
	"io"
	"log"
	"net/http"
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

func main() {
	r := chi.NewRouter()

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

	log.Fatal(http.ListenAndServe(":8080", r))
}
