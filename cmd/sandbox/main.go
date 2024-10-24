package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type Slice []byte

// MarshalJSON реализует интерфейс json.Marshaler.
func (s Slice) MarshalJSON() ([]byte, error) {
	type SliceAlias Slice
	aliaseValue := &struct {
		SliceAlias
		Value json.Marshaler
	}{
		SliceAlias: (SliceAlias)(s),
	}
	va := hex.EncodeToString(aliaseValue.SliceAlias)
	return json.Marshal(va)
}

// UnmarshalJSON реализует интерфейс json.Unmarshaler.
func (s *MySlice) UnmarshalJSON(data []byte) error {
	type SliceAlias MySlice
	aliasValue := &struct {
		*SliceAlias
		StringHex string `json:"Slice"`
	}{
		SliceAlias: (*SliceAlias)(s),
	}
	// var tmp string
	if err := json.Unmarshal(data, aliasValue); err != nil {
		return err
	}
	sl, err := hex.DecodeString(aliasValue.StringHex)
	if err != nil {
		panic(err)
	}
	aliasValue.SliceAlias.Slice = sl
	return nil
}

type MySlice struct {
	ID    int
	Slice Slice
}

func main() {
	ret, err := json.Marshal(MySlice{ID: 7, Slice: []byte{1, 2, 3, 10, 11, 255}})
	if err != nil {
		panic(err)
	}
	fmt.Println(string(ret))
	var result MySlice
	if err = json.Unmarshal(ret, &result); err != nil {
		panic(err)
	}
	fmt.Println(result)
}

// package main

// import (
// 	"encoding/json"
// 	"fmt"
// )

// func main() {
// 	var v interface{}
// 	err := json.Unmarshal([]byte(`[0, 10, 30]`), &v)
// 	fmt.Printf("%T, %[1]v, %v", v, err)
// }

// import (
// 	"encoding/json"
// 	"fmt"
// 	"log"
// )

// type Data struct {
// 	ID      int    `json:"-"`
// 	Name    string `json:"name,omitempty"`
// 	Company string `json:"comp,omitempty"`
// }

// func main() {
// 	foo := []Data{
// 		{
// 			ID:   10,
// 			Name: "Gopher",
// 		},
// 		{
// 			Name:    "Вася",
// 			Company: "Яндекс",
// 		},
// 	}
// 	out, err := json.Marshal(foo)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Println(string(out))
// }

// import (
// 	"reflect"
// 	"strconv"
// 	"strings"
// )

// // User используется для тестирования.
// type User struct {
// 	Nick string
// 	Age  int `limit:"18"`
// 	Rate int `limit:"0,100"`
// }

// // Str2Int конвертирует строку в int.
// func Str2Int(s string) int {
// 	v, err := strconv.Atoi(s)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return v
// }

// // Validate проверяет min и max для int c тегом limit.
// func Validate(obj interface{}) bool {
// 	vobj := reflect.ValueOf(obj)
// 	objType := vobj.Type() // получаем описание типа

// 	// перебираем все поля структуры
// 	for i := 0; i < objType.NumField(); i++ {
// 		// берём значение текущего поля и проверяем, что это int
// 		if v, ok := vobj.Field(i).Interface().(int); ok {
// 			// подсказка: тег limit надо искать в поле objType.Field(i)
// 			// objType.Field(i).Tag.Lookup или objType.Field(i).Tag.Get
// 			if lim, ok := objType.Field(i).Tag.Lookup("limit"); ok {
// 				lims := strings.Split(lim, ",")
// 				if v < Str2Int(lims[0]) {
// 					return false
// 				}
// 				if len(lims) > 1 && v > Str2Int(lims[1]) {
// 					return false
// 				}
// 			}
// 		}
// 	}
// 	return true
// }

// import (
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"os"
// 	"strconv"
// 	"strings"

// 	"github.com/go-chi/chi"
// 	log "github.com/sirupsen/logrus"
// )

// type CarsList map[string]string

// var cars = CarsList{
// 	"id1": "Renault Logan",
// 	"id2": "Renault Duster",
// 	"id3": "BMW X6",
// 	"id4": "BMW M5",
// 	"id5": "VW Passat",
// 	"id6": "VW Jetta",
// 	"id7": "Audi A4",
// 	"id8": "Audi Q7",
// }

// // carsListFunc — вспомогательная функция для вывода всех машин.
// func carsListFunc(list CarsList) []string {
// 	var result []string
// 	for _, c := range list {
// 		result = append(result, c)
// 	}
// 	return result
// }

// // carFunc — вспомогательная функция для вывода определённой машины.
// func carFunc(id string) string {
// 	if c, ok := cars[id]; ok {
// 		return c
// 	}
// 	return "unknown identifier " + id
// }

// // srhBrandFunc - функция для поиска всех автомобилей одного бренда
// func srhBrandFunc(br string) CarsList {
// 	var result = make(map[string]string)
// 	for i, e := range cars {
// 		if br == strings.Split(e, " ")[0] {
// 			result[i] = e
// 		}
// 	}
// 	return result
// }

// // srhModelFunc - функция для поиска всех автомобилей одного бренда
// func srhModelFunc(mdl string, list CarsList) string {
// 	var result string
// 	for i, e := range list {
// 		if mdl == strings.Split(e, " ")[1] {
// 			result = i
// 		}
// 	}
// 	return result
// }

// func carsHandle(rw http.ResponseWriter, r *http.Request) {
// 	carsList := carsListFunc(cars)
// 	io.WriteString(rw, strings.Join(carsList, ", "))
// }

// func carHandle(rw http.ResponseWriter, r *http.Request) {
// 	carID := chi.URLParam(r, "id")
// 	if carID == "" {
// 		http.Error(rw, "carID param is missed", http.StatusBadRequest)
// 		return
// 	}
// 	rw.Write([]byte(carFunc(carID)))
// }

// func brandHandle(rw http.ResponseWriter, r *http.Request) {
// 	brand := chi.URLParam(r, "brand")
// 	extCars := srhBrandFunc(brand)
// 	if len(extCars) == 0 {
// 		http.Error(rw, "Brand param is missed", http.StatusBadRequest)
// 		return
// 	}
// 	carsList := carsListFunc(extCars)
// 	io.WriteString(rw, strings.Join(carsList, ", "))
// }

// func modelHandle(rw http.ResponseWriter, r *http.Request) {
// 	brand := chi.URLParam(r, "brand")
// 	model := chi.URLParam(r, "model")
// 	extCars := srhBrandFunc(brand)
// 	if len(extCars) == 0 {
// 		http.Error(rw, "Brand params is missed", http.StatusBadRequest)
// 		return
// 	}
// 	carID := srhModelFunc(model, extCars)
// 	if carID == "" {
// 		http.Error(rw, "Model params is missed", http.StatusBadRequest)
// 		return
// 	}
// 	rw.Header().Add("Content-Type", "text/html")
// 	rw.Write([]byte(carFunc(carID)))
// }

// type NetAddress struct {
// 	Host string
// 	Port int
// }

// func (n *NetAddress) String() string {
// 	return fmt.Sprint(n.Host)
// }

// func (n *NetAddress) Set(flagValue string) error {
// 	var e error
// 	v := strings.Split(flagValue, "://")
// 	if len(v) != 2 {
// 		return fmt.Errorf("%s", "incorrect net address.")
// 	}
// 	a := strings.Split(v[1], ":")
// 	if len(a) < 1 || len(a) > 2 {
// 		return fmt.Errorf("%s", "incorrect net address.")
// 	}
// 	n.Host = a[0]
// 	n.Port, e = strconv.Atoi(a[1])
// 	if e != nil {
// 		return e
// 	}
// 	return nil
// }

// var version = "0.0.1"

// type User struct {
// 	Name string `env:"USERNAME"`
// }

// // допишите код реализации методов интерфейса
// // ...

// func main() {
// 	// создаём файл info.log и обрабатываем ошибку
// 	file, err := os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// откладываем закрытие файла
// 	defer file.Close()

// 	// устанавливаем вывод логов в файл
// 	log.SetOutput(file)
// 	// устанавливаем вывод логов в формате JSON
// 	log.SetFormatter(&log.JSONFormatter{})
// 	// устанавливаем уровень предупреждений
// 	log.SetLevel(log.WarnLevel)

// 	// определяем стандартные поля JSON
// 	log.WithFields(log.Fields{
// 		"genre": "metal",
// 		"name":  "Rammstein",
// 	}).Info("Немецкая метал-группа, образованная в январе 1994 года в Берлине.")

// 	log.WithFields(log.Fields{
// 		"omg":  true,
// 		"name": "Garbage",
// 	}).Warn("В 2021 году вышел новый альбом No Gods No Masters.")

// 	log.WithFields(log.Fields{
// 		"omg":  true,
// 		"name": "Linkin Park",
// 	}).Fatal("Группа Linkin Park взяла паузу после смерти вокалиста Честера Беннингтона 20 июля 2017 года.")

// 	// var buf bytes.Buffer
// 	// var nr = io.Writer(&buf)
// 	// var mylog = log.New(nr, "mylog: ", 0)
// 	// mylog.Print("Hello, world!")
// 	// mylog.Print("Goodbye")

// 	// fmt.Print(&buf)
// 	// var user User
// 	// err := env.Parse(&user)
// 	// if err != nil {
// 	// 	log.Fatal(err)
// 	// }
// 	// log.Println(user.Name)

// 	// u := os.Getenv("USERNAME")
// 	// fmt.Println(u)
// 	// envList := os.Environ()
// 	// // выводим первые пять элементов
// 	// for i := 0; i < 5 && i < len(envList); i++ {
// 	// 	fmt.Println(envList[i])
// 	// }
// 	// addr := new(NetAddress)
// 	// // если интерфейс не реализован,
// 	// // здесь будет ошибка компиляции
// 	// _ = flag.Value(addr)
// 	// // проверка реализации
// 	// flag.Var(addr, "addr", "Net address host:port")

// 	// flag.Usage = func() {
// 	// 	fmt.Fprintf(flag.CommandLine.Output(), "Version: %s\nUsage of %s:\n", version, os.Args[0])
// 	// 	flag.PrintDefaults()

// 	// }
// 	// flag.Parse()
// 	// fmt.Println(addr.Host)
// 	// fmt.Println(addr.Port)

// 	/*r := chi.NewRouter()

// 	r.Route("/car", func(r chi.Router) {
// 		r.Get("/{id}", carHandle) // POST /car
// 	})
// 	// определяем хендлер, который выводит все машины
// 	r.Route("/cars", func(r chi.Router) {
// 		r.Get("/", carsHandle) // POST /car
// 		r.Route("/{brand}", func(r chi.Router) {
// 			r.Get("/", brandHandle)        // GET /car/1234
// 			r.Get("/{model}", modelHandle) // PUT /car/1234
// 		})
// 	})

// 	log.Fatal(http.ListenAndServe(":8080", r))*/
// }
