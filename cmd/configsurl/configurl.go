package configsurl

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Config struct {
	NetAddressServerShortener NetAddressServer
	NetAddressServerExpand    NetAddressServer
}

type NetAddressServer struct {
	Host string
	Port int
}

// checkNetAddress - функция проверяющая на корректность указания пары host:port и в случае ошибки передающей значения по умолчанию
func checkNetAddress(s string, h string, p int) (host string, port int, e error) {
	host, port = h, p
	v := strings.Split(s, "://")
	if len(v) < 1 || len(v) > 2 {
		e = errors.New("incorrect net address")
		return
	}
	if len(v) == 2 {
		s = v[1]
	}
	a := strings.Split(s, ":")
	if len(a) < 1 || len(a) > 2 {
		e = errors.New("incorrect net address")
		return
	}
	host = a[0]
	if a[1] != "" {
		port, e = strconv.Atoi(a[1])
		if e != nil {
			return
		}
	}
	return
}

func (c *Config) SetConfig() {

}

// возвращаем адрес вида host:port
func (n *NetAddressServer) String() string {
	return fmt.Sprint(n.Host + ":" + strconv.Itoa(n.Port))
}

// устанавливаем значения host и port в переменные
func (n *NetAddressServer) Set(s string) (err error) {
	n.Host, n.Port, err = checkNetAddress(s, n.Host, n.Port)
	if err != nil {
		return err
	}
	return nil
}
