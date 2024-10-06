package configSUrl

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var Config struct {
	NetAddressServerShortener NetAddressServer
	NetAddressServerExpand    NetAddressServer
}

type NetAddressServer struct {
	Host string
	Port int
}

// checkNetAddress - функция проверяющая на корректность указания пары host:port и в случае ошибки передающей значения по умолчанию
func checkNetAddress(s string) (host string, port int, e error) {
	v := strings.Split(s, "://")
	if len(v) != 2 {
		e = errors.New("Incorrect net address.")
		return "localhost", 8080, e
	}
	a := strings.Split(v[1], ":")
	if len(a) < 1 || len(a) > 2 {
		e = errors.New("Incorrect net address.")
		return "localhost", 8080, e
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

// возвращаем адрес вида host:port
func (n *NetAddressServer) String() string {
	return fmt.Sprint(n.Host + ":" + strconv.Itoa(n.Port))
}

// устанавливаем значения в состояние аргументов
func (n *NetAddressServer) Set(flagValue string) (err error) {
	n.Host, n.Port, err = checkNetAddress(flagValue)
	if err != nil {
		return err
	}
	return nil
}
